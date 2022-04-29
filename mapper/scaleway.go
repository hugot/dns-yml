package mapper

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	scwdomain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"snorba.art/hugo/dns-yml/document"
)

const (
	env_scw_access_key = "DNS_YML_SCW_ACCESS_KEY"
	env_scw_secret_key = "DNS_YML_SCW_SECRET_KEY"
	env_scw_org_id     = "DNS_YML_SCW_ORG_ID"
)

func NewScalewayMapper(env func(string) string) (*ScalewayMapper, error) {
	accessKey := env(env_scw_access_key)
	secretKey := env(env_scw_secret_key)
	orgID := env(env_scw_org_id)

	if accessKey == "" || secretKey == "" {
		return nil, errors.New(
			fmt.Sprintf(
				`At least one config parameter is not set. All of the following environment variables must be set:
                 - %s: Scaleway access key
                 - %s: Scaleway secret key,
				- %s: Scaleway organizaition ID`,
				env_scw_access_key,
				env_scw_secret_key,
				env_scw_org_id,
			),
		)
	}

	client, err := scw.NewClient(
		scw.WithAuth(accessKey, secretKey),
		scw.WithDefaultOrganizationID(orgID),
		scw.WithDefaultZone(scw.ZoneNlAms1),
	)
	if err != nil {
		return nil, err
	}

	return &ScalewayMapper{
		client: client,
	}, nil
}

type ScalewayMapper struct {
	client *scw.Client
}

func (m *ScalewayMapper) MapYaml(directory string, ymlReader io.Reader) error {
	return mapYaml(m, directory, ymlReader)
}

func (m *ScalewayMapper) Map(directory string, root *document.Root) error {
	dnsApi := scwdomain.NewAPI(m.client)

	for domainName, domain := range root.Domains {
		existingZone, err := m.getOrCreateExistingZone(domainName, dnsApi)
		if err != nil {
			return err
		}

		log.Printf("Retrieving existing records for zone %s", existingZone.Domain)
		existingRecords, err := dnsApi.ListDNSZoneRecords(&scwdomain.ListDNSZoneRecordsRequest{
			DNSZone: existingZone.Domain,
		})
		if err != nil {
			return err
		}

		desiredRecords := make(UniqueRecordMap)
		creations := &scwdomain.RecordChangeAdd{}

		for _, record := range domain.Records {
			ttl := record.TTL
			if ttl == 0 {
				ttl = DefaultTTL
			}

			values, err := record.Content.ResolveValue(directory, root.RoundRobins)
			if err != nil {
				return err
			}

			for _, value := range values {
				log.Printf(
					"Ensuring presence of record \"%s\" \"%s\" \"%s\"",
					record.Type,
					record.Name,
					value,
				)
				desired := &scwdomain.Record{
					Name: strings.TrimSuffix(
						strings.TrimSuffix(record.Name, domainName),
						".",
					),
					Type: scwdomain.RecordType(record.Type),
					Data: value,
					TTL:  uint32(ttl),
				}

				if record.Type == "MX" || record.Type == "SRV" {
					desired.Priority = uint32(record.Priority)
				}

				// Make domain names absolute. If we don't, scaleway will append
				// the domain of the zone to the end of them.
				if record.Type == "CNAME" ||
					record.Type == "DNAME" ||
					record.Type == "ALIAS" ||
					record.Type == "NS" ||
					record.Type == "MX" ||
					record.Type == "PTR" {
					desired.Data = desired.Data + "."
				}

				desiredRecords.AddRecord(desired)
				creations.Records = append(creations.Records, desired)
			}
		}

		var deletions []*scwdomain.RecordChangeDelete
		for _, rec := range existingRecords.Records {
			if !desiredRecords.HasRecord(rec.Name, rec.Type.String(), rec.Data) {
				deletions = append(deletions, &scwdomain.RecordChangeDelete{
					ID: &rec.ID,
				})
			}
		}

		changeReq := &scwdomain.UpdateDNSZoneRecordsRequest{
			DNSZone: existingZone.Domain,
		}

		for _, deletion := range deletions {
			changeReq.Changes = append(
				changeReq.Changes,
				&scwdomain.RecordChange{
					Delete: deletion,
				},
			)
		}

		changeReq.Changes = append(changeReq.Changes, &scwdomain.RecordChange{
			Add: creations,
		})

		_, err = dnsApi.UpdateDNSZoneRecords(changeReq)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *ScalewayMapper) createChangeSet(
	existing *scwdomain.Record,
	record document.Record,
	ttl int,
	desiredRecords UniqueRecordMap,
	directory string,
	root *document.Root,
) (*scwdomain.RecordChangeSet, error) {
	var changeRecords []*scwdomain.Record
	changeSet := &scwdomain.RecordChangeSet{
		ID: &existing.ID,
	}

	existing.TTL = uint32(ttl)
	values, err := record.Content.ResolveValue(directory, root.RoundRobins)
	if err != nil {
		return nil, err
	}

	if len(values) == 1 {
		existing.Data = values[0]
		desiredRecords.AddRecord(existing)
		changeRecords = append(changeRecords, existing)
	} else if len(values) < 1 {
		return nil, fmt.Errorf(
			"DNS record without value detected: %s, %s, %s",
			record.Name,
			record.Type,
			record.Content.Type,
		)
	} else {
		for _, value := range values {
			newRec := &scwdomain.Record{
				Name: record.Name,
				Type: scwdomain.RecordType(record.Type),
				TTL:  uint32(ttl),
				Data: value,
			}

			desiredRecords.AddRecord(newRec)
			changeRecords = append(changeRecords, newRec)
		}
	}

	for _, rec := range changeRecords {
		if record.Type == "MX" || record.Type == "SRV" {
			rec.Priority = uint32(record.Priority)
		}
	}

	changeSet.Records = changeRecords

	return changeSet, nil
}

func (m *ScalewayMapper) getOrCreateExistingZone(
	domainName string,
	api *scwdomain.API,
) (*scwdomain.DNSZone, error) {
	log.Printf("Retrieving existing DNS Zones for %s", domainName)
	existingZones, err := api.ListDNSZones(&scwdomain.ListDNSZonesRequest{
		Domain:  domainName,
		DNSZone: domainName,
	})
	if err != nil {
		return nil, err
	}

	if existingZones.TotalCount == 0 {
		log.Printf("No zones found for %s, requesting creation", domainName)
		return api.CreateDNSZone(&scwdomain.CreateDNSZoneRequest{
			Domain:    domainName,
			Subdomain: domainName,
		})
	} else if existingZones.TotalCount > 1 {
		return nil, errors.New(
			"More than one zone detected for configured domain. This is not supported.",
		)
	}

	return existingZones.DNSZones[0], nil
}

type RecordMap map[string]*scwdomain.Record

func (m RecordMap) AddRecord(rec *scwdomain.Record) {
	m[rec.Name+":"+rec.Type.String()] = rec
}

func (m RecordMap) HasRecord(name string, rType string) bool {
	_, ok := m[name+":"+rType]

	return ok
}

func (m RecordMap) GetRecord(name string, rType string) *scwdomain.Record {
	return m[name+":"+rType]
}

type UniqueRecordMap map[string]*scwdomain.Record

func (m UniqueRecordMap) AddRecord(rec *scwdomain.Record) {
	m[rec.Name+":"+rec.Type.String()+":"+rec.Data] = rec
}

func (m UniqueRecordMap) HasRecord(name string, rType string, data string) bool {
	_, ok := m[name+":"+rType+":"+data]

	return ok
}

func (m UniqueRecordMap) GetRecord(name string, rType string, data string) *scwdomain.Record {
	return m[name+":"+rType+":"+data]
}

package mapper

import (
	"fmt"
	"io"
	"log"
	"strings"

	"snorba.art/hugo/dns-yml/document"
)

func NewDryMapper(env func(string) string) (*DryMapper, error) {
	return &DryMapper{}, nil
}

type DryMapper struct {
}

func (m *DryMapper) MapYaml(directory string, ymlReader io.Reader) error {
	return mapYaml(m, directory, ymlReader)
}

func (m *DryMapper) Map(directory string, root *document.Root) error {
	for domainName, domain := range root.Domains {
		for _, record := range domain.Records {
			if !rTypeValid(record.Type) {
				return fmt.Errorf(
					"Invalid record type \"%s\" used for record with name \"%s\"",
					record.Type,
					record.Name,
				)
			}

			if !strings.HasSuffix(record.Name, domainName) {
				return fmt.Errorf(
					"Record name %s is not a subdomain of root %s",
					record.Name,
					domainName,
				)
			}

			if record.Type == rtype_mx {
				if record.Priority == 0 {
					log.Printf(
						"Warning: MX record with name %s does not have a priority configured",
						record.Name,
					)
				}
			}

			values, err := record.Content.ResolveValue(directory, root.RoundRobins)
			if err != nil {
				return err
			}

			for _, value := range values {
				if value == "" {
					return fmt.Errorf(
						"%s record for %s is empty",
						record.Type,
						record.Name,
					)
				}
			}
		}
	}

	return nil
}

package mapper

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
	"snorba.art/hugo/dns-yml/database"
	"snorba.art/hugo/dns-yml/document"

	_ "github.com/go-sql-driver/mysql"
)

const DefaultTTL = 86400

func NewMapper(config *Config) (*Mapper, error) {
	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s",
		config.Database.Username,
		config.Database.Password,
		config.Database.Address,
		config.Database.Name,
	)

	DB, err := sql.Open("mysql", connectionString)

	if err != nil {
		return nil, err
	}

	err = DB.Ping()

	return &Mapper{DB: DB}, err
}

type Mapper struct {
	DB *sql.DB
}

// Note: directory is used to prefix relative filepaths in "file" record types.
func (m *Mapper) MapYaml(directory string, ymlReader io.Reader) error {
	ymlData, err := ioutil.ReadAll(ymlReader)

	docRoot := &document.Root{}

	err = yaml.Unmarshal(ymlData, docRoot)

	if err != nil {
		return err
	}

	return m.Map(directory, docRoot)
}

func (m *Mapper) getOrCreateSavedDomain(domain string) (*database.Domain, error) {
	domainStmt, err := m.DB.Prepare("SELECT id, name, type FROM domains WHERE name = ?")
	if err != nil {
		return nil, err
	}
	defer domainStmt.Close()

	savedDomain := &database.Domain{}
	err = domainStmt.QueryRow(domain).Scan(
		&savedDomain.ID, &savedDomain.Name, &savedDomain.Type,
	)

	if err == sql.ErrNoRows {
		savedDomain = &database.Domain{
			Name: domain,
			Type: "MASTER",
		}

		insertStmt, err := m.DB.Prepare("INSERT INTO domains (name, type) VALUES (?, ?)")
		if err != nil {
			return nil, err
		}

		_, err = insertStmt.Exec(savedDomain.Name, savedDomain.Type)

		if err != nil {
			return nil, err
		}

		insertStmt.Close()

		return m.getOrCreateSavedDomain(domain)
	}

	return savedDomain, err
}

// Note: directory is used to prefix relative filepaths in "file" record types.
func (m *Mapper) Map(directory string, root *document.Root) error {
	for domainName, domain := range root.Domains {
		if domain.SOARecord.Hostmaster == "" || domain.SOARecord.Primary == "" {
			return errors.New(
				"either one of required SOA fields hostmaster or primary is not set" +
					" see see https://doc.powerdns.com/authoritative/appendices/types.html#soa" +
					" for documentation about these fields",
			)
		}

		savedDomain, err := m.getOrCreateSavedDomain(domainName)

		recordsToCreate := make([]database.Record, 0)

		// a SOA record for domain
		recordsToCreate = append(recordsToCreate, database.Record{
			DomainID: savedDomain.ID,
			Name:     domainName,
			Type:     "SOA",
			Content:  domain.SOARecord.ToContent(),
			TTL:      DefaultTTL,
		})

		for _, record := range domain.Records {
			ttl := record.TTL
			if ttl == 0 {
				ttl = DefaultTTL
			}

			recordToCreate := &database.Record{
				DomainID: savedDomain.ID,
				Type:     record.Type,
				Name:     record.Name,
				Priority: record.Priority,
				TTL:      ttl,
			}

			switch record.Content.Type {
			case "raw":
				recordToCreate.Content = record.Content.Value
				recordsToCreate = append(recordsToCreate, *recordToCreate)
			case "file":
				filePath := record.Content.Value
				if filePath[0] != '/' {
					filePath = directory + "/" + filePath
				}

				file, err := os.Open(filePath)
				if err != nil {
					return err
				}

				fileContents, err := ioutil.ReadAll(file)
				if err != nil {
					return err
				}

				recordToCreate.Content = string(fileContents)
				recordsToCreate = append(recordsToCreate, *recordToCreate)
			case "round-robin":
				if contentValues, ok := root.RoundRobins[record.Content.Value]; ok {
					for _, value := range contentValues {
						recordsToCreate = append(recordsToCreate, database.Record{
							DomainID: savedDomain.ID,
							Type:     record.Type,
							Name:     record.Name,
							Content:  value,
							Priority: record.Priority,
							TTL:      ttl,
						})
					}
				} else {
					return errors.New(
						fmt.Sprintf(
							"No round_robin of name \"%s\" was found",
							record.Content.Value,
						),
					)
				}
			}
		}

		err = m.applyRecords(savedDomain, recordsToCreate)

		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Mapper) applyRecords(domain *database.Domain, records []database.Record) error {
	existingRecStmt, err := m.DB.Prepare("SELECT id FROM records WHERE domain_id = ?")
	if err != nil {
		return err
	}
	defer existingRecStmt.Close()

	rows, err := existingRecStmt.Query(domain.ID)
	if err != nil {
		return err
	}

	existingRecordIDs := make([]int, 0)
	for rows.Next() {
		var ID int

		err := rows.Scan(&ID)
		if err != nil {
			return err
		}

		existingRecordIDs = append(existingRecordIDs, ID)
	}

	insertRecStmt, err := m.DB.Prepare(
		"INSERT INTO records (domain_id, type, name, content, prio, ttl) VALUES (?,?,?,?,?,?)",
	)
	if err != nil {
		return err
	}
	defer insertRecStmt.Close()

	for _, record := range records {
		_, err := insertRecStmt.Exec(
			record.DomainID,
			record.Type,
			record.Name,
			record.Content,
			record.Priority,
			record.TTL,
		)

		if err != nil {
			return err
		}
	}

	deleteExistingRecStmt, err := m.DB.Prepare("DELETE FROM records WHERE id = ?")
	if err != nil {
		return err
	}
	defer deleteExistingRecStmt.Close()

	for _, ID := range existingRecordIDs {
		_, err := deleteExistingRecStmt.Exec(ID)
		if err != nil {
			return err
		}
	}

	return err
}

func (m *Mapper) Close() error {
	return m.DB.Close()
}

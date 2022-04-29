package document

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type Root struct {
	Domains     map[string]Domain     `yaml:"domains"`
	RoundRobins map[string]RoundRobin `yaml:"round_robins"`
}

type Domain struct {
	Records   []Record  `yaml:"records"`
	SOARecord SOARecord `yaml:"soa"`
}

type Record struct {
	Type     string        `yaml:"type"`
	Name     string        `yaml:"name"`
	Content  RecordContent `yaml:"content"`
	Priority int           `yaml:"priority"`
	TTL      int           `yaml:"ttl"`
}

type RecordContent struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

func (c RecordContent) ResolveValue(
	directory string,
	rrs map[string]RoundRobin,
) ([]string, error) {
	switch c.Type {
	case "raw":
		return []string{c.Value}, nil
	case "file":
		filePath := c.Value
		if filePath[0] != '/' {
			filePath = directory + "/" + filePath
		}

		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}

		contents, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}

		return []string{strings.TrimSpace(string(contents))}, err
	case "round-robin":
		var values []string
		if contentValues, ok := rrs[c.Value]; ok {
			for _, value := range contentValues {
				values = append(values, value)
			}

			return values, nil
		}

		return nil, errors.New(
			fmt.Sprintf(
				"No round robin by name \"%s\" was found",
				c.Value,
			),
		)
	default:
		return nil, fmt.Errorf("Invalid record value type \"%s\"", c.Type)
	}
}

type RoundRobin []string

type SOARecord struct {
	Primary    string `yaml:"primary"`
	Hostmaster string `yaml:"hostmaster"`
	Refresh    int    `yaml:"refresh"`
	Retry      int    `yaml:"retry"`
	Expire     int    `yaml:"expire"`
	DefaultTTL int    `yaml:"default_ttl"`
}

// Fill in 0 integer values with defaults and return SOA formatted record contents
func (r SOARecord) ToContent() string {
	if r.Refresh == 0 {
		r.Refresh = 10800
	}

	if r.Retry == 0 {
		r.Retry = 3600
	}

	if r.Expire == 0 {
		r.Expire = 604800
	}

	if r.DefaultTTL == 0 {
		r.DefaultTTL = 3600
	}

	return fmt.Sprintf(
		"%s %s %d %d %d %d %d",
		r.Primary,
		strings.Replace(r.Hostmaster, "@", ".", len(r.Hostmaster)),
		time.Now().Unix(),
		r.Refresh,
		r.Retry,
		r.Expire,
		r.DefaultTTL,
	)
}

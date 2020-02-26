package document

import (
	"fmt"
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

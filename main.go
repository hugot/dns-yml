package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"snorba.art/hugo/dns-yml/mapper"
)

const mapperPDNS = "pdns"
const mapperScaleway = "scaleway"

var mappers = []string{mapperPDNS, mapperScaleway}

func sliceContains(slice []string, thing string) bool {
	for _, i := range slice {
		if thing == i {
			return true
		}
	}

	return false
}

func main() {
	argLen := len(os.Args)

	if argLen != 3 {
		fmt.Fprintln(os.Stderr, "Usage: dns-yml CONFIG DNS_DEFINITION")
		return
	}

	cmd := flag.NewFlagSet("dns-yml", flag.ExitOnError)
	mapperFlag := cmd.String(
		"mapper",
		"scaleway",
		"Mapper to use. Available mappers: "+strings.Join(mappers, ", "),
	)
	err := cmd.Parse(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	var dnsMapper mapper.Mapper
	if *mapperFlag == mapperPDNS {
		dnsMapper, err = mapper.NewPDNSMapper(os.Getenv)
	}

	if *mapperFlag == mapperScaleway {
		log.Fatal("not implemented")
	}

	definitionPath := cmd.Arg(0)
	if definitionPath == "" {
		cmd.Usage()
		os.Exit(1)
	}

	definitionReader, err := os.Open(definitionPath)

	err = dnsMapper.MapYaml(filepath.Dir(definitionPath), definitionReader)
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"flag"
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
	cmd := flag.NewFlagSet("dns-yml", flag.ExitOnError)
	mapperFlag := cmd.String(
		"mapper",
		"scaleway",
		"Mapper to use. Available mappers: "+strings.Join(mappers, ", "),
	)
	help := cmd.Bool("help", false, "Show this help")

	err := cmd.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if *help {
		cmd.Usage()
		os.Exit(0)
	}

	definitionPath := cmd.Arg(0)
	if definitionPath == "" {
		cmd.Usage()
		os.Exit(1)
	}

	definitionReader, err := os.Open(definitionPath)

	var dnsMapper mapper.Mapper
	if *mapperFlag == mapperPDNS {
		dnsMapper, err = mapper.NewPDNSMapper(os.Getenv)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *mapperFlag == mapperScaleway {
		dnsMapper, err = mapper.NewScalewayMapper(os.Getenv)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = dnsMapper.MapYaml(filepath.Dir(definitionPath), definitionReader)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Successfully finished execution of %s mapper", *mapperFlag)
}

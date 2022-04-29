package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"snorba.art/hugo/dns-yml/mapper"
	"snorba.art/hugo/dns-yml/util"
)

const mapperPDNS = "pdns"
const mapperScaleway = "scaleway"
const mapperDry = "dry"

var mappers = []string{mapperPDNS, mapperScaleway, mapperDry}

func main() {
	cmd := flag.NewFlagSet("dns-yml", flag.ExitOnError)
	mapperFlag := cmd.String(
		"mapper",
		"scaleway",
		"Mapper to use. Use the \"dry\" mapper to check the config without persisting it. Available mappers: "+strings.Join(mappers, ", "),
	)
	help := cmd.Bool("help", false, "Show this help")

	err := cmd.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if !util.SliceContainsString(mappers, *mapperFlag) {
		log.Printf("Invalid mapper parameter %s", *mapperFlag)
		cmd.Usage()
		os.Exit(1)
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
	} else if *mapperFlag == mapperScaleway {
		dnsMapper, err = mapper.NewScalewayMapper(os.Getenv)
		if err != nil {
			log.Fatal(err)
		}
	} else if *mapperFlag == mapperDry {
		dnsMapper, err = mapper.NewDryMapper(os.Getenv)
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

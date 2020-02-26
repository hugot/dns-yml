package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"snorba.art/hugo/dns-yml/mapper"
)

func main() {
	argLen := len(os.Args)

	if argLen != 3 {
		fmt.Fprintln(os.Stderr, "Usage: dns-yml CONFIG DNS_DEFINITION")
		return
	}

	configPath := os.Args[1]

	configReader, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}

	config, err := mapper.ConfigFromYaml(configReader)
	if err != nil {
		log.Fatal(err)
	}

	mapper, err := mapper.NewMapper(config)
	if err != nil {
		log.Fatal(err)
	}

	definitionPath := os.Args[2]
	definitionReader, err := os.Open(definitionPath)

	err = mapper.MapYaml(filepath.Dir(definitionPath), definitionReader)
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

var botConfig botConfiguration

type botConfiguration struct {
	Webhooks     []string            `toml:"webhooks"`
	Endpoint     string              `toml:"endpoint"`
	Repositories map[string][]string `toml:"repositories"`
}

func readConfigOrPanic() {
	if _, err := os.Stat("config.toml"); err != nil {
		log.Fatalln("no config.toml; use the sample one")
	}

	content, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Fatalln("config.toml couldn't be read:", err)
	}

	err = toml.Unmarshal(content, &botConfig)
	if err != nil {
		log.Fatalln("couldn't unmarshal config:", err)
	}
}

// vim: set ff=unix autoindent ts=4 sw=4 tw=0 noet :

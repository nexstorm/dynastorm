package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

var (
	cfg = make(map[string]string)
)

func parseconfig(cfgpath string) map[string]string {
	// read config from config.yaml
	configf, err := os.ReadFile(cfgpath)
	if err != nil {
		panic(err)
	}
	yaml.Unmarshal(configf, &cfg)
	return cfg

}

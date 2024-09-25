package main

import (
	"log"
	"os"

	commonConfig "github.com/meirongdev/movie-microservice/pkg/config"
	"gopkg.in/yaml.v3"
)

type config struct {
	API apiConfig `yaml:"api"`
}

type apiConfig struct {
	Port        int                      `yaml:"port"`
	MysqlConfig commonConfig.MySQLConfig `yaml:"mysql"`
}

func loadConfig(path string) (config, error) {
	log.Println("Loading config from", path)
	var cfg config
	// check path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, err
	}
	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

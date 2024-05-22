package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type Config struct {
	Http     Http     `yaml:"http"`
	Database Database `yaml:"database"`
}

type Http struct {
	Port string `yaml:"port"`
}

type Database struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func MustLoad(path string) *Config {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatal("config file does not exists: " + path)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		log.Fatalf("somwting went wrong %v", err)
	}

	return &cfg

}

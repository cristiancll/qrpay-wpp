package configs

import (
	"encoding/json"
	"os"
)

type Server struct {
	Name     string `json:"name"`
	Port     int    `json:"port"`
	KeysPath string `json:"keys_path"`
}

type Database struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type Config struct {
	Server   Server   `json:"server"`
	Database Database `json:"db"`
}

var instance *Config

func Load(settingsPath string) error {
	if instance != nil {
		return nil
	}
	file, err := os.Open(settingsPath)
	if err != nil {
		return err
	}
	defer file.Close()
	configs := &Config{}
	err = json.NewDecoder(file).Decode(configs)
	if err != nil {
		return err
	}

	instance = configs
	return nil
}

func Get() *Config {
	return instance
}

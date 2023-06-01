package main

import (
	"flag"
	server "qrpay-wpp/internal/api"
)

func main() {
	var settingsPath string
	flag.StringVar(&settingsPath, "settings", "./configs/settings.json", "Path to settings file")
	flag.Parse()
	s := server.New(settingsPath)
	err := s.Start()
	if err != nil {
		panic(err)
	}
}

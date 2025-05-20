package main

import (
	"log"

	"github.com/timmbarton/layout/configloader"
	"github.com/timmbarton/layout/executor"

	"backend/internal/app"
	"backend/internal/config"
)

func main() {
	cfg := config.Config{}

	err := configloader.Load(&cfg)
	if err != nil {
		log.Println(err)
		return
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Println(err)
		return
	}

	err = executor.Run(a)
	if err != nil {
		log.Println(err)
		return
	}
}

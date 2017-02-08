package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/rancher/catalog-service/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

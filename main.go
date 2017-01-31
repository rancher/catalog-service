package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/rancher/catalog-service/manager"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/catalog-service/service"
)

var (
	refreshInterval = flag.Int("refresh-interval", 60, "Time interval (in seconds) to periodically pull the catalog from git repo")
	port            = flag.Int("port", 8088, "HTTP listen port")
	cacheRoot       = flag.String("cache-root", "./cache", "Cache root")
	configFile      = flag.String("configFile", "./repo.json", "Config file")
	refresh         = flag.Bool("refresh", false, "Refresh and exit")
)

func main() {
	flag.Parse()

	config, err := readConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: add a flag for this
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.AutoMigrate(&model.CatalogModel{})
	db.AutoMigrate(&model.TemplateModel{})
	db.AutoMigrate(&model.VersionModel{})
	db.AutoMigrate(&model.FileModel{})

	go func() {
		m := manager.NewManager(*cacheRoot, config, db)
		if err = m.CreateConfigCatalogs(); err != nil {
			log.Fatal(err)
		}
		if err = m.RefreshAll(); err != nil {
			log.Fatal(err)
		}
		if *refresh {
			os.Exit(0)
		}
	}()

	if *refresh {
		select {}
	}

	log.Infof("Starting Catalog Service on port %d", *port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), &service.MuxWrapper{
		IsReady: false,
		Router:  service.NewRouter(manager.NewManager(*cacheRoot, config, db), db),
	}))
}

func readConfig(configFile string) (map[string]manager.CatalogConfig, error) {
	configContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config map[string]map[string]manager.CatalogConfig
	if err = json.Unmarshal(configContents, &config); err != nil {
		return nil, err
	}
	return config["catalogs"], nil
}

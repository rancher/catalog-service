package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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
	//catalogURL = flag.String("catalog-url", "", "Catalog URL of the form name=URL")
	configFile = flag.String("config", "./repo.json", "Config file")
)

func main() {
	flag.Parse()

	bytes, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	var config map[string]manager.CatalogConfig
	if err = json.Unmarshal(bytes, &config); err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.AutoMigrate(&model.CatalogModel{})
	db.AutoMigrate(&model.TemplateModel{})
	db.AutoMigrate(&model.VersionModel{})

	m := manager.NewManager(*cacheRoot, config, db)
	if err = m.RefreshAll(); err != nil {
		// TODO
		//log.Fatal(err)
		fmt.Println(err)
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 8088), &service.MuxWrapper{
		IsReady: false,
		Router:  service.NewRouter(manager.NewManager(*cacheRoot, config, db), db),
	}))
}

func readConfig(configFile string) (map[string]manager.CatalogConfig, error) {
	configContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config map[string]manager.CatalogConfig
	err = json.Unmarshal(configContents, &config)
	return config, err
}

package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/rancher/catalog-service/manager"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/catalog-service/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	refreshInterval int
	port            int
	cacheRoot       string
	configFile      string
	validateOnly    bool
	sqlite          bool
	migrateDb       bool
	debug           bool
)

var RootCmd = &cobra.Command{
	Use: "catalog-service",
	Run: run,
}

func init() {
	viper.SetEnvPrefix("catalog_service")
	viper.AutomaticEnv()

	RootCmd.PersistentFlags().IntVar(&refreshInterval, "refresh-interval", 60, "")
	RootCmd.PersistentFlags().IntVarP(&port, "port", "p", 8088, "")
	RootCmd.PersistentFlags().StringVar(&cacheRoot, "cache", "./cache", "")
	RootCmd.PersistentFlags().StringVar(&configFile, "config", "./repo.json", "")
	RootCmd.PersistentFlags().BoolVar(&validateOnly, "validate", false, "")
	RootCmd.PersistentFlags().BoolVar(&sqlite, "sqlite", false, "")
	RootCmd.PersistentFlags().BoolVar(&migrateDb, "migrate-db", false, "")
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "")

	RootCmd.PersistentFlags().String("mysql-user", "", "")
	viper.BindPFlag("mysql_user", RootCmd.PersistentFlags().Lookup("mysql-user"))

	RootCmd.PersistentFlags().String("mysql-password", "", "")
	viper.BindPFlag("mysql_password", RootCmd.PersistentFlags().Lookup("mysql-password"))

	RootCmd.PersistentFlags().String("mysql-address", "", "")
	viper.BindPFlag("mysql_address", RootCmd.PersistentFlags().Lookup("mysql-address"))

	RootCmd.PersistentFlags().String("mysql-dbname", "", "")
	viper.BindPFlag("mysql_dbname", RootCmd.PersistentFlags().Lookup("mysql-dbname"))

	RootCmd.PersistentFlags().String("mysql-params", "", "")
	viper.BindPFlag("mysql_params", RootCmd.PersistentFlags().Lookup("mysql-params"))
}

func run(cmd *cobra.Command, args []string) {
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	var db *gorm.DB
	var err error
	if sqlite {
		db, err = gorm.Open("sqlite3", "local.db")
		if err != nil {
			log.Fatal(err)
		}
		db.Exec("PRAGMA foreign_keys = ON")
		migrateDb = true
	} else {
		user := viper.GetString("mysql_user")
		password := viper.GetString("mysql_password")
		address := viper.GetString("mysql_address")
		dbname := viper.GetString("mysql_dbname")
		params := viper.GetString("mysql_params")

		db, err = gorm.Open("mysql", formatDSN(user, password, address, dbname, params))
		if err != nil {
			log.Fatal(err)
		}
	}
	defer db.Close()

	db.SingularTable(true)
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		defaultTableName = strings.TrimSuffix(defaultTableName, "_model")
		if defaultTableName == "catalog" {
			return defaultTableName
		}
		return "catalog_" + defaultTableName
	}

	if migrateDb {
		log.Info("Migrating DB")
		db.AutoMigrate(&model.CatalogModel{})

		db.AutoMigrate(&model.TemplateModel{})
		db.AutoMigrate(&model.CategoryModel{})
		db.AutoMigrate(&model.TemplateCategoryModel{})
		db.AutoMigrate(&model.LabelModel{})

		db.AutoMigrate(&model.VersionModel{})
		db.AutoMigrate(&model.FileModel{})
	}

	m := manager.NewManager(cacheRoot, configFile, db)
	go refresh(m, refreshInterval, validateOnly)
	if validateOnly {
		select {}
	}

	log.Infof("Starting Catalog Service (port %d, refresh interval %d seconds)", port, refreshInterval)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), &service.MuxWrapper{
		IsReady: false,
		Router:  service.NewRouter(m, db),
	}))
}

func formatDSN(user, password, address, dbname, params string) string {
	paramsMap := map[string]string{
		"parseTime": "true",
	}
	for _, param := range strings.Split(params, "&") {
		split := strings.SplitN(param, "=", 2)
		if len(split) > 1 {
			paramsMap[split[0]] = split[1]
		}
	}
	mysqlConfig := &mysql.Config{
		User:   user,
		Passwd: password,
		Addr:   address,
		DBName: dbname,
		Params: paramsMap,
	}
	return mysqlConfig.FormatDSN()
}

func refresh(m *manager.Manager, refreshInterval int, validateOnly bool) {
	if err := m.RefreshAll(); err != nil {
		log.Fatalf("Failed to do initial refresh of catalogs: %v", err)
	}
	if validateOnly {
		os.Exit(0)
	}
	// TODO: don't want to have refresh running twice at the same time
	for range time.Tick(time.Duration(refreshInterval) * time.Second) {
		log.Debugf("Performing automatic refresh of all catalogs (interval %d seconds)", refreshInterval)
		go m.RefreshAll()
	}
}

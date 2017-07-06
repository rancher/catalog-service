package tracking

import (
	"os"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"
	rancher "github.com/rancher/go-rancher/v2"
)

const (
	uuidSetting = "install.uuid"
	uuidPattern = "[[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12}"
)

var logger = log.WithFields(log.Fields{"service": "catalog"})

func LoadRancherUUID() string {
	uuid := ""

	client, err := rancher.NewRancherClient(&rancher.ClientOpts{
		Url:       os.Getenv("CATALOG_SERVICE_CATTLE_URL"),
		AccessKey: os.Getenv("CATALOG_SERVICE_CATTLE_ACCESS_KEY"),
		SecretKey: os.Getenv("CATALOG_SERVICE_CATTLE_SECRET_KEY"),
		Timeout:   5 * time.Second,
	})

	if err != nil {
		logger.WithField("error", err.Error()).Fatal("Failed to create client")
	}

	var setting *rancher.Setting
	if setting, err = client.Setting.ById(uuidSetting); err != nil {
		logger.WithFields(log.Fields{
			"setting": "install.uuid",
			"error":   err.Error(),
		}).Warn("Failed to read setting")

	} else if setting.Value == "" {
		logger.WithField("setting", "install.uuid").Warn("Setting is empty")

	} else if matched := regexp.MustCompile(uuidPattern).MatchString(setting.Value); matched {
		uuid = setting.Value

	} else {
		logger.WithField("uuid", setting.Value).Warn("Malformed")
	}

	return uuid
}

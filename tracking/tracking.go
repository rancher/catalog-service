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
	nilUUID     = "00000000-0000-0000-0000-000000000000"
)

func LoadRancherUUID() string {
	client, err := rancher.NewRancherClient(&rancher.ClientOpts{
		Url:       os.Getenv("CATTLE_URL"),
		AccessKey: os.Getenv("CATTLE_ACCESS_KEY"),
		SecretKey: os.Getenv("CATTLE_SECRET_KEY"),
		Timeout:   5 * time.Second,
	})

	if err != nil {
		log.Warnf("Error creating client: %v", err)
		return nilUUID
	}

	var setting *rancher.Setting
	setting, err = client.Setting.ById(uuidSetting)
	if err != nil {
		log.Warnf("Error retrieving setting: %v", err)
		return nilUUID
	}

	matched := regexp.MustCompile(uuidPattern).MatchString(setting.Value)
	if !matched {
		log.Warnf("Malformed UUID: %s", setting.Value)
		return nilUUID
	}

	return setting.Value
}

package configuration

import (
	"io/ioutil"
	"sync"

	"gitlab.com/i4s-edu/petstore-kovalyk/services/auth"

	"gitlab.com/i4s-edu/petstore-kovalyk/services/storage"
	"gitlab.com/i4s-edu/petstore-kovalyk/utils"

	"gitlab.com/i4s-edu/petstore-kovalyk/workers"

	"gitlab.com/i4s-edu/petstore-kovalyk/db"

	"github.com/sirupsen/logrus"

	"github.com/BurntSushi/toml"
)

const configFile = "config.toml"

type Server struct {
	Host            string
	Port            string
	ShutdownTimeout utils.Duration
}

type Config struct {
	Server  Server
	DB      db.Config
	Workers workers.Config
	Storage storage.Config
	Auth    auth.Config
}

var config Config
var once sync.Once

func Load() {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		logrus.Fatalf("Problem with %v, %v", configFile, err)
	}
	if _, err := toml.Decode(string(data), &config); err != nil {
		logrus.Fatalf("Problem with configration file, %v %v", configFile, err)
	}
}
func GetConfig() Config {
	once.Do(Load)
	return config
}

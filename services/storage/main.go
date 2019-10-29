package storage

import (
	"github.com/sirupsen/logrus"
)

var storage Storage

type Config struct {
	Type  string
	Minio MinioConfig
}

type Storage interface {
	GetLink(destination, filename string) (string, error)
	Save(destination, filename, contentType, filePath string) error
	Update(destination, filename string) error
	Delete(destination, filename string) error
}

func Init(config Config) {
	switch config.Type {
	case "minio":
		InitMinio(config.Minio)
	case "filesystem":
		//TODO (OR NOT)
	default:
		logrus.Fatalf("unsupported storage type")
	}
}
func GetStorage() Storage {
	if storage == nil {
		logrus.Fatalf("storage has not initialized")
	}
	return storage
}

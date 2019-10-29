package storage

import (
	"fmt"
	"log"

	"github.com/sirupsen/logrus"

	"github.com/minio/minio-go/v6"
)

const InvoicesBucket = "invoices"

type MinioConfig struct {
	Host      string
	Port      string
	AccessKey string
	SecretKey string
	SSL       bool
}
type MinioStorage struct {
	minioClient *minio.Client
}

func InitMinio(config MinioConfig) {
	logrus.Info("init MinioStorage")
	endpoint := fmt.Sprintf("%v:%v", config.Host, config.Port)
	minioClient, err := minio.New(endpoint, config.AccessKey, config.SecretKey, config.SSL)
	if err != nil {
		log.Fatal("minio create error", err)
	}
	bucketName := InvoicesBucket
	location := "us-east-1"
	err = minioClient.MakeBucket(bucketName, location)
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(bucketName)
		if errBucketExists == nil && exists {
			logrus.Printf("Bucket exists: %s", bucketName)
		} else {
			logrus.Fatal("problem with bucket creating: ", err)
		}
	}
	logrus.Info("bucket successfully created ")

	logrus.Printf("Successfully created %s\n", bucketName)
	storage = MinioStorage{minioClient: minioClient}

}

func (mc MinioStorage) Save(bucketName, filename, contentType, filePath string) error {
	_, err := mc.minioClient.FPutObject(
		bucketName,
		filename,
		filePath,
		minio.PutObjectOptions{ContentType: contentType})
	logrus.Infof("file saved to minio storage bucket=%s filename=%s", bucketName, filename)
	return err
}

func (mc MinioStorage) GetLink(bucket, filename string) (string, error) {
	return "", nil
}

func (mc MinioStorage) Update(bucket, filename string) error {
	return nil
}

func (mc MinioStorage) Delete(bucket, filename string) error {
	return nil
}

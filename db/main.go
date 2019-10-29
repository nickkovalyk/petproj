package db

import (
	"github.com/jmoiron/sqlx"

	"fmt"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Postgres PostgresConfig
}

type PostgresConfig struct {
	DBName   string
	Host     string
	Port     string
	User     string
	Password string
	SSLMode  string
}

func GetPostgresConnection(config PostgresConfig) (*sqlx.DB, error) {
	dbInfo := fmt.Sprintf("user=%s password=%s dbname=%s  host=%s port=%s sslmode=%s",
		config.User,
		config.Password,
		config.DBName,
		config.Host,
		config.Port,
		config.SSLMode)
	logrus.Info(dbInfo)
	db, err := sqlx.Connect("postgres", dbInfo)
	if err != nil {
		return nil, err
	}

	return db, nil
}

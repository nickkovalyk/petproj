package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"gitlab.com/i4s-edu/petstore-kovalyk/services/storage"
	"gitlab.com/i4s-edu/petstore-kovalyk/workers"

	"gitlab.com/i4s-edu/petstore-kovalyk/db/migrations"
	"gitlab.com/i4s-edu/petstore-kovalyk/services/auth"

	db2 "gitlab.com/i4s-edu/petstore-kovalyk/db"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"gitlab.com/i4s-edu/petstore-kovalyk/api/routing"
	"gitlab.com/i4s-edu/petstore-kovalyk/configuration"
)

type App struct {
	Config configuration.Config
	Server *http.Server
	DB     *sqlx.DB
}

func NewApp() *App {
	config := configuration.GetConfig()
	logrus.Info("Current config:", config)

	db, err := db2.GetPostgresConnection(config.DB.Postgres)
	if err != nil {
		logrus.Fatalf("Problem with database connection: %v", err)
	}

	err = migrations.Run(db)
	if err != nil {
		logrus.Fatalf("Problem with migrations: %v", err)
	}
	storage.Init(config.Storage)
	auth.Init(config.Auth)

	srv := http.Server{
		Addr:         fmt.Sprintf("%v:%v", config.Server.Host, config.Server.Port),
		Handler:      routing.NewRouter(db),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &App{Config: config, Server: &srv, DB: db}
}

func (a *App) Run() {
	go func() {
		err := a.Server.ListenAndServe()
		if err != nil {
			logrus.Fatalf("Cannot run app: %v", err)
		}
	}()
	workers.DispatchInvoiceWorker(a.Config.Workers.Invoice.Interval.Duration, a.DB)

	a.gracefulShutdown()
}

func (a *App) gracefulShutdown() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)
	sig := <-quit
	logrus.Info("Server is shutting down reason:" + sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), a.Config.Server.ShutdownTimeout.Duration)
	defer cancel()

	a.Server.SetKeepAlivesEnabled(false)
	if err := a.Server.Shutdown(ctx); err != nil {
		logrus.Fatalf("Could not gracefully shutdown the server: %v", err)
	}
	logrus.Info("Server stopped")
}

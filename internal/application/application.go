package application

import (
	"go.uber.org/zap"
	"service/internal/adapters/storage/postgres"
	"service/internal/cases"
	"service/internal/config"
	"service/internal/ports/http"
)

type Application struct {
	log    *zap.SugaredLogger
	cfg    *config.Config
	server *http.Server
}

func (a *Application) Build(configPath string) {
	var err error

	logger, _ := zap.NewDevelopment()

	a.log = logger.Sugar()

	a.cfg, err = config.NewConfig(configPath)
	if err != nil {
		logger.Fatal("init config")
	}

	storage := a.buildPostgresStorage()

	svc := a.buildService(storage)

	a.server = a.buildServer(svc)
}

func (a *Application) Close() {
	a.log.Sync()
}

func (a *Application) Run() {
	a.log.Info("app running")
	a.server.Run()
}

func (a *Application) buildPostgresStorage() *postgres.Storage {
	st, err := postgres.NewStorage(a.log, a.cfg.PostgresDSN())
	if err != nil {
		a.log.Fatal(err)
	}

	return st
}

func (a *Application) buildService(storage cases.Storage) *cases.BalanceService {
	svc, err := cases.NewBalanceService(a.log, storage)
	if err != nil {
		a.log.Fatal(err)
	}

	return svc
}

func (a *Application) buildServer(svc *cases.BalanceService) *http.Server {
	srv, err := http.NewServer(a.log, svc, a.cfg.ServerPort())
	if err != nil {
		a.log.Fatal(err)
	}

	return srv
}

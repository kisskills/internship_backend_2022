package application

import (
	"context"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"service/internal/adapters/storage/postgres"
	"service/internal/cases"
	"service/internal/config"
	"service/internal/ports/http"
	"syscall"
)

type Application struct {
	cancel  context.CancelFunc
	log     *zap.SugaredLogger
	storage *postgres.Storage
	cfg     *config.Config
	server  *http.Server
}

func (a *Application) Build(configPath string) {
	var err error

	a.log = a.initConfig()

	a.cfg, err = config.NewConfig(configPath)
	if err != nil {
		a.log.Fatal("init config")
	}

	a.storage = a.buildPostgresStorage()

	svc := a.buildService(a.storage)

	a.server = a.buildServer(svc)
}

func (a *Application) Run() {
	a.log.Info("application started")
	defer a.log.Info("application stopped")

	var ctx context.Context

	ctx, a.cancel = context.WithCancel(context.Background())
	defer a.cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		select {
		case <-sig:
		case <-ctx.Done():
		}

		a.Stop()
	}()

	a.server.Run(ctx)
}

func (a *Application) Stop() {
	a.storage.Close()
	a.cancel()
	_ = a.log.Sync()
}

func (a *Application) initConfig() *zap.SugaredLogger {
	logger, err := zap.NewProduction()
	if err != nil {
		a.log.Fatal(err)
	}

	return logger.Sugar()
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

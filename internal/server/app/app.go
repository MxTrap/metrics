package app

import (
	"context"
	"github.com/MxTrap/metrics/config/serverconfig"
	"github.com/MxTrap/metrics/internal/server/grpc"
	"github.com/MxTrap/metrics/internal/server/httpserver"
	"github.com/MxTrap/metrics/internal/server/httpserver/handlers"
	"github.com/MxTrap/metrics/internal/server/logger"
	"github.com/MxTrap/metrics/internal/server/migrator"
	"github.com/MxTrap/metrics/internal/server/repository"
	"github.com/MxTrap/metrics/internal/server/repository/postgres"
	"github.com/MxTrap/metrics/internal/server/service"
	"github.com/MxTrap/metrics/internal/utils"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"sync"
)

type App struct {
	httpServer     *httpserver.HTTPServer
	grpcServer     *grpc.Server
	metricsService *service.MetricsService
	logger         *logger.Logger
}

func NewApp(cfg *serverconfig.ServerConfig, ctx context.Context) (*App, error) {
	log := logger.NewLogger()

	fileStorage := repository.NewMetricsFileStorage(cfg.FileStoragePath)
	var storage service.Storage
	var storageErr error
	storage, storageErr = repository.NewMemStorage()
	if cfg.DatabaseDSN != "" {
		pgPool, err := pgxpool.New(ctx, cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		m, err := migrator.NewMigrator(pgPool, utils.GetProjectPath()+"/migrations")
		if err != nil {
			log.Logger.Error("could not create migrator ", err)
			return nil, err
		}
		err = m.InitializeDB()

		if err != nil {
			log.Logger.Error("could not initialize database ", err)
			return nil, err
		}
		storage, storageErr = postgres.NewPostgresStorage(pgPool, log)
	}
	if storageErr != nil {
		log.Logger.Error(storageErr)
		return nil, storageErr
	}

	metricsService := service.NewMetricsService(fileStorage, storage, cfg.StoreInterval, cfg.Restore)
	httpRouter := httpserver.NewRouter(cfg.HTTPAddr, log, cfg.Key, cfg.CryptoKey, cfg.TrustedSubnet)
	metricHandler := handlers.NewMetricHandler(metricsService, httpRouter.Router)
	metricHandler.RegisterRoutes()
	grpcServer := grpc.NewGRPCServer(cfg.GRPCAddr, log.LoggerInterceptor, cfg.TrustedSubnet)
	grpcServer.Register(grpc.NewMetricsServiceServer(metricsService))

	return &App{
		httpServer:     httpRouter,
		logger:         log,
		metricsService: metricsService,
		grpcServer:     grpcServer,
	}, nil
}

func (a App) Run(ctx context.Context) error {
	a.logger.Logger.Info("starting server")
	err := a.metricsService.Start(ctx)
	if err != nil {
		a.logger.Logger.Error(err.Error())
		return err
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		err = a.httpServer.Run()
		wg.Done()
	}()

	go func() {
		err = a.grpcServer.Run()
		wg.Done()
	}()

	wg.Wait()

	if err != nil {
		a.logger.Logger.Error(err)
	}

	return err
}

func (a App) GracefulShutdown(ctx context.Context) error {
	a.logger.Logger.Info("shutting down server")
	a.metricsService.Stop()
	err := a.httpServer.Stop(ctx)
	if err != nil {
		a.logger.Logger.Error(err.Error())
		return err
	}
	return nil
}

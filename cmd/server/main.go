package main

import (
	"context"
	"log"

	"go.uber.org/zap"

	"github.com/your-org/service-name/config"
	pkgcache "github.com/your-org/service-name/pkg/cache"
	pkgdb "github.com/your-org/service-name/pkg/db"
	"github.com/your-org/service-name/pkg/httpserver"
	pkgkafka "github.com/your-org/service-name/pkg/kafka"
	"github.com/your-org/service-name/pkg/logger"
	pkgstorage "github.com/your-org/service-name/pkg/storage"

	httphandler "github.com/your-org/service-name/internal/handler/http"
	kafkahandler "github.com/your-org/service-name/internal/handler/kafka"
	dbrepo "github.com/your-org/service-name/internal/repository/db"
	"github.com/your-org/service-name/internal/router"
	productusecase "github.com/your-org/service-name/internal/usecase/product"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	if err := logger.Init(cfg.Log); err != nil {
		log.Fatalf("logger init failed: %v", err)
	}

	db, err := pkgdb.NewGORM(cfg.DB)
	if err != nil {
		zap.L().Fatal("db init failed", zap.Error(err))
	}

	rdb, err := pkgcache.NewRedis(cfg.Cache)
	if err != nil {
		zap.L().Fatal("redis init failed", zap.Error(err))
	}

	store, err := pkgstorage.NewStorage(cfg.Storage)
	if err != nil {
		zap.L().Fatal("storage init failed", zap.Error(err))
	}

	// Producer — wire into your usecase or a dedicated event publisher when needed:
	// producer, err := pkgkafka.NewProducer(cfg.Kafka)
	// if err != nil { zap.L().Fatal("kafka producer init failed", zap.Error(err)) }
	// defer producer.Close()

	consumer, err := pkgkafka.NewConsumer(cfg.Kafka)
	if err != nil {
		zap.L().Fatal("kafka consumer init failed", zap.Error(err))
	}
	defer consumer.Close()

	repo := dbrepo.NewProductRepo(db, rdb, store)
	uc := productusecase.New(repo)

	r := router.New(httphandler.NewProductHandler(uc))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go consumer.Run(ctx, kafkahandler.NewProductConsumer(uc).Handle)

	httpserver.NewServer(cfg.Server, r).Start()
}

package main

import (
	"context"
	"log"
	"stori-challenge/internal/infra/bootstrap"
	"stori-challenge/internal/infra/config"
	"stori-challenge/internal/infra/logger"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var appCtx *bootstrap.AppContext

func init() {
	if err := logger.Init(); err != nil {
		log.Fatalf("error iniciando logger: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Logger.Fatal("error cargando configuración", zap.Error(err))
	}

	logger.Logger.Info("configuración cargada",
		zap.String("db_host", cfg.DBHost),
		zap.String("db_name", cfg.DBName),
		zap.String("s3_bucket", cfg.S3BucketName),
		zap.String("s3_region", cfg.S3Region),
		zap.String("ssl_mode", cfg.DBSSLMode),
	)

	appCtx, err = bootstrap.InitializeApp(cfg)
	if err != nil {
		logger.Logger.Fatal("error inicializando aplicación", zap.Error(err))
	}
}

func processS3Record(ctx context.Context, rec events.S3EventRecord) error {
	bucket := rec.S3.Bucket.Name
	key := rec.S3.Object.Key

	logger.Logger.Info("procesando objeto S3",
		zap.String("bucket", bucket),
		zap.String("key", key),
	)

	if err := appCtx.SummaryUseCase.ProcessTransactionsFromObject(ctx, bucket, key); err != nil {
		logger.Logger.Error("error procesando transacciones",
			zap.String("bucket", bucket),
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func handler(ctx context.Context, evt events.S3Event) error {
	logger.Logger.Info("evento S3 recibido",
		zap.Int("records", len(evt.Records)),
	)

	if len(evt.Records) == 1 {
		if err := processS3Record(ctx, evt.Records[0]); err != nil {
			return err
		}
		logger.Logger.Info("evento S3 procesado correctamente")
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(4)

	for _, rec := range evt.Records {
		r := rec
		g.Go(func() error {
			return processS3Record(ctx, r)
		})
	}

	if err := g.Wait(); err != nil {
		logger.Logger.Error("falló al menos un procesamiento de objeto S3", zap.Error(err))
		return err
	}

	logger.Logger.Info("evento S3 procesado correctamente (procesamiento concurrente)")
	return nil
}

func main() {
	defer logger.Sync()
	log.Println("Lambda S3 de Stori iniciando...")
	lambda.Start(handler)
}

package bootstrap

import (
	"context"
	"stori-challenge/internal/core/ports/in"
	"stori-challenge/internal/core/ports/out"
	"stori-challenge/internal/infra/database"
	"stori-challenge/internal/interfaces/out/rds"

	"stori-challenge/internal/core/application"
	"stori-challenge/internal/infra/config"
	"stori-challenge/internal/interfaces/out/csvreader"
	"stori-challenge/internal/interfaces/out/email"
	"stori-challenge/internal/interfaces/out/rds/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

type AppContext struct {
	SummaryUseCase in.SummaryUseCase
}

func InitializeApp(cfg *config.Config) (*AppContext, error) {
	ctx := context.Background()

	awsCfg, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(cfg.S3Region),
	)
	if err != nil {
		return nil, err
	}

	if cfg.AWSEndpointURL != "" {
		awsCfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               cfg.AWSEndpointURL,
					HostnameImmutable: true,
				}, nil
			},
		)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})

	var emailSender out.EmailSender

	if cfg.AWSEndpointURL != "" {
		emailSender = email.NewNoopEmailSender(cfg)
	} else {
		sesClient := sesv2.NewFromConfig(awsCfg)
		emailSender = email.NewSESEmailSender(sesClient, cfg)
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.Transaction{}, &models.AccountSummary{}); err != nil {
		return nil, err
	}

	txReader := csvreader.NewS3CSVReader(s3Client)
	txRepo := rds.NewTransactionRepo(db)

	summaryService := application.NewSummaryService(
		txReader,
		emailSender,
		txRepo,
	)

	return &AppContext{
		SummaryUseCase: summaryService,
	}, nil
}

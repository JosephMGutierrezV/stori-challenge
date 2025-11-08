package bootstrap

import (
	"testing"

	"stori-challenge/internal/infra/config"
)

func TestInitializeApp_WithPostgresAndLocalstack(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5434")
	t.Setenv("DB_USER", "app")
	t.Setenv("DB_PASSWORD", "app")
	t.Setenv("DB_NAME", "app")
	t.Setenv("DB_SCHEMA", "public")

	t.Setenv("S3_BUCKET_NAME", "stori-transactions-local")
	t.Setenv("S3_REGION", "us-east-1")
	t.Setenv("SES_FROM", "no-reply@stori-local.test")
	t.Setenv("EMAIL_DEFAULT", "integration-test@local.test")

	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_ENDPOINT_URL", "http://localhost:4566")
	t.Setenv("AWS_S3_USE_PATH_STYLE", "true")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	appCtx, err := InitializeApp(cfg)
	if err != nil {
		t.Fatalf("InitializeApp returned error: %v", err)
	}

	if appCtx == nil {
		t.Fatalf("InitializeApp returned nil AppContext")
	}
	if appCtx.SummaryUseCase == nil {
		t.Fatalf("SummaryService is nil in AppContext")
	}
}

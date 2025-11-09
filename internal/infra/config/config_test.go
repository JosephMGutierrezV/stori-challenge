package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
}

func TestLoadConfig_SuccessFromEnv(t *testing.T) {
	resetViper(t)

	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "appuser")
	t.Setenv("DB_PASSWORD", "s3cr3t")
	t.Setenv("DB_NAME", "stori_db")

	t.Setenv("S3_BUCKET_NAME", "stori-transactions-local")
	t.Setenv("S3_REGION", "us-east-1")
	t.Setenv("SES_FROM", "no-reply@stori-local.test")
	t.Setenv("EMAIL_DEFAULT", "user@example.com")

	t.Setenv("AWS_ENDPOINT_URL", "http://localstack:4566")
	t.Setenv("AWS_S3_USE_PATH_STYLE", "true")
	t.Setenv("STORI_LOGO_URL", "http://example.com/logo.png")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.DBHost != "localhost" {
		t.Errorf("DBHost = %q, want %q", cfg.DBHost, "localhost")
	}
	if cfg.DBPort != "5432" {
		t.Errorf("DBPort = %q, want %q", cfg.DBPort, "5432")
	}
	if cfg.DBUser != "appuser" {
		t.Errorf("DBUser = %q, want %q", cfg.DBUser, "appuser")
	}
	if cfg.DBPassword != "s3cr3t" {
		t.Errorf("DBPassword = %q, want %q", cfg.DBPassword, "s3cr3t")
	}
	if cfg.DBName != "stori_db" {
		t.Errorf("DBName = %q, want %q", cfg.DBName, "stori_db")
	}

	if cfg.DBSchema != "public" {
		t.Errorf("DBSchema = %q, want %q (default)", cfg.DBSchema, "public")
	}

	if cfg.S3BucketName != "stori-transactions-local" {
		t.Errorf("S3BucketName = %q, want %q", cfg.S3BucketName, "stori-transactions-local")
	}
	if cfg.S3Region != "us-east-1" {
		t.Errorf("S3Region = %q, want %q", cfg.S3Region, "us-east-1")
	}
	if cfg.SESFrom != "no-reply@stori-local.test" {
		t.Errorf("SESFrom = %q, want %q", cfg.SESFrom, "no-reply@stori-local.test")
	}
	if cfg.EmailDefault != "user@example.com" {
		t.Errorf("EmailDefault = %q, want %q", cfg.EmailDefault, "user@example.com")
	}

	if cfg.AWSEndpointURL != "http://localstack:4566" {
		t.Errorf("AWSEndpointURL = %q, want %q", cfg.AWSEndpointURL, "http://localstack:4566")
	}
	if cfg.UsePathStyle != true {
		t.Errorf("UsePathStyle = %v, want true", cfg.UsePathStyle)
	}

	if cfg.StoriLogoURL != "http://example.com/logo.png" {
		t.Errorf("StoriLogoURL = %q, want %q", cfg.StoriLogoURL, "http://example.com/logo.png")
	}
}

func TestLoadConfig_MissingRequiredVariables(t *testing.T) {
	resetViper(t)

	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "appuser")
	t.Setenv("DB_PASSWORD", "s3cr3t")
	t.Setenv("DB_NAME", "stori_db")
	t.Setenv("S3_BUCKET_NAME", "stori-transactions-local")
	t.Setenv("S3_REGION", "us-east-1")
	t.Setenv("SES_FROM", "no-reply@stori-local.test")
	t.Setenv("EMAIL_DEFAULT", "user@example.com")
	t.Setenv("AWS_ENDPOINT_URL", "http://localstack:4566")
	t.Setenv("STORI_LOGO_URL", "http://example.com/logo.png")

	_, err := LoadConfig()
	if err == nil {
		t.Fatalf("expected error when DB_HOST is missing, got nil")
	}

	if !strings.Contains(err.Error(), "DB_HOST") {
		t.Fatalf("expected error message to mention DB_HOST, got: %v", err)
	}
}

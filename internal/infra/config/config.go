package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	DBHost         string `mapstructure:"DB_HOST"`
	DBPort         string `mapstructure:"DB_PORT"`
	DBUser         string `mapstructure:"DB_USER"`
	DBPassword     string `mapstructure:"DB_PASSWORD"`
	DBName         string `mapstructure:"DB_NAME"`
	DBSchema       string `mapstructure:"DB_SCHEMA"`
	S3BucketName   string `mapstructure:"S3_BUCKET_NAME"`
	S3Region       string `mapstructure:"S3_REGION"`
	SESFrom        string `mapstructure:"SES_FROM"`
	EmailDefault   string `mapstructure:"EMAIL_DEFAULT"`
	AWSEndpointURL string `mapstructure:"AWS_ENDPOINT_URL"`
	UsePathStyle   bool   `mapstructure:"AWS_S3_USE_PATH_STYLE"`
	StoriLogoURL   string `mapstructure:"STORI_LOGO_URL"`
	DBSSLMode      string `mapstructure:"DB_SSL_MODE"`
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SCHEMA", "public")
	viper.SetDefault("EMAIL_DEFAULT", "josephmauricio23@hotmail.com")
	viper.SetDefault("AWS_S3_USE_PATH_STYLE", false)
	viper.SetDefault("STORI_LOGO_URL", "https://media.licdn.com/dms/image/v2/D4E0BAQHuxJutLmsBFQ/company-logo_200_200/company-logo_200_200/0/1700583469952?e=1764201600&v=beta&t=yAwe1j0mbzSEM19MZSGWYt1RWiD9l7rPcgjSxGZSp_Q")
	viper.SetDefault("DB_SSL_MODE", "disable")

	for _, k := range []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD",
		"DB_NAME", "DB_SCHEMA",
		"S3_BUCKET_NAME", "S3_REGION",
		"SES_FROM", "EMAIL_DEFAULT",
		"AWS_ENDPOINT_URL", "AWS_S3_USE_PATH_STYLE",
		"STORI_LOGO_URL",
		"DB_SSL_MODE",
	} {
		_ = viper.BindEnv(k)
	}

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		viper.SetConfigType("env")
		_ = viper.ReadInConfig()
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	missing := []string{}
	req := func(k, v string) {
		if strings.TrimSpace(v) == "" {
			missing = append(missing, k)
		}
	}
	req("DB_HOST", cfg.DBHost)
	req("DB_USER", cfg.DBUser)
	req("DB_PASSWORD", cfg.DBPassword)
	req("DB_NAME", cfg.DBName)
	req("DB_PORT", cfg.DBPort)
	req("S3_BUCKET_NAME", cfg.S3BucketName)
	req("S3_REGION", cfg.S3Region)
	req("SES_FROM", cfg.SESFrom)
	req("EMAIL_DEFAULT", cfg.EmailDefault)
	req("STORI_LOGO_URL", cfg.StoriLogoURL)

	if len(missing) > 0 {
		return nil, fmt.Errorf("faltan variables: %v", missing)
	}

	return &cfg, nil
}

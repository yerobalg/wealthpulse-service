package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"github.com/yerobalg/wealthpulse-service/helper/async"
	"github.com/yerobalg/wealthpulse-service/helper/cryptolib"
	"github.com/yerobalg/wealthpulse-service/helper/db"
	"github.com/yerobalg/wealthpulse-service/helper/httpclient"
	"github.com/yerobalg/wealthpulse-service/helper/logger"

	"github.com/yerobalg/wealthpulse-service/src/entity"
	"github.com/yerobalg/wealthpulse-service/src/handler"
	"github.com/yerobalg/wealthpulse-service/src/repository"
	"github.com/yerobalg/wealthpulse-service/src/usecase"
)

// @title Go API Boilerplate
// @description A layered Go REST API boilerplate (Gin + GORM + JWT)
// @version 1.0

// @contact.name 	Yerobal Gustaf Sekeon
// @contact.email 	yerobalg@gmail.com

// @securitydefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @value Bearer {token}

// httpClientTimeout caps every outbound market-data request.
const httpClientTimeout = 10 * time.Second

func main() {
	loadEnv()
	initialize()
}

func loadEnv() {
	if os.Getenv("ENV") == "production" || os.Getenv("ENV") == "staging" {
		return
	}

	if err := godotenv.Load(); err != nil {
		panic(err)
	}
}

func initialize() {
	ctx := context.Background()

	// logger
	logger := logger.Init()

	// db
	dbCred := db.Credential{
		Path: os.Getenv("DB_PATH"),
	}
	dbConnectionPool := getDBConnectionPool(ctx, logger)
	database, err := db.Init(logger, dbCred, dbConnectionPool)
	if err != nil {
		logger.Fatal(ctx, "failed to connect to database", err)
		panic(err)
	}

	// password
	password := cryptolib.InitPassword(getEnvAsInt(ctx, logger, "PASSWORD_SALT_ROUND"))

	// jwt
	jwt := cryptolib.InitJWT(int64(getEnvAsInt(ctx, logger, "JWT_EXPIRED_TIME_SEC")), os.Getenv("JWT_SECRET_KEY"))

	// async
	async := async.Init(logger)

	// http client — shared outbound gateway for market-data providers
	httpClient := httpclient.Init(httpClientTimeout)

	repo := repository.Init(repository.InitParam{
		DB:         *database,
		HTTPClient: httpClient,
		CoinGecko: repository.CoinGeckoConfig{
			BaseURL: os.Getenv("COINGECKO_BASE_URL"),
			APIKey:  os.Getenv("COINGECKO_API_KEY"),
		},
		YahooFinance: repository.YahooFinanceConfig{
			BaseURL: os.Getenv("YAHOO_FINANCE_BASE_URL"),
		},
		ExchangeRate: repository.ExchangeRateConfig{
			BaseURL: os.Getenv("EXCHANGERATE_BASE_URL"),
			AppID:   os.Getenv("EXCHANGERATE_API_KEY"),
		},
	})
	usecase := usecase.Init(usecase.InitParam{
		Repository: repo,
		Password:   password,
		JWT:        jwt,
		Async:      async,
		Log:        logger,
		TxManager:  database,
	})
	if err := seedSuperuser(ctx, usecase); err != nil {
		logger.Fatal(ctx, "failed to seed superuser", err)
		panic(err)
	}

	handler := handler.Init(handler.InitParam{
		Log:     logger,
		JWT:     jwt,
		Async:   async,
		Usecase: usecase,
		AppHost: os.Getenv("APP_HOST"),
		AppPort: os.Getenv("APP_PORT"),
	})

	handler.Run()
}

// seedSuperuser idempotently creates the single owner from the environment. The
// pre-bcrypt-hashed SUPERUSER_PASSWORD_HASH is read verbatim (os.Getenv does not
// interpret '$', so the hash needs no escaping) and stored as-is.
func seedSuperuser(ctx context.Context, uc *usecase.Usecase) error {
	return uc.User.EnsureSuperuser(ctx, entity.EnsureSuperuserRequest{
		Username:     os.Getenv("SUPERUSER_USERNAME"),
		Name:         os.Getenv("SUPERUSER_NAME"),
		PasswordHash: os.Getenv("SUPERUSER_PASSWORD_HASH"),
		IsMale:       true,
	})
}

func getEnvAsInt(ctx context.Context, logger logger.Interface, key string) int {
	val, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		logger.Fatal(ctx, "failed to parse "+key, err)
		panic(err)
	}
	return val
}

func getDBConnectionPool(ctx context.Context, logger logger.Interface) db.PoolConfig {
	return db.PoolConfig{
		MaxOpenConns:    getEnvAsInt(ctx, logger, "DB_MAX_OPEN_CONNS"),
		MaxIdleConns:    getEnvAsInt(ctx, logger, "DB_MAX_IDLE_CONNS"),
		ConnMaxLifetime: time.Duration(getEnvAsInt(ctx, logger, "DB_CONN_MAX_LIFETIME_MINUTE")) * time.Second,
		ConnMaxIdleTime: time.Duration(getEnvAsInt(ctx, logger, "DB_CONN_MAX_IDLE_TIME_MINUTE")) * time.Second,
	}
}

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
	"github.com/yerobalg/wealthpulse-service/helper/logger"

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

	repo := repository.Init(*database)
	usecase := usecase.Init(usecase.InitParam{
		Repository: repo,
		Password:   password,
		JWT:        jwt,
		Async:      async,
		Log:        logger,
		TxManager:  database,
	})
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

package db

import (
	"encoding/base64"
	"fmt"
	"time"

	"go.pitz.tech/gorm/encryption"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/yerobalg/wealthpulse-service/helper/logger"
)

type DB struct {
	conn *gorm.DB
}

type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type Credential struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
}

func Init(serverLogger logger.Interface, cred Credential, poolConfig PoolConfig, encryptionKeyB64 string) (*DB, error) {
	db, err := initPostgres(serverLogger, cred, poolConfig)
	if err != nil {
		return nil, err
	}

	key, err := base64.StdEncoding.DecodeString(encryptionKeyB64)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key: %w", err)
	}
	if err := encryption.Register(db, encryption.WithKey(key), encryption.WithMigration()); err != nil {
		return nil, fmt.Errorf("failed to register encryption serializer: %w", err)
	}

	return &DB{db}, nil
}

func initPostgres(serverLogger logger.Interface, cred Credential, poolConfig PoolConfig) (*gorm.DB, error) {
	dataSourceName := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
		cred.Host,
		cred.Port,
		cred.Username,
		cred.Password,
		cred.DBName,
	)

	db, err := gorm.Open(postgres.Open(dataSourceName), &gorm.Config{
		Logger: InitGormLogger(serverLogger),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Pool configuration
	sqlDB.SetMaxIdleConns(poolConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(poolConfig.MaxOpenConns)
	sqlDB.SetConnMaxIdleTime(poolConfig.ConnMaxIdleTime)
	sqlDB.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)

	return db, nil
}

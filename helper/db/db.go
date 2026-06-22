package db

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"go.pitz.tech/gorm/encryption"
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
	Path string
}

func Init(serverLogger logger.Interface, cred Credential, poolConfig PoolConfig, encryptionKeyB64 string) (*DB, error) {
	db, err := initSQLite(serverLogger, cred, poolConfig)
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

func initSQLite(serverLogger logger.Interface, cred Credential, poolConfig PoolConfig) (*gorm.DB, error) {
	// Enable foreign keys (off by default in SQLite), WAL for concurrent reads, and a
	// busy timeout so the in-process scheduler and HTTP handlers don't trip over each other.
	dataSourceName := fmt.Sprintf(
		"file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)",
		cred.Path,
	)

	db, err := gorm.Open(sqlite.Open(dataSourceName), &gorm.Config{
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

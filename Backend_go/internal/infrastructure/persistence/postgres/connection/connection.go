package connection

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/config"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	*gorm.DB
	dsn string
}

// Reconnect attempts to reconnect to the database if the connection is lost
func (db *Database) Reconnect() error {
	newDB, err := gorm.Open(postgres.Open(db.dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to reconnect to database: %w", err)
	}

	// Update the connection
	db.DB = newDB

	// Configure the new connection
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}

	// Use default connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return nil
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	// Construct DSN
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	// First try to establish a basic SQL connection to verify connectivity
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create sql.DB: %w", err)
	}
	defer sqlDB.Close()

	// Test the connection with a short timeout
	sqlDB.SetConnMaxLifetime(10 * time.Second)
	err = sqlDB.Ping()
	if err != nil {
		sqlErr, ok := err.(*pq.Error)
		if ok {
			return nil, fmt.Errorf("postgres error: code=%s, message=%s, detail=%s", sqlErr.Code, sqlErr.Message, sqlErr.Detail)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Now set up GORM with detailed logging and connection configured
	gormConfig := &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Info),
		PrepareStmt: true, // Enables prepared statement caching
		NowFunc: func() time.Time {
			return time.Now().UTC() // Standardize time
		},
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database with GORM: %w", err)
	}

	// Get the underlying *sql.DB to configure the connection pool
	sqlDB, err = db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}

	// Configure connection pool with reasonable defaults if not specified
	maxIdleConns := 10
	maxOpenConns := 100

	if cfg.Database.MaxIdleConns > 0 {
		maxIdleConns = cfg.Database.MaxIdleConns
	}

	if cfg.Database.MaxOpenConns > 0 {
		maxOpenConns = cfg.Database.MaxOpenConns
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Verify the connection pool is working
	err = sqlDB.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping connection pool: %w", err)
	}

	return &Database{
		DB:  db,
		dsn: dsn,
	}, nil
}

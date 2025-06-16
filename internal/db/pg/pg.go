package pg

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PGConfig struct {
	Host            string
	Port            int
	DBName          string
	DefaultDBName   string
	User            string
	Pass            string
	SSLMode         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
	Charset         string
	Extras          string
}

func InitPluginDB(config *PGConfig) (*gorm.DB, error) {
	// first try to connect to target database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Pass,
		config.DBName,
		config.SSLMode,
	)
	if config.Charset != "" {
		dsn = fmt.Sprintf("%s client_encoding=%s", dsn, config.Charset)
	}
	if config.Extras != "" {
		dsn = fmt.Sprintf("%s %s", dsn, config.Extras)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// if connection fails, try to create database
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host,
			config.Port,
			config.User,
			config.Pass,
			config.DefaultDBName,
			config.SSLMode,
		)
		if config.Charset != "" {
			dsn = fmt.Sprintf("%s client_encoding=%s", dsn, config.Charset)
		}
		if config.Extras != "" {
			dsn = fmt.Sprintf("%s %s", dsn, config.Extras)
		}

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}

		pgsqlDB, err := db.DB()
		if err != nil {
			return nil, err
		}
		defer pgsqlDB.Close()

		// check if the db exists
		rows, err := pgsqlDB.Query(fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", config.DBName))
		if err != nil {
			return nil, err
		}

		if !rows.Next() {
			// create database
			_, err = pgsqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", config.DBName))
			if err != nil {
				return nil, err
			}
		}

		// connect to the new db
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", config.Host, config.Port, config.User, config.Pass, config.DBName, config.SSLMode)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
	}

	pgsqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// check if uuid-ossp extension exists
	rows, err := pgsqlDB.Query("SELECT 1 FROM pg_extension WHERE extname = 'uuid-ossp'")
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		// create the uuid-ossp extension
		_, err = pgsqlDB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
		if err != nil {
			return nil, err
		}
	}

	// configure connection pool
	pgsqlDB.SetMaxIdleConns(config.MaxIdleConns)
	pgsqlDB.SetMaxOpenConns(config.MaxOpenConns)
	pgsqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)

	return db, nil
}

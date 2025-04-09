package pg

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitPluginDB(host string, port int, db_name string, default_db_name string, user string, pass string, sslmode string) (*gorm.DB, error) {
	// first try to connect to target database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, pass, db_name, sslmode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// if connection fails, try to create database
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, pass, default_db_name, sslmode)
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
		rows, err := pgsqlDB.Query(fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", db_name))
		if err != nil {
			return nil, err
		}

		if !rows.Next() {
			// create database
			_, err = pgsqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", db_name))
			if err != nil {
				return nil, err
			}
		}

		// connect to the new db
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, pass, db_name, sslmode)
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

	pgsqlDB.SetConnMaxIdleTime(time.Minute * 1)

	return db, nil
}

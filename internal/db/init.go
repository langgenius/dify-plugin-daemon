package db

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDifyEnterpriseDB(host string, port int, dbname string, user string, pass string, sslmode string) error {
	// create db if not exists
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, pass, dbname, sslmode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	pgsqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// check if the db exists
	rows, err := pgsqlDB.Query(fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", dbname))
	if err != nil {
		return err
	}

	if !rows.Next() {
		// create database
		_, err = pgsqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
		if err != nil {
			return err
		}
	}

	// close db
	err = pgsqlDB.Close()
	if err != nil {
		return err
	}

	// connect to the new db
	dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, pass, dbname, sslmode)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	pgsqlDB, err = db.DB()
	if err != nil {
		return err
	}

	// check if uuid-ossp extension exists
	rows, err = pgsqlDB.Query("SELECT 1 FROM pg_extension WHERE extname = 'uuid-ossp'")
	if err != nil {
		return err
	}

	if !rows.Next() {
		// create the uuid-ossp extension
		_, err = pgsqlDB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
		if err != nil {
			return err
		}
	}

	pgsqlDB.SetConnMaxIdleTime(time.Minute * 1)
	DifyPluginDB = db

	return nil
}

func AutoMigrate() error {
	return DifyPluginDB.AutoMigrate()
}

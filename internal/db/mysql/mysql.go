package mysql

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitPluginDB(host string, port int, dbName string, defaultDbName string, user string, password string, sslMode string) (*gorm.DB, error) {
	initializer := mysqlDbInitializer{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		sslMode:  sslMode,
	}

	// first try to connect to target database
	db, err := initializer.connect(dbName)
	if err != nil {
		// if connection fails, try to create database
		db, err = initializer.connect(defaultDbName)
		if err != nil {
			return nil, err
		}

		err = initializer.createDatabaseIfNotExists(db, dbName)
		if err != nil {
			return nil, err
		}

		// connect to the new db
		db, err = initializer.connect(dbName)
		if err != nil {
			return nil, err
		}
	}

	pool, err := db.DB()
	if err != nil {
		return nil, err
	}

	pool.SetConnMaxIdleTime(time.Minute * 1)

	return db, nil
}

// mysqlDbInitializer initializes database for MySQL.
type mysqlDbInitializer struct {
	host     string
	port     int
	user     string
	password string
	sslMode  string
}

func (m *mysqlDbInitializer) connect(dbName string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&tls=%v", m.user, m.password, m.host, m.port, dbName, m.sslMode == "require")
	return gorm.Open(myDialector{Dialector: mysql.Open(dsn).(*mysql.Dialector)}, &gorm.Config{})
}

func (m *mysqlDbInitializer) createDatabaseIfNotExists(db *gorm.DB, dbName string) error {
	pool, err := db.DB()
	if err != nil {
		return err
	}
	defer pool.Close()

	rows, err := pool.Query(fmt.Sprintf("SELECT 1 FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '%s'", dbName))
	if err != nil {
		return err
	}

	if !rows.Next() {
		// create database
		_, err = pool.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	}
	return err
}

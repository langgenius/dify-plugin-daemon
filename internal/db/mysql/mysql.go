package mysql

import (
	"fmt"
	"time"

	gormConfig "github.com/langgenius/dify-plugin-daemon/internal/db/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MySQLConfig struct {
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
	LogLevel        string
}

func InitPluginDB(config *MySQLConfig) (*gorm.DB, error) {
	// TODO: MySQL dose not support DB_EXTRAS now
	initializer := mysqlDbInitializer{
		host:     config.Host,
		port:     config.Port,
		user:     config.User,
		password: config.Pass,
		sslMode:  config.SSLMode,
		logLevel: config.LogLevel,
	}

	// first try to connect to target database
	db, err := initializer.connect(config.DBName)
	if err != nil {
		// if connection fails, try to create database
		db, err = initializer.connect(config.DefaultDBName)
		if err != nil {
			return nil, err
		}

		err = initializer.createDatabaseIfNotExists(db, config.DBName)
		if err != nil {
			return nil, err
		}

		// connect to the new db
		db, err = initializer.connect(config.DBName)
		if err != nil {
			return nil, err
		}
	}

	pool, err := db.DB()
	if err != nil {
		return nil, err
	}

	// configure connection pool
	pool.SetMaxIdleConns(config.MaxIdleConns)
	pool.SetMaxOpenConns(config.MaxOpenConns)
	pool.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)

	return db, nil
}

// mysqlDbInitializer initializes database for MySQL.
type mysqlDbInitializer struct {
	host     string
	port     int
	user     string
	password string
	sslMode  string
	logLevel string
}

func (m *mysqlDbInitializer) connect(dbName string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&tls=%v", m.user, m.password, m.host, m.port, dbName, m.sslMode == "require")

	config := &gorm.Config{}
	if m.logLevel != "" {
		config.Logger = logger.Default.LogMode(gormConfig.GetGormLogLevel(m.logLevel))
	}
	return gorm.Open(myDialector{Dialector: mysql.Open(dsn).(*mysql.Dialector)}, config)
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

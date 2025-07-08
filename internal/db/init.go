package db

import (
	"github.com/langgenius/dify-plugin-daemon/internal/db/mysql"
	"github.com/langgenius/dify-plugin-daemon/internal/db/pg"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func autoMigrate() error {
	err := DifyPluginDB.AutoMigrate(
		models.Plugin{},
		models.PluginInstallation{},
		models.PluginDeclaration{},
		models.Endpoint{},
		models.ServerlessRuntime{},
		models.DatasourceInstallation{},
		models.ToolInstallation{},
		models.AIModelInstallation{},
		models.InstallTask{},
		models.TenantStorage{},
		models.AgentStrategyInstallation{},
	)

	if err != nil {
		return err
	}

	// check if "declaration" column exists in Plugin/ServerlessRuntime/ToolInstallation/AIModelInstallation/AgentStrategyInstallation
	// drop the "declaration" column not null constraint if exists
	ignoreDeclarationColumn := func(table string) error {
		if DifyPluginDB.Migrator().HasColumn(table, "declaration") {
			// remove NOT NULL constraint on declaration column
			if err := DifyPluginDB.Exec("ALTER TABLE " + table + " ALTER COLUMN declaration DROP NOT NULL").Error; err != nil {
				return err
			}
		}
		return nil
	}

	tables := []string{
		"plugins",
		"serverless_runtimes",
		"tool_installations",
		"ai_model_installations",
		"agent_strategy_installations",
	}

	for _, table := range tables {
		if err := ignoreDeclarationColumn(table); err != nil {
			return err
		}
	}

	return nil
}

func Init(config *app.Config) {
	var err error
	if config.DBType == "postgresql" {
		DifyPluginDB, err = pg.InitPluginDB(&pg.PGConfig{
			Host:            config.DBHost,
			Port:            int(config.DBPort),
			DBName:          config.DBDatabase,
			DefaultDBName:   config.DBDefaultDatabase,
			User:            config.DBUsername,
			Pass:            config.DBPassword,
			SSLMode:         config.DBSslMode,
			MaxIdleConns:    config.DBMaxIdleConns,
			MaxOpenConns:    config.DBMaxOpenConns,
			ConnMaxLifetime: config.DBConnMaxLifetime,
			Charset:         config.DBCharset,
			Extras:          config.DBExtras,
		})
	} else if config.DBType == "mysql" {
		DifyPluginDB, err = mysql.InitPluginDB(&mysql.MySQLConfig{
			Host:            config.DBHost,
			Port:            int(config.DBPort),
			DBName:          config.DBDatabase,
			DefaultDBName:   config.DBDefaultDatabase,
			User:            config.DBUsername,
			Pass:            config.DBPassword,
			SSLMode:         config.DBSslMode,
			MaxIdleConns:    config.DBMaxIdleConns,
			MaxOpenConns:    config.DBMaxOpenConns,
			ConnMaxLifetime: config.DBConnMaxLifetime,
			Charset:         config.DBCharset,
			Extras:          config.DBExtras,
		})
	} else {
		log.Panic("unsupported database type: %v", config.DBType)
	}

	if err != nil {
		log.Panic("failed to init dify plugin db: %v", err)
	}

	err = autoMigrate()
	if err != nil {
		log.Panic("failed to auto migrate: %v", err)
	}

	log.Info("dify plugin db initialized")
}

func Close() {
	db, err := DifyPluginDB.DB()
	if err != nil {
		log.Error("failed to close dify plugin db: %v", err)
		return
	}

	err = db.Close()
	if err != nil {
		log.Error("failed to close dify plugin db: %v", err)
		return
	}

	log.Info("dify plugin db closed")
}

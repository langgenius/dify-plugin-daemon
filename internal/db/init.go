package db

import (
	"github.com/langgenius/dify-plugin-daemon/internal/db/mysql"
	"github.com/langgenius/dify-plugin-daemon/internal/db/pg"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

func cleanupDuplicatePluginInstallations() error {
	// Check if plugin_installations table exists first
	if !DifyPluginDB.Migrator().HasTable("plugin_installations") {
		// Table doesn't exist yet, no need to clean up
		return nil
	}

	// Check if there are any duplicates first
	var count int64
	if err := DifyPluginDB.Raw(`
		SELECT COUNT(*) FROM (
			SELECT tenant_id, plugin_id
			FROM plugin_installations
			GROUP BY tenant_id, plugin_id
			HAVING COUNT(*) > 1
		) AS duplicates
	`).Scan(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		log.Info("Found %d duplicate (tenant_id, plugin_id) groups in plugin_installations, cleaning up...", count)

		var deleteSQL string
		switch DifyPluginDB.Dialector.Name() {
		case "postgres":
			deleteSQL = `
				DELETE FROM plugin_installations
				WHERE id NOT IN (
					SELECT DISTINCT ON (tenant_id, plugin_id) id
					FROM plugin_installations
					ORDER BY tenant_id, plugin_id, created_at DESC
				)
			`
		case "mysql":
			// This query requires MySQL 8.0+ due to the use of window functions (ROW_NUMBER).
			deleteSQL = `
				DELETE FROM plugin_installations
				WHERE id IN (
					SELECT id FROM (
						SELECT id, ROW_NUMBER() OVER (PARTITION BY tenant_id, plugin_id ORDER BY created_at DESC, id DESC) as rn
						FROM plugin_installations
					) t
					WHERE rn > 1
				)
			`
		default:
			// Log a warning and proceed. The migration will fail if duplicates exist, which is the original behavior for unsupported DBs.
			log.Warn("unsupported database dialect for duplicate cleanup: %s. Skipping cleanup.", DifyPluginDB.Dialector.Name())
			return nil
		}

		if err := DifyPluginDB.Exec(deleteSQL).Error; err != nil {
			return err
		}

		log.Info("Successfully cleaned up duplicate plugin installation records")
	}

	return nil
}

func autoMigrate() error {
	// Clean up duplicate plugin_installations before auto-migrate to avoid unique index creation failure
	if err := cleanupDuplicatePluginInstallations(); err != nil {
		return err
	}

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
		models.TriggerInstallation{},
		models.PluginReadme{},
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
		"trigger_installations",
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
	if config.DBType == app.DB_TYPE_POSTGRESQL {
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
	} else if config.DBType == app.DB_TYPE_MYSQL {
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

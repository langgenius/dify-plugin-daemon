package dbintegration_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestTransactionAcrossDatabases(t *testing.T) {
	cases := []struct {
		name   string
		config app.Config
	}{
		{
			name: "postgresql",
			config: app.Config{
				DBType:     app.DB_TYPE_POSTGRESQL,
				DBUsername: "postgres",
				DBPassword: "difyai123456",
				DBHost:     "localhost",
				DBPort:     5432,
				DBDatabase: "dify_plugin_daemon",
				DBSslMode:  "disable",
			},
		},
		{
			name: "pgbouncer",
			config: app.Config{
				DBType:     app.DB_TYPE_PG_BOUNCER,
				DBUsername: "postgres",
				DBPassword: "difyai123456",
				DBHost:     "localhost",
				DBPort:     6432,
				DBDatabase: "dify_plugin_daemon",
				DBSslMode:  "disable",
			},
		},
		{
			name: "mysql",
			config: app.Config{
				DBType:     app.DB_TYPE_MYSQL,
				DBUsername: "root",
				DBPassword: "difyai123456",
				DBHost:     "localhost",
				DBPort:     3306,
				DBDatabase: "dify_plugin_daemon",
				DBSslMode:  "disable",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.config.SetDefault()
			db.Init(&tc.config)
			t.Cleanup(db.Close)

			model := &models.ToolInstallation{
				PluginID:               uuid.New().String(),
				PluginUniqueIdentifier: uuid.New().String(),
				TenantID:               uuid.New().String(),
				Provider:               "provider_xxx",
			}

			if err := db.WithTransaction(func(tx *gorm.DB) error {
				return db.Create(model, tx)
			}); err != nil {
				t.Fatal(err.Error())
			}

			assert.NotEmpty(t, model.ID)
			assert.NotEmpty(t, model.CreatedAt)
			assert.NotEmpty(t, model.UpdatedAt)

			var record models.ToolInstallation
			if err := db.WithTransaction(func(tx *gorm.DB) error {
				row, err := db.GetOne[models.ToolInstallation](
					db.WithTransactionContext(tx),
					db.Equal("plugin_unique_identifier", model.PluginUniqueIdentifier),
					db.Equal("tenant_id", model.TenantID),
					db.WLock(),
				)
				record = row
				return err
			}); err != nil {
				t.Fatal(err.Error())
			}

			assert.Equal(t, model.ID, record.ID)
			assert.Equal(t, model.TenantID, record.TenantID)
			assert.Equal(t, model.Provider, record.Provider)
			assert.Equal(t, model.PluginUniqueIdentifier, record.PluginUniqueIdentifier)
			assert.Equal(t, model.PluginID, record.PluginID)

			newProvider := "provider_yyy"
			model.Provider = newProvider
			if err := db.WithTransaction(func(tx *gorm.DB) error {
				return db.Update(model, tx)
			}); err != nil {
				t.Fatal(err.Error())
			}

			rows, err := db.GetAll[models.ToolInstallation](db.Equal("id", model.ID))
			if err != nil {
				t.Fatal(err.Error())
			}
			if len(rows) != 1 {
				t.Fatal("expected 1 row")
			}

			updated := rows[0]
			assert.Equal(t, newProvider, updated.Provider)

			if err = db.WithTransaction(func(tx *gorm.DB) error {
				return db.Delete(model, tx)
			}); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

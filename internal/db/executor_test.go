package db

import (
	"testing"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestTransactionOnPg(t *testing.T) {
	testTransaction(t, &app.Config{
		DBType:     "postgresql",
		DBUsername: "postgres",
		DBPassword: "difyai123456",
		DBHost:     "0.0.0.0",
		DBPort:     5432,
		DBDatabase: "testing",
		DBSslMode:  "disable",
	})
}

func TestTransactionOnMySQL(t *testing.T) {
	testTransaction(t, &app.Config{
		DBType:     "mysql",
		DBUsername: "root",
		DBPassword: "difyai123456",
		DBHost:     "0.0.0.0",
		DBPort:     3306,
		DBDatabase: "testing",
		DBSslMode:  "disable",
	})
}

func testTransaction(t *testing.T, config *app.Config) {
	config.SetDefault()
	Init(config)
	defer Close()

	model := &models.ToolInstallation{
		PluginID:               uuid.New().String(),
		PluginUniqueIdentifier: "plugin_xxx",
		TenantID:               uuid.New().String(),
		Provider:               "provider_xxx",
	}

	// create
	if err := WithTransaction(func(tx *gorm.DB) error {
		return Create(model, tx)
	}); err != nil {
		t.Fatal(err.Error())
	}

	// check columns with default value
	assert.NotEmpty(t, model.ID)
	assert.NotEmpty(t, model.CreatedAt)
	assert.NotEmpty(t, model.UpdatedAt)

	// get one
	var record models.ToolInstallation
	if err := WithTransaction(func(tx *gorm.DB) error {
		row, err := GetOne[models.ToolInstallation](
			WithTransactionContext(tx),
			Equal("plugin_unique_identifier", model.PluginUniqueIdentifier),
			Equal("tenant_id", model.TenantID),
			WLock(),
		)
		record = row
		return err
	}); err != nil {
		t.Fatal(err.Error())
	}

	// check all fields
	assert.Equal(t, model.ID, record.ID)
	assert.Equal(t, model.CreatedAt.Second(), record.CreatedAt.Second())
	assert.Equal(t, model.UpdatedAt.Second(), record.UpdatedAt.Second())
	assert.Equal(t, model.TenantID, record.TenantID)
	assert.Equal(t, model.Provider, record.Provider)
	assert.Equal(t, model.PluginUniqueIdentifier, record.PluginUniqueIdentifier)
	assert.Equal(t, model.PluginID, record.PluginID)

	// update
	newProvider := "provider_yyy"
	model.Provider = newProvider
	if err := WithTransaction(func(tx *gorm.DB) error {
		return Update(model, tx)
	}); err != nil {
		t.Fatal(err.Error())
	}

	// get all
	rows, err := GetAll[models.ToolInstallation](Equal("id", model.ID))
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(rows) != 1 {
		t.Fatal("expected 1 row")
	}

	// check updated column
	updated := rows[0]
	assert.Equal(t, newProvider, updated.Provider)

	// delete
	if err = WithTransaction(func(tx *gorm.DB) error {
		return Delete(model, tx)
	}); err != nil {
		t.Fatal(err.Error())
	}
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID        string    `gorm:"column:id;primaryKey;type:uuid;" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate sets the ID field to a new UUID v7 value before creating a new record.
// ISSUE: https://github.com/langgenius/dify-plugin-daemon/issues/469
//
// to support both pgsql17 and mysql, `generate_uuid_v4` was removed from the database side.
// as it's not compatible to some of the DB versions.
func (m *Model) BeforeCreate(tx *gorm.DB) error {
	uuid, err := uuid.NewV7()
	if err != nil {
		return err
	}
	m.ID = uuid.String()

	return nil
}

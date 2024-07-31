package models

import (
	"time"
)

type Model struct {
	ID        string `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

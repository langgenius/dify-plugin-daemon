package models

import "time"

type PluginInvocation struct {
	Model
	PluginUniqueIdentifier string    `json:"plugin_unique_identifier" gorm:"index;size:255"`
	PluginID               string    `json:"plugin_id" gorm:"index;size:255"`
	InvocationCount        int       `json:"invocation_count" gorm:"default:0"`
	Timestamp              time.Time `json:"timestamp" gorm:"index"`
}

package models

// PypiMirror is a candidate PyPI index (pip -i value) persisted in the database.
//
// At most one row has Selected=true, which represents the globally selected
// mirror used when installing plugin dependencies. Rows are also used as custom
// mirror candidates added at runtime.
type PypiMirror struct {
	Model
	Name     string `json:"name" gorm:"size:127;default:''"`
	URL      string `json:"url" gorm:"size:512;uniqueIndex"`
	Selected bool   `json:"selected" gorm:"default:false;index"`
}

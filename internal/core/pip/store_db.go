package pip

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"gorm.io/gorm"
)

// dbStore is a MirrorStore backed by the plugin daemon database. It persists a
// single globally-selected mirror (the row with selected=true) and treats every
// row as a custom mirror candidate.
type dbStore struct{}

// newDBStore builds a database-backed MirrorStore.
func newDBStore() *dbStore { return &dbStore{} }

// SelectedMirror returns the row marked as selected, if any.
func (s *dbStore) SelectedMirror() (Mirror, bool, error) {
	if db.DifyPluginDB == nil {
		return Mirror{}, false, nil
	}
	row, err := db.GetOne[models.PypiMirror](db.Equal("selected", true))
	if errors.Is(err, db.ErrDatabaseNotFound) {
		return Mirror{}, false, nil
	}
	if err != nil {
		return Mirror{}, false, err
	}
	return Mirror{Name: row.Name, URL: row.URL}, true, nil
}

// CustomMirrors returns all persisted mirror rows as candidates.
func (s *dbStore) CustomMirrors() ([]Mirror, error) {
	if db.DifyPluginDB == nil {
		return nil, nil
	}
	rows, err := db.GetAll[models.PypiMirror]()
	if err != nil {
		return nil, err
	}
	mirrors := make([]Mirror, 0, len(rows))
	for _, row := range rows {
		mirrors = append(mirrors, Mirror{Name: row.Name, URL: row.URL})
	}
	return mirrors, nil
}

// Select marks the given mirror as the single selected one, inserting it as a
// custom mirror when it does not already exist. All other rows are unselected.
func (s *dbStore) Select(mirror Mirror) error {
	if db.DifyPluginDB == nil {
		return errors.New("database not initialized")
	}
	return db.WithTransaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.PypiMirror{}).
			Where("selected = ?", true).
			Update("selected", false).Error; err != nil {
			return err
		}

		var existing models.PypiMirror
		err := tx.Where("url = ?", mirror.URL).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&models.PypiMirror{
				Name:     mirror.Name,
				URL:      mirror.URL,
				Selected: true,
			}).Error
		}
		if err != nil {
			return err
		}

		existing.Selected = true
		if mirror.Name != "" {
			existing.Name = mirror.Name
		}
		return tx.Save(&existing).Error
	})
}

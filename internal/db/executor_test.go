package db

import (
	"errors"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/db/pg"
	"gorm.io/gorm"
)

func TestTransaction(t *testing.T) {
	var err error
	if DifyPluginDB, err = pg.InitPluginDB("0.0.0.0", 5432, "testing", "postgres", "postgres", "difyai123456", "disable"); err != nil {
		t.Fatal(err)
	}
	defer Close()

	type TestTable struct {
		gorm.Model
	}

	err = CreateTable(&TestTable{})
	if err != nil {
		t.Fatal(err)
	}
	defer DropTable(&TestTable{})

	err = WithTransaction(func(tx *gorm.DB) error {
		data := TestTable{}
		err := tx.Create(&data).Error
		if err != nil {
			return err
		}

		return errors.New("rollback")
	})

	if err == nil {
		t.Fatal("expected error")
	} else if err.Error() != "rollback" {
		t.Fatal("unexpected error")
	}

	count, err := GetCount[TestTable]()
	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Fatal("unexpected count")
	}
}

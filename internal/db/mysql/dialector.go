package mysql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type myDialector struct {
	*mysql.Dialector
}

func (dialector myDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return myMigrator{dialector.Dialector.Migrator(db).(mysql.Migrator)}
}

func (dialector myDialector) DataTypeOf(field *schema.Field) string {
	dataType := dialector.Dialector.DataTypeOf(field)
	switch dataType {
	case "uuid":
		return "char(36)"
	case "text":
		return "longtext"
	default:
		return dataType
	}
}

type myMigrator struct {
	mysql.Migrator
}

func (migrator myMigrator) FullDataTypeOf(field *schema.Field) clause.Expr {
	if field.DataType == "uuid" {
		field.DataType = "char(36)"
		if field.HasDefaultValue && field.DefaultValue == "uuid_generate_v4()" {
			field.HasDefaultValue = false
			field.DefaultValue = ""
		}
	} else if field.DataType == "text" {
		field.DataType = "longtext"
	}
	return migrator.Migrator.FullDataTypeOf(field)
}

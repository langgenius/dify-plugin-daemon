package db

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

/*
	ORM for pgsql
*/

var DifyPluginDB *gorm.DB

var (
	ErrDatabaseNotFound = gorm.ErrRecordNotFound
)

func Create(data any, ctx ...*gorm.DB) error {
	if len(ctx) > 0 {
		return ctx[0].Create(data).Error
	}
	return DifyPluginDB.Create(data).Error
}

func Update(data any, ctx ...*gorm.DB) error {
	if len(ctx) > 0 {
		return ctx[0].Save(data).Error
	}
	return DifyPluginDB.Save(data).Error
}

func Delete(data any, ctx ...*gorm.DB) error {
	if len(ctx) > 0 {
		return ctx[0].Delete(data).Error
	}
	return DifyPluginDB.Delete(data).Error
}

func DeleteByCondition[T any](condition T, ctx ...*gorm.DB) error {
	var model T
	if len(ctx) > 0 {
		return ctx[0].Where(condition).Delete(&model).Error
	}
	return DifyPluginDB.Where(condition).Delete(&model).Error
}

func ReplaceAssociation[T any, R any](source *T, field string, associations []R, ctx ...*gorm.DB) error {
	if len(ctx) > 0 {
		return ctx[0].Model(source).Association(field).Replace(associations)
	}
	return DifyPluginDB.Model(source).Association(field).Replace(associations)
}

func AppendAssociation[T any, R any](source *T, field string, associations R, ctx ...*gorm.DB) error {
	if len(ctx) > 0 {
		return ctx[0].Model(source).Association(field).Append(associations)
	}
	return DifyPluginDB.Model(source).Association(field).Append(associations)
}

type genericComparableConstraint interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64 |
		bool
}

type genericEqualConstraint interface {
	genericComparableConstraint | string
}

type GenericQuery func(tx *gorm.DB) *gorm.DB

func Equal[T genericEqualConstraint](field string, value T) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s = ?", field), value)
	}
}

func EqualOr[T genericEqualConstraint](field string, value T) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Or(fmt.Sprintf("%s = ?", field), value)
	}
}

func NotEqual[T genericEqualConstraint](field string, value T) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s <> ?", field), value)
	}
}

func GreaterThan[T genericComparableConstraint](field string, value T) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s > ?", field), value)
	}
}

func GreaterThanOrEqual[T genericComparableConstraint](field string, value T) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s >= ?", field), value)
	}
}

func LessThan[T genericComparableConstraint](field string, value T) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s < ?", field), value)
	}
}

func LessThanOrEqual[T genericComparableConstraint](field string, value T) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s <= ?", field), value)
	}
}

func Like(field string, value string) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value+"%")
	}
}

func Page(page int, pageSize int) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Offset((page - 1) * pageSize).Limit(pageSize)
	}
}

func OrderBy(field string, desc bool) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		if desc {
			return tx.Order(fmt.Sprintf("%s DESC", field))
		}
		return tx.Order(field)
	}
}

// bitwise operation
func WithBit[T genericComparableConstraint](field string, value T) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s & ? = ?", field), value, value)
	}
}

func WithoutBit[T genericComparableConstraint](field string, value T) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s & ~? != 0", field), value)
	}
}

func Inc[T genericComparableConstraint](updates map[string]T) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		m := make(map[string]any)
		for field, value := range updates {
			m[field] = gorm.Expr(fmt.Sprintf("%s + ?", field), value)
		}
		return tx.UpdateColumns(m)
	}
}

func Dec[T genericComparableConstraint](updates map[string]T) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		m := make(map[string]any)
		for field, value := range updates {
			m[field] = gorm.Expr(fmt.Sprintf("%s - ?", field), value)
		}
		return tx.UpdateColumns(m)
	}
}

func Model(model any) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Model(model)
	}
}

func Fields(fields ...string) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Select(fields)
	}
}

func Preload(model string, args ...interface{}) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Preload(model, args...)
	}
}

func Join(field string) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Joins(field)
	}
}

func WLock /* write lock */ () GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Clauses(clause.Locking{Strength: "UPDATE"})
	}
}

func Where[T any](model *T) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(model)
	}
}

func WhereSQL(sql string, args ...interface{}) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(sql, args...)
	}
}

func Action(fn func(tx *gorm.DB)) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		fn(tx)
		return tx
	}
}

/*
Should be used first in query chain
*/
func WithTransactionContext(tx *gorm.DB) GenericQuery {
	return func(_ *gorm.DB) *gorm.DB {
		return tx
	}
}

func InArray(field string, value []interface{}) GenericQuery {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(fmt.Sprintf("%s IN ?", field), value)
	}
}

func Run(query ...GenericQuery) error {
	tmp := DifyPluginDB
	for _, q := range query {
		tmp = q(tmp)
	}

	// execute query
	return tmp.Error
}

func GetAny[T any](sql string, data ...interface{}) (T /* data */, error) {
	var result T
	err := DifyPluginDB.Raw(sql, data...).Scan(&result).Error
	return result, err
}

func GetOne[T any](query ...GenericQuery) (T /* data */, error) {
	var data T
	tmp := DifyPluginDB
	for _, q := range query {
		tmp = q(tmp)
	}
	err := tmp.First(&data).Error
	return data, err
}

func GetAll[T any](query ...GenericQuery) ([]T /* data */, error) {
	var data []T
	tmp := DifyPluginDB
	for _, q := range query {
		tmp = q(tmp)
	}
	err := tmp.Find(&data).Error
	return data, err
}

func GetCount[T any](query ...GenericQuery) (int64 /* count */, error) {
	var model T
	var count int64
	tmp := DifyPluginDB
	for _, q := range query {
		tmp = q(tmp)
	}
	err := tmp.Model(&model).Count(&count).Error
	return count, err
}

func GetSum[T any, R genericComparableConstraint](fields string, query ...GenericQuery) (R, error) {
	var model T
	var sum R
	tmp := DifyPluginDB
	for _, q := range query {
		tmp = q(tmp)
	}
	err := tmp.Model(&model).Select(fmt.Sprintf("SUM(%s)", fields)).Scan(&sum).Error
	return sum, err
}

func DelAssociation[T any](field string, query ...GenericQuery) error {
	var model T
	tmp := DifyPluginDB.Model(&model)
	for _, q := range query {
		tmp = q(tmp)
	}
	return tmp.Association(field).Unscoped().Clear()
}

func WithTransaction(fn func(tx *gorm.DB) error, ctx ...*gorm.DB) error {
	// Start a transaction
	db := DifyPluginDB
	if len(ctx) > 0 {
		db = ctx[0]
	}

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	err := fn(tx)
	if err != nil {
		if err := tx.Rollback().Error; err != nil {
			log.Error("failed to rollback tx: %v", err)
		}
		return err
	}
	tx.Commit()
	return nil
}

// NOTE: not used in production, only for testing
func DropTable(model any) error {
	return DifyPluginDB.Migrator().DropTable(model)
}

// NOTE: not used in production, only for testing
func CreateDatabase(dbname string) error {
	return DifyPluginDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname)).Error
}

// NOTE: not used in production, only for testing
func CreateTable(model any) error {
	return DifyPluginDB.Migrator().CreateTable(model)
}

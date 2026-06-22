package db

import (
	"errors"
	"reflect"
	"strings"

	sqlite "github.com/glebarez/go-sqlite"
	"gorm.io/gorm"
)

func (db *DB) GetWhereClauseFromParamTag(params any) map[string]any {
	whereClause := make(map[string]any)
	value := reflect.ValueOf(params)
	for i := 0; i < value.NumField(); i++ {
		paramTag := value.Type().Field(i).Tag.Get("param")
		valueField := value.Field(i).Interface()
		isNullableTag := value.Type().Field(i).Tag.Get("allow_zero_value") == "true"
		if paramTag == "" || (!isNullableTag && !db.checkIfNotNull(valueField)) {
			continue
		}

		whereClause[paramTag] = valueField
	}

	return whereClause
}

// IsRecordNotFound reports whether err is (or wraps) gorm.ErrRecordNotFound.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// asSQLiteError unwraps err into a *sqlite.Error if possible.
func asSQLiteError(err error) (*sqlite.Error, bool) {
	if err == nil {
		return nil, false
	}
	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) {
		return nil, false
	}
	return sqliteErr, true
}

// IsUniqueViolation reports whether err is a SQLite UNIQUE/PRIMARY KEY constraint violation.
// SQLite's message includes the offending "table.column", so if constraintOrColumn is non-empty
// it must appear (case-insensitive) in the message. Pass "" to match any unique violation.
func IsUniqueViolation(err error, constraintOrColumn string) bool {
	sqliteErr, ok := asSQLiteError(err)
	if !ok {
		return false
	}

	msg := strings.ToLower(sqliteErr.Error())
	if !strings.Contains(msg, "unique constraint failed") {
		return false
	}
	if constraintOrColumn == "" {
		return true
	}

	return strings.Contains(msg, strings.ToLower(constraintOrColumn))
}

// IsForeignKeyViolation reports whether err is a SQLite FOREIGN KEY constraint violation.
// Unlike Postgres, SQLite's FK error carries NO constraint or column name — so constraintOrColumn
// cannot be matched and is ignored. Callers that need to know which FK failed must disambiguate
// another way (e.g. only one FK on the insert).
func IsForeignKeyViolation(err error, _ string) bool {
	sqliteErr, ok := asSQLiteError(err)
	if !ok {
		return false
	}

	return strings.Contains(strings.ToLower(sqliteErr.Error()), "foreign key constraint failed")
}

func (db *DB) checkIfNotNull(value any) bool {
	switch value.(type) {
	case int64:
		if value.(int64) == int64(0) {
			return false
		}
	case string:
		if value.(string) == "" {
			return false
		}
	case bool:
		if value.(bool) == false {
			return false
		}
	case float64:
		if value.(float64) == float64(0) {
			return false
		}
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			return rv.Len() > 0
		}
	}

	return true
}

package db

import (
	"errors"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
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

// IsUniqueViolation reports whether err is a Postgres unique_violation (SQLSTATE 23505).
// If constraintOrColumn is non-empty it must appear (case-insensitive) in the constraint name,
// message, or detail of the violation. Pass "" to match any unique violation.
func IsUniqueViolation(err error, constraintOrColumn string) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "23505" {
		return false
	}
	if constraintOrColumn == "" {
		return true
	}

	needle := strings.ToLower(constraintOrColumn)
	return strings.Contains(strings.ToLower(pgErr.ConstraintName), needle) ||
		strings.Contains(strings.ToLower(pgErr.Message), needle) ||
		strings.Contains(strings.ToLower(pgErr.Detail), needle)
}

// IsForeignKeyViolation reports whether err is a Postgres foreign_key_violation (SQLSTATE 23503).
// If constraintOrColumn is non-empty it must appear (case-insensitive) in the constraint name,
// message, or detail of the violation. Pass "" to match any FK violation.
func IsForeignKeyViolation(err error, constraintOrColumn string) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "23503" {
		return false
	}
	if constraintOrColumn == "" {
		return true
	}

	needle := strings.ToLower(constraintOrColumn)
	return strings.Contains(strings.ToLower(pgErr.ConstraintName), needle) ||
		strings.Contains(strings.ToLower(pgErr.Message), needle) ||
		strings.Contains(strings.ToLower(pgErr.Detail), needle)
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

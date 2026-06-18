package types

import (
	"strings"

	"github.com/iancoleman/strcase"
)

// ToSnakeCase converts a display string into snake_case.
// Example: "Group Alpha 1" -> "group_alpha_1".
func ToSnakeCase(input string) string {
	return strcase.ToSnake(strings.TrimSpace(input))
}

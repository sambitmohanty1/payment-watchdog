package mediators

import "strings"

// escapeQuickBooksString escapes single quotes in strings for QuickBooks SQL
func escapeQuickBooksString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

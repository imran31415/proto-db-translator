package proto_db

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	userauth "github.com/imran31415/proto-db-translator/user"
	"github.com/stretchr/testify/require"
)

func TestGenerateSchema(t *testing.T) {
	// Create a sample User message
	user := &userauth.User{}

	// Generate schema
	schema, err := NewSqliteTranslator().GenerateSchema(user)
	require.NoError(t, err)

	// Check schema details
	if schema.TableName != "User" {
		t.Errorf("expected table name 'User', got '%s'", schema.TableName)
	}

	// Check column details (example for username)
	found := false
	for _, column := range schema.Columns {
		if column.Name == "username" {
			found = true
			if column.Type != "VARCHAR(255)" {
				t.Errorf("expected type 'VARCHAR(255)' for username, got '%s'", column.Type)
			}
			if !contains(column.Constraints, "NOT NULL") || !contains(column.Constraints, "UNIQUE") {
				t.Errorf("username column constraints mismatch")
			}
		}
	}
	if !found {
		t.Errorf("username column not found in schema")
	}
}

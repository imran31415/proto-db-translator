package proto_db

import (
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/imran31415/proto-db-translator/translator/db"
)

func TestGenerateMigration(t *testing.T) {
	// Mock old schema
	oldSchema := Schema{
		TableName: "User",
		Columns: []ColumnSchema{
			{Name: "id", Type: "INT", Constraints: []string{"NOT NULL", "PRIMARY KEY"}},
			{Name: "username", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL"}},
		},
	}

	// Mock new schema with changes
	newSchema := Schema{
		TableName: "User",
		Columns: []ColumnSchema{
			{Name: "id", Type: "INT", Constraints: []string{"NOT NULL", "PRIMARY KEY"}},
			{Name: "username", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL", "UNIQUE"}},
			{Name: "email", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL"}},
		},
	}

	// Generate migration
	migration := NewTranslator(db.DefaultMysqlConnection()).GenerateMigration(oldSchema, newSchema)

	// Validate migration SQL
	expected := `-- Migration for table: User
ALTER TABLE User MODIFY COLUMN username VARCHAR(255) NOT NULL UNIQUE;
ALTER TABLE User ADD COLUMN email VARCHAR(255) NOT NULL;
`

	// Normalize strings to trim whitespace and standardize newlines
	normalize := func(input string) string {
		lines := strings.Split(input, "\n")
		var trimmedLines []string
		for _, line := range lines {
			trimmedLines = append(trimmedLines, strings.TrimSpace(line))
		}
		return strings.Join(trimmedLines, "\n")
	}

	if normalize(migration) != normalize(expected) {
		t.Logf("Expected: [%q]", normalize(expected))
		t.Logf("Got:      [%q]", normalize(migration))
		t.Errorf("expected migration:\n%s\ngot:\n%s", normalize(expected), normalize(migration))
	}
}
func TestGenerateMigrationCases(t *testing.T) {
	tests := []struct {
		name      string
		oldSchema Schema
		newSchema Schema
		expected  string
	}{
		// Scenario 1: Adding a new column
		{
			name: "Add New Column",
			oldSchema: Schema{
				TableName: "User",
				Columns: []ColumnSchema{
					{Name: "id", Type: "INT", Constraints: []string{"NOT NULL", "PRIMARY KEY"}},
				},
			},
			newSchema: Schema{
				TableName: "User",
				Columns: []ColumnSchema{
					{Name: "id", Type: "INT", Constraints: []string{"NOT NULL", "PRIMARY KEY"}},
					{Name: "email", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL"}},
				},
			},
			expected: `-- Migration for table: User
ALTER TABLE User ADD COLUMN email VARCHAR(255) NOT NULL;
`,
		},
		// Scenario 2: Modifying an existing column's type
		{
			name: "Modify Column Type",
			oldSchema: Schema{
				TableName: "User",
				Columns: []ColumnSchema{
					{Name: "username", Type: "VARCHAR(100)", Constraints: []string{"NOT NULL"}},
				},
			},
			newSchema: Schema{
				TableName: "User",
				Columns: []ColumnSchema{
					{Name: "username", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL"}},
				},
			},
			expected: `-- Migration for table: User
ALTER TABLE User MODIFY COLUMN username VARCHAR(255) NOT NULL;
`,
		},
		// Scenario 3: Adding a constraint to an existing column
		{
			name: "Add Constraint to Column",
			oldSchema: Schema{
				TableName: "User",
				Columns: []ColumnSchema{
					{Name: "username", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL"}},
				},
			},
			newSchema: Schema{
				TableName: "User",
				Columns: []ColumnSchema{
					{Name: "username", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL", "UNIQUE"}},
				},
			},
			expected: `-- Migration for table: User
ALTER TABLE User MODIFY COLUMN username VARCHAR(255) NOT NULL UNIQUE;
`,
		},
		// Scenario 4: Removing a column
		{
			name: "Remove Column",
			oldSchema: Schema{
				TableName: "User",
				Columns: []ColumnSchema{
					{Name: "id", Type: "INT", Constraints: []string{"NOT NULL", "PRIMARY KEY"}},
					{Name: "email", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL"}},
				},
			},
			newSchema: Schema{
				TableName: "User",
				Columns: []ColumnSchema{
					{Name: "id", Type: "INT", Constraints: []string{"NOT NULL", "PRIMARY KEY"}},
				},
			},
			expected: `-- Migration for table: User
ALTER TABLE User DROP COLUMN email;
`,
		},
		// Scenario 5: Complex changes (add, modify, remove columns)
		{
			name: "Complex Changes",
			oldSchema: Schema{
				TableName: "User",
				Columns: []ColumnSchema{
					{Name: "id", Type: "INT", Constraints: []string{"NOT NULL", "PRIMARY KEY"}},
					{Name: "username", Type: "VARCHAR(100)", Constraints: []string{"NOT NULL"}},
				},
			},
			newSchema: Schema{
				TableName: "User",
				Columns: []ColumnSchema{
					{Name: "id", Type: "INT", Constraints: []string{"NOT NULL", "PRIMARY KEY"}},
					{Name: "username", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL", "UNIQUE"}},
					{Name: "email", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL"}},
				},
			},
			expected: `-- Migration for table: User
ALTER TABLE User MODIFY COLUMN username VARCHAR(255) NOT NULL UNIQUE;
ALTER TABLE User ADD COLUMN email VARCHAR(255) NOT NULL;
`,
		},
	}

	// Run all test cases
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			migration := NewTranslator(db.DefaultMysqlConnection()).GenerateMigration(test.oldSchema, test.newSchema)

			// Normalize strings for comparison
			normalize := func(input string) string {
				lines := strings.Split(input, "\n")
				var trimmedLines []string
				for _, line := range lines {
					trimmedLines = append(trimmedLines, strings.TrimSpace(line))
				}
				return strings.Join(trimmedLines, "\n")
			}

			if normalize(migration) != normalize(test.expected) {
				t.Logf("Expected:\n[%q]", normalize(test.expected))
				t.Logf("Got:\n[%q]", normalize(migration))
				t.Errorf("expected migration:\n%s\ngot:\n%s", normalize(test.expected), normalize(migration))
			}
		})
	}
}

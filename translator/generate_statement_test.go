package proto_db

import (
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	userauth "github.com/imran31415/proto-db-translator/user"
	"github.com/stretchr/testify/require"
)

func TestGenerateCreateTableSQL(t *testing.T) {
	tests := []struct {
		name     string
		schema   Schema
		expected string
	}{
		// Test Case 1: Basic Table with Primary Key
		{
			name: "Basic Table with Primary Key",
			schema: Schema{
				TableName: "User",
				Columns: []ColumnSchema{
					{Name: "id", Type: "INT", Constraints: []string{"NOT NULL"}, IsPrimaryKey: true},
					{Name: "username", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL"}},
				},
			},
			expected: `CREATE TABLE ` + "`User`" + ` (
  id INT NOT NULL PRIMARY KEY,
  username VARCHAR(255) NOT NULL
);`,
		},
		// Test Case 2: Table with Multiple Constraints
		{
			name: "Table with Multiple Constraints",
			schema: Schema{
				TableName: "User",
				Columns: []ColumnSchema{
					{Name: "id", Type: "INT", Constraints: []string{"NOT NULL"}, IsPrimaryKey: true},
					{Name: "username", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL", "UNIQUE"}},
					{Name: "email", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL"}},
				},
			},
			expected: `CREATE TABLE ` + "`User`" + ` (
  id INT NOT NULL PRIMARY KEY,
  username VARCHAR(255) NOT NULL UNIQUE,
  email VARCHAR(255) NOT NULL
);`,
		},
		// Test Case 3: Table with No Primary Key
		{
			name: "Table with No Primary Key",
			schema: Schema{
				TableName: "Logs",
				Columns: []ColumnSchema{
					{Name: "log_id", Type: "INT", Constraints: []string{"NOT NULL"}},
					{Name: "message", Type: "TEXT", Constraints: []string{}},
				},
			},
			expected: `CREATE TABLE ` + "`Logs`" + ` (
  log_id INT NOT NULL,
  message TEXT
);`,
		},
		// Test Case 4: Empty Table
		{
			name: "Empty Table",
			schema: Schema{
				TableName: "Empty",
				Columns:   []ColumnSchema{},
			},
			expected: `CREATE TABLE ` + "`Empty`" + ` (

			);`,
		},
		{
			name: "Composite Primary Keys",
			schema: Schema{
				TableName: "UserRoles",
				Columns: []ColumnSchema{
					{Name: "user_id", Type: "INT", Constraints: []string{"NOT NULL"}},
					{Name: "role_id", Type: "INT", Constraints: []string{"NOT NULL"}},
				},
				CompositePrimaryKeys: "user_id, role_id",
			},
			expected: `CREATE TABLE ` + "`UserRoles`" + ` (
		  user_id INT NOT NULL,
		  role_id INT NOT NULL,
		  PRIMARY KEY (user_id, role_id)
		);`,
		},
		{
			name: "Field with FULLTEXT Index",
			schema: Schema{
				TableName: "Articles",
				Columns: []ColumnSchema{
					{Name: "content", Type: "TEXT", Constraints: []string{"NOT NULL"}},
				},
				Indexes: []string{"FULLTEXT INDEX (content)"},
			},
			expected: `CREATE TABLE ` + "`Articles`" + ` (
		  content TEXT NOT NULL
		);
		CREATE FULLTEXT INDEX (content);`,
		},
		{
			name: "Composite Index",
			schema: Schema{
				TableName: "Orders",
				Columns: []ColumnSchema{
					{Name: "order_id", Type: "INT", Constraints: []string{"NOT NULL"}},
					{Name: "customer_id", Type: "INT", Constraints: []string{"NOT NULL"}},
				},
				Indexes: []string{"INDEX (order_id, customer_id)"},
			},
			expected: `CREATE TABLE ` + "`Orders`" + ` (
		  order_id INT NOT NULL,
		  customer_id INT NOT NULL
		);
		CREATE INDEX (order_id, customer_id);`,
		},
		{
			name: "Table with Foreign Keys",
			schema: Schema{
				TableName: "Orders",
				Columns: []ColumnSchema{
					{Name: "order_id", Type: "INT", Constraints: []string{"NOT NULL"}, IsPrimaryKey: true},
					{Name: "customer_id", Type: "INT", Constraints: []string{"NOT NULL"}, ForeignKeyTable: "Customers", ForeignKeyColumn: "id", OnDelete: "CASCADE", OnUpdate: "NO ACTION"},
				},
			},
			expected: `CREATE TABLE ` + "`Orders`" + ` (
		  order_id INT NOT NULL PRIMARY KEY,
		  customer_id INT NOT NULL,
		  FOREIGN KEY (customer_id) REFERENCES ` +
				"`Customers`" + ` (id) ON DELETE CASCADE ON UPDATE NO ACTION
		);`,
		},
	}

	// Add more updated test cases here as needed.

	// Run all test cases
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := NewTranslator(DefaultMysqlConnection()).GenerateCreateTableSQL(test.schema)

			// Normalize strings to trim whitespace and standardize newlines
			normalize := func(input string) string {
				lines := strings.Split(input, "\n")
				var trimmedLines []string
				for _, line := range lines {
					trimmedLines = append(trimmedLines, strings.TrimSpace(line))
				}
				return strings.Join(trimmedLines, "\n")
			}

			if normalize(actual) != normalize(test.expected) {
				t.Logf("Expected:\n[%q]", normalize(test.expected))
				t.Logf("Got:\n[%q]", normalize(actual))
				t.Errorf("expected SQL:\n%s\ngot:\n%s", normalize(test.expected), normalize(actual))
			}
		})
	}
}

func TestGenerateCreateTableSQLFromUserProto(t *testing.T) {
	// Define the expected SQL statement for the User table
	expected := `CREATE TABLE ` + "`User`" + ` (
  id INT NOT NULL PRIMARY KEY,
  username VARCHAR(255) NOT NULL UNIQUE,
  email VARCHAR(255) NOT NULL UNIQUE,
  hashed_password TEXT NOT NULL,
  is_2fa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  two_factor_secret VARCHAR(255),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);`

	// Validated by running through mysql:
	// mysql> describe user;
	// +-------------------+--------------+------+-----+-------------------+-----------------------------------------------+
	// | Field             | Type         | Null | Key | Default           | Extra                                         |
	// +-------------------+--------------+------+-----+-------------------+-----------------------------------------------+
	// | id                | int          | NO   | PRI | NULL              |                                               |
	// | username          | varchar(255) | NO   | UNI | NULL              |                                               |
	// | email             | varchar(255) | NO   | UNI | NULL              |                                               |
	// | hashed_password   | text         | NO   |     | NULL              |                                               |
	// | is_2fa_enabled    | tinyint(1)   | NO   |     | 0                 |                                               |
	// | two_factor_secret | varchar(255) | YES  |     | NULL              |                                               |
	// | created_at        | datetime     | NO   |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED                             |
	// | updated_at        | datetime     | NO   |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED on update CURRENT_TIMESTAMP |
	// +-------------------+--------------+------+-----+-------------------+-----------------------------------------------+

	// Generate the schema from the User proto
	user := &userauth.User{}
	schema, err := NewTranslator(DefaultMysqlConnection()).GenerateSchema(user)
	require.NoError(t, err)

	// Generate the CREATE TABLE SQL statement
	actual := NewTranslator(DefaultMysqlConnection()).GenerateCreateTableSQL(schema)

	// Normalize strings to trim whitespace and standardize newlines
	normalize := func(input string) string {
		lines := strings.Split(input, "\n")
		var trimmedLines []string
		for _, line := range lines {
			trimmedLines = append(trimmedLines, strings.TrimSpace(line))
		}
		return strings.Join(trimmedLines, "\n")
	}

	if normalize(actual) != normalize(expected) {
		t.Logf("Expected:\n[%q]", normalize(expected))
		t.Logf("Got:\n[%q]", normalize(actual))
		t.Errorf("expected SQL:\n%s\ngot:\n%s", normalize(expected), normalize(actual))
	}
}

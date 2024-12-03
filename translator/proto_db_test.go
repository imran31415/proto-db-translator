package proto_db

import (
	"fmt"
	"strings"
	"testing"

	userauth "github.com/imran31415/proto-db-translator/user"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
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
	migration := NewTranslator(DefaultMysqlConnection()).GenerateMigration(oldSchema, newSchema)

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
			migration := NewTranslator(DefaultMysqlConnection()).GenerateMigration(test.oldSchema, test.newSchema)

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
			expected: `CREATE TABLE User (
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
			expected: `CREATE TABLE User (
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
			expected: `CREATE TABLE Logs (
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
			expected: `CREATE TABLE Empty (

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
				CompositePrimaryKeys: []string{"user_id", "role_id"},
			},
			expected: `CREATE TABLE UserRoles (
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
			expected: `CREATE TABLE Articles (
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
			expected: `CREATE TABLE Orders (
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
			expected: `CREATE TABLE Orders (
		  order_id INT NOT NULL PRIMARY KEY,
		  customer_id INT NOT NULL,
		  FOREIGN KEY (customer_id) REFERENCES Customers (id) ON DELETE CASCADE ON UPDATE NO ACTION
		);`,
		},
		{
			name: "Table with AUTO_INCREMENT and DECIMAL precision/scale",
			schema: Schema{
				TableName: "Accounts",
				Columns: []ColumnSchema{
					{Name: "id", Type: "INT", Constraints: []string{"NOT NULL"}, IsPrimaryKey: true, AutoIncrement: true},
					{Name: "balance", Type: "DECIMAL", Constraints: []string{"NOT NULL"}, Precision: 10, Scale: 2},
				},
			},
			expected: `CREATE TABLE Accounts (
		  id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
		  balance DECIMAL(10,2) NOT NULL
		);`,
		},
		{
			name: "Table with Table-Level Constraints",
			schema: Schema{
				TableName: "UserRoles",
				Columns: []ColumnSchema{
					{Name: "user_id", Type: "INT", Constraints: []string{"NOT NULL"}},
					{Name: "role_id", Type: "INT", Constraints: []string{"NOT NULL"}},
				},
				UniqueConstraints: []string{"user_id, role_id"},
				CheckConstraints:  []string{"user_id > 0"},
			},
			expected: `CREATE TABLE UserRoles (
		  user_id INT NOT NULL,
		  role_id INT NOT NULL,
		  UNIQUE (user_id, role_id),
		  CHECK (user_id > 0)
		);`,
		},
		{
			name: "Table with Composite Indexes",
			schema: Schema{
				TableName: "Transactions",
				Columns: []ColumnSchema{
					{Name: "id", Type: "INT", Constraints: []string{"NOT NULL"}, IsPrimaryKey: true},
					{Name: "user_id", Type: "INT", Constraints: []string{"NOT NULL"}},
					{Name: "amount", Type: "DECIMAL", Constraints: []string{"NOT NULL"}, Precision: 10, Scale: 2},
				},
				Indexes: []string{"INDEX (user_id, amount)"},
			},
			expected: `CREATE TABLE Transactions (
			  id INT NOT NULL PRIMARY KEY,
			  user_id INT NOT NULL,
			  amount DECIMAL(10,2) NOT NULL
			);
			CREATE INDEX (user_id, amount);`,
		},
		{
			name: "Table with Multi-Column Unique Constraints",
			schema: Schema{
				TableName: "ProductPricing",
				Columns: []ColumnSchema{
					{Name: "product_id", Type: "INT", Constraints: []string{"NOT NULL"}},
					{Name: "region", Type: "VARCHAR(100)", Constraints: []string{"NOT NULL"}},
					{Name: "price", Type: "DECIMAL", Constraints: []string{"NOT NULL"}, Precision: 10, Scale: 2},
				},
				UniqueConstraints: []string{"product_id, region"},
			},
			expected: `CREATE TABLE ProductPricing (
			  product_id INT NOT NULL,
			  region VARCHAR(100) NOT NULL,
			  price DECIMAL(10,2) NOT NULL,
			  UNIQUE (product_id, region)
			);`,
		},
		{
			name: "Table with SPATIAL Index",
			schema: Schema{
				TableName: "GeoLocations",
				Columns: []ColumnSchema{
					{Name: "id", Type: "INT", Constraints: []string{"NOT NULL"}, IsPrimaryKey: true},
					{Name: "location", Type: "GEOMETRY", Constraints: []string{"NOT NULL"}},
				},
				Indexes: []string{"SPATIAL INDEX (location)"},
			},
			expected: `CREATE TABLE GeoLocations (
			  id INT NOT NULL PRIMARY KEY,
			  location GEOMETRY NOT NULL
			);
			CREATE SPATIAL INDEX (location);`,
		},
		{
			name: "Table with Nested Constraints",
			schema: Schema{
				TableName: "AuditLog",
				Columns: []ColumnSchema{
					{Name: "id", Type: "INT", Constraints: []string{"NOT NULL"}, IsPrimaryKey: true, AutoIncrement: true},
					{Name: "user_id", Type: "INT", Constraints: []string{"NOT NULL"}, ForeignKeyTable: "Users", ForeignKeyColumn: "id"},
					{Name: "action", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL", "CHECK (action IN ('INSERT', 'UPDATE', 'DELETE'))"}},
					{Name: "timestamp", Type: "DATETIME", Constraints: []string{"NOT NULL", "DEFAULT CURRENT_TIMESTAMP", "ON UPDATE CURRENT_TIMESTAMP"}},
				},
			},
			expected: `CREATE TABLE AuditLog (
			  id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			  user_id INT NOT NULL,
			  action VARCHAR(255) NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
			  timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			  FOREIGN KEY (user_id) REFERENCES Users (id)
			);`,
		},
		{
			name: "Table with Multiple Foreign Keys",
			schema: Schema{
				TableName: "OrderItems",
				Columns: []ColumnSchema{
					{
						Name:             "order_id",
						Type:             "INT",
						Constraints:      []string{"NOT NULL"},
						ForeignKeyTable:  "Orders",
						ForeignKeyColumn: "id",
						OnDelete:         "CASCADE",
						OnUpdate:         "NO ACTION",
					},
					{
						Name:             "product_id",
						Type:             "INT",
						Constraints:      []string{"NOT NULL"},
						ForeignKeyTable:  "Products",
						ForeignKeyColumn: "id",
						OnDelete:         "RESTRICT",
						OnUpdate:         "CASCADE",
					},
					{
						Name:        "quantity",
						Type:        "INT",
						Constraints: []string{"NOT NULL"},
					},
				},
				CompositePrimaryKeys: []string{"order_id", "product_id"},
			},
			expected: `CREATE TABLE OrderItems (
			  order_id INT NOT NULL,
			  product_id INT NOT NULL,
			  quantity INT NOT NULL,
			  PRIMARY KEY (order_id, product_id),
			  FOREIGN KEY (order_id) REFERENCES Orders (id) ON DELETE CASCADE ON UPDATE NO ACTION,
			  FOREIGN KEY (product_id) REFERENCES Products (id) ON DELETE RESTRICT ON UPDATE CASCADE
			);`,
		},
		{
			name: "Column with CHARACTER SET",
			schema: Schema{
				TableName: "Users",
				Columns: []ColumnSchema{
					{Name: "username", Type: "VARCHAR(255)", CharacterSet: "utf8mb4"},
				},
			},
			expected: `CREATE TABLE Users (
  username VARCHAR(255) CHARACTER SET utf8mb4
);`,
		},
		{
			name: "Column with COLLATE",
			schema: Schema{
				TableName: "Users",
				Columns: []ColumnSchema{
					{Name: "username", Type: "VARCHAR(255)", Collation: "utf8mb4_general_ci"},
				},
			},
			expected: `CREATE TABLE Users (
  username VARCHAR(255) COLLATE utf8mb4_general_ci
);`,
		},
		{
			name: "Column with CHARACTER SET and COLLATE",
			schema: Schema{
				TableName: "Users",
				Columns: []ColumnSchema{
					{Name: "username", Type: "VARCHAR(255)", CharacterSet: "utf8mb4", Collation: "utf8mb4_general_ci"},
				},
			},
			expected: `CREATE TABLE Users (
  username VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci
);`,
		},
		{
			name: "Column with UUID Default",
			schema: Schema{
				TableName: "Users",
				Columns: []ColumnSchema{
					{Name: "id", Type: "CHAR(36)", Constraints: []string{"NOT NULL"}, DefaultFunction: "UUID()"},
				},
			},
			expected: `CREATE TABLE Users (
  id CHAR(36) NOT NULL DEFAULT UUID()
);`,
		},
		{
			name: "Column with NOW Default",
			schema: Schema{
				TableName: "Events",
				Columns: []ColumnSchema{
					{Name: "created_at", Type: "DATETIME", Constraints: []string{"NOT NULL"}, DefaultFunction: "NOW()"},
				},
			},
			expected: `CREATE TABLE Events (
  created_at DATETIME NOT NULL DEFAULT NOW()
);`,
		},
		{
			name: "Column with Constraints and Default Function",
			schema: Schema{
				TableName: "Logs",
				Columns: []ColumnSchema{
					{Name: "log_id", Type: "INT", Constraints: []string{"NOT NULL", "AUTO_INCREMENT"}, DefaultFunction: ""},
					{Name: "timestamp", Type: "DATETIME", Constraints: []string{"NOT NULL"}, DefaultFunction: "NOW()"},
				},
			},
			expected: `CREATE TABLE Logs (
  log_id INT NOT NULL AUTO_INCREMENT,
  timestamp DATETIME NOT NULL DEFAULT NOW()
);`,
		},
		{
			name: "Message with Composite Primary Keys",
			schema: Schema{
				TableName: "CompositePrimaryKeyTable",
				Columns: []ColumnSchema{
					{Name: "key1", Type: "INT", Constraints: []string{"NOT NULL"}},
					{Name: "key2", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL"}},
				},
				CompositePrimaryKeys: []string{"key1", "key2"},
			},
			expected: `CREATE TABLE CompositePrimaryKeyTable (
		  key1 INT NOT NULL,
		  key2 VARCHAR(255) NOT NULL,
		  PRIMARY KEY (key1, key2)
		);`,
		},
		{
			name: "Message with Composite Unique Constraint",
			schema: Schema{
				TableName: "CompositeUniqueTable",
				Columns: []ColumnSchema{
					{Name: "field1", Type: "INT", Constraints: []string{"NOT NULL"}},
					{Name: "field2", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL"}},
				},
				UniqueConstraints: []string{"field1, field2"},
			},
			expected: `CREATE TABLE CompositeUniqueTable (
		  field1 INT NOT NULL,
		  field2 VARCHAR(255) NOT NULL,
		  UNIQUE (field1, field2)
		);`,
		},
		{
			name: "Message with Composite Index",
			schema: Schema{
				TableName: "CompositeIndexTable",
				Columns: []ColumnSchema{
					{Name: "field1", Type: "INT", Constraints: []string{"NOT NULL"}},
					{Name: "field2", Type: "VARCHAR(255)", Constraints: []string{"NOT NULL"}},
				},
				Indexes: []string{"INDEX (field1, field2)"},
			},
			expected: `CREATE TABLE CompositeIndexTable (
		  field1 INT NOT NULL,
		  field2 VARCHAR(255) NOT NULL
		);
		CREATE INDEX (field1, field2);`,
		},
		{
			name: "Message with Check Constraints",
			schema: Schema{
				TableName: "CheckConstraintTable",
				Columns: []ColumnSchema{
					{Name: "age", Type: "INT", Constraints: []string{"NOT NULL"}},
				},
				CheckConstraints: []string{"age > 18"},
			},
			expected: `CREATE TABLE CheckConstraintTable (
		  age INT NOT NULL,
		  CHECK (age > 18)
		);`,
		},
	}

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
	expected := `CREATE TABLE User (
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

func TestValidateSqlite(t *testing.T) {
	tests := []struct {
		name      string
		proto     proto.Message
		expectErr bool
	}{
		{
			name:      "Valid User Table",
			proto:     &userauth.User{},
			expectErr: false,
		},
		{
			name:      "Valid Role Table",
			proto:     &userauth.Role{},
			expectErr: false,
		},
		{
			name:      "Invalid Schema Example",
			proto:     &userauth.InvalidSqlSchema1{}, // Add an invalid schema case if needed
			expectErr: true,                          // Expect an error if schema is intentionally invalid
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Initialize translator
			translator := NewTranslator(DefaultSqliteConnection())

			// Validate schema
			err := translator.ValidateSchema(protoList(test.proto), "root:Password123!@tcp(localhost)/testt")

			if test.expectErr {
				require.Error(t, err)
				t.Logf("Expected error: %v", err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMysql(t *testing.T) {
	tests := []struct {
		name      string
		proto     proto.Message
		expectErr bool
	}{
		{
			name:      "Valid User Table",
			proto:     &userauth.User{},
			expectErr: false,
		},
		// {
		// 	name:      "Valid Role Table",
		// 	proto:     &userauth.Role{},
		// 	expectErr: false,
		// },
		{
			name:      "Invalid Schema Example",
			proto:     &userauth.InvalidSqlSchema1{}, // Add an invalid schema case if needed
			expectErr: true,                          // Expect an error if schema is intentionally invalid
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Initialize translator
			translator := NewTranslator(DefaultMysqlConnection())

			// Validate schema
			err := translator.ValidateSchema(protoList(test.proto), "root:Password123!@tcp(localhost)/testt")

			if test.expectErr {
				require.Error(t, err)
				t.Logf("Expected error: %v", err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
func TestInvalidSqlSchemaValidation(t *testing.T) {
	// TODO: the commented out test cases should fail but they dont because sqlite isnt very strict with the schema it allows.

	// Define test cases
	testCases := []struct {
		name             string
		dbType           DatabaseType
		schema           proto.Message
		connectionString string
		expectedErrors   []string
	}{
		{
			name:             "Invalid Sql Schema 1 - SQLite",
			dbType:           DatabaseTypeSQLite,
			schema:           &userauth.InvalidSqlSchema1{},
			connectionString: ":memory:",
			expectedErrors:   []string{"missing or invalid db_column annotation"},
		},
		{
			name:             "Invalid Sql Schema 1 - MYSQL",
			dbType:           DatabaseTypeMySQL,
			schema:           &userauth.InvalidSqlSchema1{},
			connectionString: "root:Password123!@tcp(localhost)/testt",
			expectedErrors:   []string{"missing or invalid db_column annotation"},
		},
		// {
		// 	name:           "Invalid Sql Schema 2 - SQLITE",
		// 	dbType:         DatabaseTypeSQLite,
		// 	schema:         &userauth.InvalidSqlSchema2{},
		// 	expectedErrors: []string{"Failed to open the referenced table 'NonExistentTable'"},
		// },
		{
			name:             "Invalid Sql Schema 2 - MySQL",
			dbType:           DatabaseTypeMySQL,
			schema:           &userauth.InvalidSqlSchema2{},
			connectionString: "root:Password123!@tcp(localhost)/testt",
			expectedErrors:   []string{"Failed to open the referenced table 'NonExistentTable'"},
		},
		// {
		// 	name:             "Invalid Sql Schema 4- Sqlite",
		// 	dbType:           DatabaseTypeMySQL,
		// 	schema:           &userauth.InvalidSqlSchema4{},
		// 	connectionString: ":memory",
		// 	expectedErrors:   []string{"You have an error in your SQL syntax;"},
		// },
		{
			name:             "Invalid Sql Schema 4 - MySQL",
			dbType:           DatabaseTypeMySQL,
			schema:           &userauth.InvalidSqlSchema4{},
			connectionString: "root:Password123!@tcp(localhost)/testt",
			expectedErrors:   []string{"You have an error in your SQL syntax"},
		},
		{
			name:             "Invalid Sql Schema 5 - MySQL",
			dbType:           DatabaseTypeMySQL,
			schema:           &userauth.InvalidSqlSchema5{},
			connectionString: "root:Password123!@tcp(localhost)/testt",
			expectedErrors:   []string{"Unknown character set: 'unsupported_charset"},
		},
		// {
		// 	name:             "Invalid Sql Schema 5 - Sqlit",
		// 	dbType:           DatabaseTypeMySQL,
		// 	schema:           &userauth.InvalidSqlSchema5{},
		// 	connectionString: ":memory",
		// 	expectedErrors:   []string{"You have an error in your SQL syntax"},
		// },
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var translator Translator
			switch tc.dbType {
			case DatabaseTypeMySQL:
				translator = NewTranslator(DefaultMysqlConnection())

			case DatabaseTypeSQLite:
				translator = NewTranslator(DefaultSqliteConnection())

			}

			// Validate schema
			err := translator.ValidateSchema(protoList(tc.schema), tc.connectionString)
			require.Error(t, err, "Validation should fail for invalid schema")

			// Assert the error message includes known issues
			for _, expected := range tc.expectedErrors {
				require.Contains(t, err.Error(), expected, "Error message should contain: "+expected)
			}
		})
	}
}
func TestProcessProtoMessages(t *testing.T) {
	translator := NewTranslator(DefaultMysqlConnection())

	// Protobuf messages to process
	protoMessages := []proto.Message{
		&userauth.User{}, // Replace with your actual proto message types
		&userauth.Role{},
		&userauth.RoleHierarchy{},
	}

	// Directory for generated models
	outputDir := "../generated_models"

	// Call the refactored method
	err := translator.ProcessProtoMessages(outputDir, protoMessages)
	if err != nil {
		fmt.Println("err is", err)
	}
	require.NoError(t, err, "ProcessProtoMessages failed")
}

package proto_db

import (
	"testing"

	userauth "github.com/imran31415/proto-db-translator/user"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

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
			translator := NewTranslator(DefaultMysqlConnection())

			// Validate schema
			err := translator.ValidateSchema(protoList(test.proto), "root:Password123!@tcp(localhost)/")

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
			connectionString: "root:Password123!@tcp(localhost)/",
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
			connectionString: "root:Password123!@tcp(localhost)/",
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
			connectionString: "root:Password123!@tcp(localhost)/",
			expectedErrors:   []string{"You have an error in your SQL syntax"},
		},
		{
			name:             "Invalid Sql Schema 5 - MySQL",
			dbType:           DatabaseTypeMySQL,
			schema:           &userauth.InvalidSqlSchema5{},
			connectionString: "root:Password123!@tcp(localhost)/",
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

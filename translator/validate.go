package proto_db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/mattn/go-sqlite3"

	"google.golang.org/protobuf/proto"
)

func protoList(p proto.Message) []proto.Message {
	return []proto.Message{p}
}

// ValidateSchema validates the schema by applying it to a test database
func (t Translator) ValidateSchema(protoMessages []proto.Message, dsn string) error {
	var db *sql.DB
	var openErr error

	switch t.dbConnection.DbType {
	case DatabaseTypeSQLite:
		// Open an in-memory SQLite database
		db, openErr = sql.Open("sqlite3", ":memory:")
		if openErr != nil {
			return fmt.Errorf("failed to connect to SQLite database: %w", openErr)
		}
		defer db.Close()

		// Enable foreign key constraints for SQLite
		_, err := db.Exec("PRAGMA foreign_keys = ON;")
		if err != nil {
			return fmt.Errorf("failed to enable foreign key constraints: %w", err)
		}

	case DatabaseTypeMySQL:
		// Connect to MySQL
		db, openErr = sql.Open("mysql", dsn)
		if openErr != nil {
			return fmt.Errorf("failed to connect to MySQL database: %w", openErr)
		}
		defer db.Close()

		// Create a temporary database for validation
		tempDB := "test_validate"
		_, err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", tempDB))
		if err != nil {
			return fmt.Errorf("failed to create temporary database: %w", err)
		}

		// Switch to the temporary database
		_, err = db.Exec(fmt.Sprintf("USE %s", tempDB))
		if err != nil {
			return fmt.Errorf("failed to switch to temporary database: %w", err)
		}

		// Ensure the temporary database is dropped after validation
		defer func() {
			_, _ = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", tempDB))
		}()
	}

	// Process each proto message
	for _, protoMessage := range protoMessages {
		// Extract table name from the Protobuf message
		tableName := string(protoMessage.ProtoReflect().Descriptor().Name())

		// Generate the schema from the proto message
		schema, err := t.generateSchema(protoMessage)
		if err != nil {
			return fmt.Errorf("failed to generate schema for table '%s': %w", tableName, err)
		}

		// Generate the CREATE TABLE statement
		createTableSQL := t.GenerateCreateTableSQL(schema)
		fmt.Printf("Generated SQL for validation (table '%s'):\n%s\n", tableName, createTableSQL)

		// _, err = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", tableName))
		// if err != nil {
		// 	return fmt.Errorf("failed to drop table '%s': %w", tableName, err)
		// }
		// Apply the CREATE TABLE statement
		_, err = db.Exec(createTableSQL)
		if err != nil {
			return fmt.Errorf("schema validation failed for table '%s': %w\nSQL: %s", tableName, err, createTableSQL)
		}
	}

	return nil
}

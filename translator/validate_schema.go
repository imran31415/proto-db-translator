package proto_db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/mattn/go-sqlite3"

	"github.com/google/uuid"
	"github.com/imran31415/proto-db-translator/translator/db"

	"google.golang.org/protobuf/proto"
)

func protoList(p proto.Message) []proto.Message {
	return []proto.Message{p}
}

// ValidateSchema validates the schema by applying it to a test database
func (t Translator) ValidateSchema(protoMessages []proto.Message, dsn string) error {
	var database *sql.DB
	var openErr error

	switch t.dbConnection.DbType {
	case db.DatabaseTypeSQLite:
		// Open an in-memory SQLite database
		database, openErr = sql.Open("sqlite3", ":memory:")
		if openErr != nil {
			return fmt.Errorf("failed to connect to SQLite database: %w", openErr)
		}
		defer database.Close()

		// Enable foreign key constraints for SQLite
		_, err := database.Exec("PRAGMA foreign_keys = ON;")
		if err != nil {
			return fmt.Errorf("failed to enable foreign key constraints: %w", err)
		}

	case db.DatabaseTypeMySQL:
		// Connect to MySQL
		database, openErr = sql.Open("mysql", dsn+"?multiStatements=true")
		if openErr != nil {
			return fmt.Errorf("failed to connect to MySQL database: %w", openErr)
		}

		// Create a temporary database for validation
		tempDB := "tempdb" + strings.Replace(uuid.NewString(), "-", "", 10)
		// ensure clean validation
		_, err := database.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", tempDB))
		if err != nil {
			return fmt.Errorf("failed to create temporary database: %w", err)
		}

		// Ensure the temporary database is dropped after validation
		_, err = database.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", tempDB))
		if err != nil {
			return fmt.Errorf("failed to create temporary database: %w", err)
		}
		defer func() {
			database.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", tempDB))

		}()

		// Switch to the temporary database
		_, err = database.Exec(fmt.Sprintf("USE %s", tempDB))
		if err != nil {
			return fmt.Errorf("failed to switch to temporary database: %w", err)
		}

	}

	// Process each proto message
	for _, protoMessage := range protoMessages {
		tableName := string(protoMessage.ProtoReflect().Descriptor().Name())
		schema, err := t.GenerateSchema(protoMessage)
		if err != nil {
			return fmt.Errorf("failed to generate schema for table '%s': %w", tableName, err)
		}

		createTableSQL := t.GenerateCreateTableSQL(schema)
		// fmt.Printf("Generated SQL for validation (table '%s'):\n%s\n", tableName, createTableSQL)

		_, err = database.Exec(createTableSQL)
		if err != nil {
			return fmt.Errorf("schema validation failed for table: %s.  err: %s\nSQL: %s", tableName, err, createTableSQL)
		}
		log.Printf("Successfully validated %s", tableName)
	}

	return nil
}

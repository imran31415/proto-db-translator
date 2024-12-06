package proto_db

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"

	"google.golang.org/protobuf/proto"
)

func (t Translator) GenerateModels(outputDir string, protoMessages []proto.Message) error {
	// Database connection string (without specifying a database)
	dbConnStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/", t.dbConnection.DbUser, t.dbConnection.DbPass, t.dbConnection.DbHost, t.dbConnection.DbPort)

	// Step 1: Connect to MySQL
	db, err := sql.Open("mysql", dbConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	defer func() {
		_, dropErr := db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", t.dbConnection.DbName))
		if dropErr != nil {
			fmt.Printf("Failed to drop test database: %v\n", dropErr)
		}
		db.Close()
	}()

	// Step 2: Create and switch to the test database
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", t.dbConnection.DbName))
	if err != nil {
		return fmt.Errorf("failed to create test database: %w", err)
	}

	dbConnStr = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", t.dbConnection.DbUser, t.dbConnection.DbPass, t.dbConnection.DbHost, t.dbConnection.DbPort, t.dbConnection.DbName)
	// Step 1: Connect to MySQL
	db, err = sql.Open("mysql", dbConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// Validate schema
	err = t.ValidateSchema(protoMessages, dbConnStr)
	if err != nil {
		return fmt.Errorf("schema validation failed for tables '%v': %w", protoMessages, err)
	}

	// Step 3: Process each proto message
	for _, protoMessage := range protoMessages {
		// Extract table name from the Protobuf message
		tableName := string(protoMessage.ProtoReflect().Descriptor().Name())

		// Generate schema from the proto message
		schema, err := t.GenerateSchema(protoMessage)
		if err != nil {
			return fmt.Errorf("failed to generate schema for table '%s': %w", tableName, err)
		}

		// Generate the CREATE TABLE SQL statement
		createTableSQL := t.GenerateCreateTableSQL(schema)

		// fmt.Printf("Generated SQL for table '%s':\n%s\n", tableName, createTableSQL)

		// Ensure the table is dropped if it exists to avoid conflicts
		_, err = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", tableName))
		if err != nil {
			fmt.Printf("failed to drop table '%s': %s", tableName, err)
		}

		// Apply the schema to the database
		_, err = db.Exec(createTableSQL)
		if err != nil {
			return fmt.Errorf("failed to apply schema to table '%s': %w\nSQL: %s", tableName, err, createTableSQL)
		}

		// Update the connection string to include the database name
		fullDBConnStr := fmt.Sprintf("mysql://%s:%s@%s:%s/%s", t.dbConnection.DbUser, t.dbConnection.DbPass, t.dbConnection.DbHost, t.dbConnection.DbPort, t.dbConnection.DbName)

		// Generate XO models
		err = t.runXo(fullDBConnStr, tableName, outputDir, "../templates")
		if err != nil {
			return fmt.Errorf("model generation failed for table '%s': %w", tableName, err)
		}

		// fmt.Printf("Model generated successfully for table '%s' in directory '%s'\n", tableName, outputDir)
	}

	return nil
}

func (t Translator) runXo(dbConnStr, tableName, outputDir, templatesDir string) error {
	// Check if XO is installed
	_, err := exec.LookPath("xo")
	if err != nil {
		return fmt.Errorf("xo is not installed: %w", err)
	}

	// Build the XO command
	cmdArgs := []string{
		"schema",           // XO subcommand for schema inspection
		dbConnStr,          // Database connection string as positional argument
		"--out", outputDir, // Output directory for generated models
		"--include", tableName, // Include only the specified table
		"--src", templatesDir, // Specify the templates directory
	}

	// Prepare XO command
	cmd := exec.Command("xo", cmdArgs...)
	cmd.Stdout = os.Stdout // Forward XO's stdout for visibility
	cmd.Stderr = os.Stderr // Forward XO's stderr for error visibility

	// Run the XO command
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate model with XO: %w", err)
	}

	// Run the find and sed command to update generated files
	findCmd := exec.Command("find", outputDir, "-type", "f", "-name", "*.go", "-exec", "sed", "-i", "", "s/proto_db_default\\.//g", "{}", "+")
	findCmd.Stdout = os.Stdout // Forward find command's stdout for visibility
	findCmd.Stderr = os.Stderr // Forward find command's stderr for error visibility

	err = findCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to process generated files with find/sed: %w", err)
	}

	fmt.Printf("Model generated and processed successfully for table '%s' in directory '%s'\n", tableName, outputDir)
	return nil
}

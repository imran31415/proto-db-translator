package proto_db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/protobuf/proto"
)

func (t Translator) GenerateModels(outputDir string, protoMessages []proto.Message) error {
	// Database connection string (without specifying a database)
	dbConnStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/", t.dbConnection.DbUser, t.dbConnection.DbPass, t.dbConnection.DbHost, t.dbConnection.DbPort)
	dbConnStr += "?multiStatements=true"
	// Step 1: Connect to MySQL
	db, err := sql.Open("mysql", dbConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	defer db.Close()
	_, dropErr := db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", t.dbConnection.DbName))
	if dropErr != nil {
		fmt.Printf("Failed to drop test database: %v\n", dropErr)
	}

	// Step 2: Create and switch to the test database
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", t.dbConnection.DbName))
	if err != nil {
		return fmt.Errorf("failed to create test database: %w", err)
	}

	dbConnStr = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", t.dbConnection.DbUser, t.dbConnection.DbPass, t.dbConnection.DbHost, t.dbConnection.DbPort, t.dbConnection.DbName)
	// Step 1: Connect to MySQL
	dbConnStr += "?multiStatements=true"
	db, err = sql.Open("mysql", dbConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// Validate schema
	statements, err := t.ValidateSchema(protoMessages)
	if err != nil {
		return fmt.Errorf("schema validation failed for tables '%v': %w", protoMessages, err)
	}

	execute := ""
	for x, statement := range statements {
		if x == 0 {
			execute += statement.Statement
		} else {
			// execute := "\n"
			execute += statement.Statement
		}
	}

	// Switch to the temporary database
	_, err = db.Exec(fmt.Sprintf("USE %s", t.dbConnection.DbName))
	if err != nil {
		return fmt.Errorf("failed to switch to temporary database: %w", err)
	}

	// Step 3: Process xo
	_, err = db.Exec(strings.TrimSpace(execute))
	if err != nil {
		return fmt.Errorf("failed to apply schema err: %s. \nSQL: %s", err, execute)
	}
	log.Printf("Successfully executed statement on DB: %s ", t.dbConnection.DbName)

	// Update the connection string to include the database name
	fullDBConnStr := fmt.Sprintf("mysql://%s:%s@%s:%s/%s", t.dbConnection.DbUser, t.dbConnection.DbPass, t.dbConnection.DbHost, t.dbConnection.DbPort, t.dbConnection.DbName)

	// Generate XO models
	err = t.runXo(fullDBConnStr, outputDir, "../templates")
	if err != nil {
		return fmt.Errorf("model generation failed for dir '%s': %s", outputDir, err)
	}
	return nil
}

func (t Translator) runXo(dbConnStr, outputDir, templatesDir string) error {
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
		"--src", templatesDir, // Specify the templates directory
	}

	// Prepare XO command
	cmd := exec.Command("xo", cmdArgs...)
	cmd.Stdout = os.Stdout // Forward XO's stdout for visibility
	cmd.Stderr = os.Stderr // Forward XO's stderr for error visibility

	// Run the XO command
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate model with XO: %s", err)
	}

	// Run the find and sed command to update generated files to remove the DB name from sql statements
	findCmd := exec.Command("find", outputDir, "-type", "f", "-name", "*.go", "-exec", "sed", "-i", "", fmt.Sprintf("s/%s\\.//g", t.dbConnection.DbName), "{}", "+")
	findCmd.Stdout = os.Stdout // Forward find command's stdout for visibility
	findCmd.Stderr = os.Stderr // Forward find command's stderr for error visibility

	err = findCmd.Run()
	if err != nil {
		log.Printf("failed to process generated files with find/sed: %s", err)
	}

	fmt.Printf("Model generated and processed successfully in directory '%s'\n", outputDir)
	return nil
}

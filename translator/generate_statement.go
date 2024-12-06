package proto_db

import (
	"fmt"
	"strings"
)

func (t Translator) GenerateCreateTableSQL(schema Schema) string {
	var createStmt strings.Builder
	createStmt.WriteString(fmt.Sprintf("CREATE TABLE `%s` (\n", schema.TableName))

	// Add column definitions
	for i, col := range schema.Columns {

		if t.dbConnection.DbType == DatabaseTypeSQLite {
			// Remove CHARACTER SET and COLLATE for SQLite
			if col.CharacterSet != "" || col.Collation != "" {
				col.CharacterSet = ""
				col.Collation = ""
			}
		}
		createStmt.WriteString(fmt.Sprintf("  %s %s", col.Name, col.Type))

		// Add precision for DECIMAL
		if col.Type == "DECIMAL" && (col.Precision > 0 || col.Scale > 0) {
			createStmt.WriteString(fmt.Sprintf("(%d,%d)", col.Precision, col.Scale))
		}

		// Add character set and collation
		if col.CharacterSet != "" {
			createStmt.WriteString(fmt.Sprintf(" CHARACTER SET %s", col.CharacterSet))
		}

		// Add constraints
		if len(col.Constraints) > 0 {
			createStmt.WriteString(fmt.Sprintf(" %s", joinConstraints(col.Constraints)))
		}

		if col.Collation != "" {
			createStmt.WriteString(fmt.Sprintf(" COLLATE %s", col.Collation))
		}

		// Handle AutoIncrement
		if col.AutoIncrement {
			if t.dbConnection.DbType == DatabaseTypeSQLite {
				createStmt.WriteString(" PRIMARY KEY,\n") // SQLite-specific
				continue
			} else {
				createStmt.WriteString(" AUTO_INCREMENT PRIMARY KEY") // MySQL/PostgreSQL
			}
		}

		// Add default function
		if col.DefaultFunction != "" {
			createStmt.WriteString(fmt.Sprintf(" DEFAULT %s", col.DefaultFunction))
		}

		// Add PRIMARY KEY directly to column if applicable
		if col.IsPrimaryKey && len(schema.CompositePrimaryKeys) == 0 && !col.AutoIncrement {
			createStmt.WriteString(" PRIMARY KEY")
		}

		// Add a comma if it's not the last column
		if i < len(schema.Columns)-1 {
			createStmt.WriteString(",\n")
		}
	}

	// Add composite primary keys as a table-level constraint
	if len(schema.CompositePrimaryKeys) > 0 {
		createStmt.WriteString(fmt.Sprintf(",\n  PRIMARY KEY (%s)", schema.CompositePrimaryKeys))
	}

	// Add unique constraints
	for _, unique := range schema.UniqueConstraints {
		createStmt.WriteString(fmt.Sprintf(",\n  UNIQUE (%s)", unique))
	}

	// Generate CREATE INDEX statements for composite indexes
	for _, compositeIndex := range schema.CompositeIndexes {
		// Split the composite index definition into individual column names
		indexColumns := strings.Split(compositeIndex, ",")
		// Generate a unique index name based on the column names
		indexName := fmt.Sprintf("%s_%s_idx", schema.TableName, strings.Join(indexColumns, "_"))
		// Create the CREATE INDEX statement
		createStmt.WriteString(fmt.Sprintf(",\n  UNIQUE KEY `%s` (%s)", indexName, strings.Join(indexColumns, ", ")))
	}
	// CHECK (quantity > 0 AND price_per_unit >= 0),

	if len(schema.CheckConstraints) > 0 {
		checkConstraint := ""
		for i, x := range schema.CheckConstraints {
			if i == 0 {
				checkConstraint += x
			} else {
				checkConstraint += " AND "
				checkConstraint += x
			}
		}

		createStmt.WriteString(fmt.Sprintf(",\n  CHECK (%s)", checkConstraint))
	}
	// Add foreign key constraints as table-level constraints
	for _, col := range schema.Columns {
		if col.ForeignKeyTable != "" && col.ForeignKeyColumn != "" {
			createStmt.WriteString(fmt.Sprintf(",\n  FOREIGN KEY (%s) REFERENCES `%s` (%s)", col.Name, col.ForeignKeyTable, col.ForeignKeyColumn))
			if col.OnDelete != "" {
				createStmt.WriteString(fmt.Sprintf(" ON DELETE %s", col.OnDelete))
			}
			if col.OnUpdate != "" && t.dbConnection.DbType != DatabaseTypeSQLite {

				createStmt.WriteString(fmt.Sprintf(" ON UPDATE %s", col.OnUpdate))
			}
		}
	}

	createStmt.WriteString("\n);")
	for _, index := range schema.Indexes {
		createStmt.WriteString(fmt.Sprintf("\nCREATE %s;", index))

	}

	// Add composite index definitions

	// log.Printf("----Table: %s, Statement: %s", schema.TableName, createStmt.String())
	return createStmt.String()
}

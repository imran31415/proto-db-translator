package proto_db

import (
	"fmt"
	"strings"
)

// Compare schemas and generate migration SQL
func (t Translator) GenerateMigration(oldSchema, newSchema Schema) string {
	var migration strings.Builder
	migration.WriteString(fmt.Sprintf("-- Migration for table: %s\n", newSchema.TableName))

	// Build a map of existing columns for quick comparison
	oldColumns := make(map[string]ColumnSchema)
	for _, col := range oldSchema.Columns {
		oldColumns[col.Name] = col
	}

	// Handle new or modified columns
	for _, newCol := range newSchema.Columns {
		if oldCol, exists := oldColumns[newCol.Name]; exists {
			// Check if the column has changed
			if oldCol.Type != newCol.Type || !equalConstraints(oldCol.Constraints, newCol.Constraints) {
				migration.WriteString(fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s %s;\n",
					newSchema.TableName, newCol.Name, newCol.Type, joinConstraints(newCol.Constraints)))
			}
		} else {
			// Column is new
			migration.WriteString(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s %s;\n",
				newSchema.TableName, newCol.Name, newCol.Type, joinConstraints(newCol.Constraints)))
		}
	}

	// Handle removed columns
	for _, oldCol := range oldSchema.Columns {
		if _, exists := findColumnInSchema(oldCol.Name, newSchema.Columns); !exists {
			migration.WriteString(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;\n", newSchema.TableName, oldCol.Name))
		}
	}

	return migration.String()
}

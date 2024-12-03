package proto_db

import (
	"fmt"
	"log"
	"strings"

	dbAn "github.com/imran31415/protobuf-db/db-annotations"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Convert DbColumnType enum to MySQL type
func dbColumnTypeToMySQLType(dbType dbAn.DbColumnType) string {
	switch dbType {
	case dbAn.DbColumnType_DB_TYPE_INT:
		return "INT"
	case dbAn.DbColumnType_DB_TYPE_VARCHAR:
		return "VARCHAR(255)"
	case dbAn.DbColumnType_DB_TYPE_TEXT:
		return "TEXT"
	case dbAn.DbColumnType_DB_TYPE_BOOLEAN:
		return "BOOLEAN"
	case dbAn.DbColumnType_DB_TYPE_DATETIME:
		return "DATETIME"
	case dbAn.DbColumnType_DB_TYPE_FLOAT:
		return "FLOAT"
	case dbAn.DbColumnType_DB_TYPE_DOUBLE:
		return "DOUBLE"
	case dbAn.DbColumnType_DB_TYPE_BINARY:
		return "BLOB"
	default:
		return "TEXT" // Default fallback
	}
}

func parseConstraints(constraints []dbAn.DbConstraint, defaultVal dbAn.DbDefault, customDefault string, updateAction dbAn.DbUpdateAction, dbType DatabaseType) []string {
	var result []string
	for _, constraint := range constraints {
		switch constraint {
		case dbAn.DbConstraint_DB_CONSTRAINT_NOT_NULL:
			result = append(result, "NOT NULL")
		case dbAn.DbConstraint_DB_CONSTRAINT_UNIQUE:
			result = append(result, "UNIQUE")
		}
	}

	// Handle default values
	switch defaultVal {
	case dbAn.DbDefault_DB_DEFAULT_FALSE:
		result = append(result, "DEFAULT FALSE")
	case dbAn.DbDefault_DB_DEFAULT_TRUE:
		result = append(result, "DEFAULT TRUE")
	case dbAn.DbDefault_DB_DEFAULT_CURRENT_TIMESTAMP:
		result = append(result, "DEFAULT CURRENT_TIMESTAMP")
	case dbAn.DbDefault_DB_DEFAULT_CUSTOM:
		if customDefault != "" {
			result = append(result, fmt.Sprintf("DEFAULT %s", customDefault))
		}
	}

	switch dbType {
	case DatabaseTypeSQLite:
		// Not supported in sqlite
	default:
		switch updateAction {
		case dbAn.DbUpdateAction_DB_UPDATE_ACTION_CURRENT_TIMESTAMP:
			result = append(result, "ON UPDATE CURRENT_TIMESTAMP")
		}
	}

	return result
}

// Join constraints into a single SQL-compatible string
func joinConstraints(constraints []string) string {
	return fmt.Sprintf("%s", combineConstraints(constraints))
}

// Combine constraints into a single string, separated by spaces
func combineConstraints(constraints []string) string {
	return fmt.Sprintf("%s", join(constraints, " "))
}

// Helper to join elements of a list with a specified separator
func join(list []string, sep string) string {
	if len(list) == 0 {
		return ""
	}

	joined := list[0]
	for _, item := range list[1:] {
		joined = fmt.Sprintf("%s%s%s", joined, sep, item)
	}

	return joined
}

// Check if two slices of constraints are equal
func equalConstraints(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]bool)
	for _, v := range a {
		m[v] = true
	}
	for _, v := range b {
		if !m[v] {
			return false
		}
	}
	return true
}

func parseIndexes(field protoreflect.FieldDescriptor) (string, error) {
	options := field.Options().(*descriptorpb.FieldOptions)
	if options == nil {
		return "", nil
	}

	indexType, ok := proto.GetExtension(options, dbAn.E_DbIndex).(dbAn.DbIndexType)
	if !ok || indexType == dbAn.DbIndexType_DB_INDEX_TYPE_UNSPECIFIED {
		return "", nil
	}

	switch indexType {
	case dbAn.DbIndexType_DB_INDEX_TYPE_SIMPLE:
		return "INDEX", nil
	case dbAn.DbIndexType_DB_INDEX_TYPE_FULLTEXT:
		return "FULLTEXT INDEX", nil
	case dbAn.DbIndexType_DB_INDEX_TYPE_SPATIAL:
		return "SPATIAL INDEX", nil
	default:
		return "", fmt.Errorf("unsupported index type: %v", indexType)
	}
}
func parseCompositeIndexes(md protoreflect.MessageDescriptor) []string {
	options, ok := md.Options().(*descriptorpb.MessageOptions)
	if !ok || options == nil {
		return nil
	}
	log.Println("Attempting to get composite raw indexes")

	// Extract the composite index string
	compositeIndexStr, ok := proto.GetExtension(options, dbAn.E_DbCompositeIndex).(string)
	if !ok || compositeIndexStr == "" {
		return nil
	}

	// Split by semicolons to get individual indexes
	rawIndexes := strings.Split(compositeIndexStr, ";")
	log.Println("Found composite raw indexes: ", rawIndexes)
	return rawIndexes
}
func parseForeignKeyAction(action dbAn.DbForeignKeyAction) string {
	switch action {
	case dbAn.DbForeignKeyAction_DB_FOREIGN_KEY_ACTION_CASCADE:
		return "CASCADE"
	case dbAn.DbForeignKeyAction_DB_FOREIGN_KEY_ACTION_SET_NULL:
		return "SET NULL"
	case dbAn.DbForeignKeyAction_DB_FOREIGN_KEY_ACTION_RESTRICT:
		return "RESTRICT"
	case dbAn.DbForeignKeyAction_DB_FOREIGN_KEY_ACTION_NO_ACTION:
		return "NO ACTION"
	default:
		return ""
	}
}
func parseTableLevelConstraints(md protoreflect.MessageDescriptor) (uniqueConstraints []string, checkConstraints []string) {
	options, ok := md.Options().(*descriptorpb.MessageOptions)
	if !ok || options == nil {
		return
	}

	// Parse composite UNIQUE constraints
	if proto.HasExtension(options, dbAn.E_DbUniqueConstraint) {
		unique, ok := proto.GetExtension(options, dbAn.E_DbUniqueConstraint).([]string)
		if ok && len(unique) > 0 {
			uniqueConstraints = append(uniqueConstraints, strings.Join(unique, ", "))
		}
	}

	// Parse CHECK constraints
	if proto.HasExtension(options, dbAn.E_DbCheckConstraint) {
		check, ok := proto.GetExtension(options, dbAn.E_DbCheckConstraint).([]string)
		if ok {
			checkConstraints = append(checkConstraints, check...)
		}
	}

	return
}

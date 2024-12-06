package proto_db

import (
	"fmt"
	"strings"

	dbAn "github.com/imran31415/protobuf-db/db-annotations"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func extractFieldSchema(field protoreflect.FieldDescriptor, dbType DatabaseType) (ColumnSchema, error) {
	var column ColumnSchema

	// Extract field options
	options := field.Options().(*descriptorpb.FieldOptions)
	if options == nil {
		return column, fmt.Errorf("missing field options for field %s", field.Name())
	}
	// Extract annotations with error checking
	dbColumn, ok := proto.GetExtension(options, dbAn.E_DbColumn).(string)
	if !ok || dbColumn == "" {
		return column, fmt.Errorf("missing or invalid db_column annotation")
	}

	dbColumnType, ok := proto.GetExtension(options, dbAn.E_DbColumnType).(dbAn.DbColumnType)
	if !ok {
		return column, fmt.Errorf("missing or invalid db_type annotation")
	}

	dbConstraints, ok := proto.GetExtension(options, dbAn.E_DbConstraints).([]dbAn.DbConstraint)
	if !ok {
		dbConstraints = []dbAn.DbConstraint{} // Default to no constraints if not specified
	}

	dbDefault, ok := proto.GetExtension(options, dbAn.E_DbDefault).(dbAn.DbDefault)
	if !ok {
		dbDefault = dbAn.DbDefault_DB_DEFAULT_UNSPECIFIED
	}

	customDefaultValue, ok := proto.GetExtension(options, dbAn.E_CustomDefaultValue).(string)
	if !ok {
		customDefaultValue = ""
	}

	dbPrimaryKey, ok := proto.GetExtension(options, dbAn.E_DbPrimaryKey).(bool)
	if !ok {
		dbPrimaryKey = false
	}

	dbUpdateAction, ok := proto.GetExtension(options, dbAn.E_DbUpdateAction).(dbAn.DbUpdateAction)
	if !ok {
		dbUpdateAction = dbAn.DbUpdateAction_DB_UPDATE_ACTION_UNSPECIFIED
	}

	// Foreign key data
	foreignKeyTable, _ := proto.GetExtension(options, dbAn.E_DbForeignKeyTable).(string)
	foreignKeyColumn, _ := proto.GetExtension(options, dbAn.E_DbForeignKeyColumn).(string)
	onDelete, _ := proto.GetExtension(options, dbAn.E_DbOnDelete).(dbAn.DbForeignKeyAction)
	onUpdate, ok := proto.GetExtension(options, dbAn.E_DbOnUpdate).(dbAn.DbForeignKeyAction)
	dbAutoIncrement, _ := proto.GetExtension(options, dbAn.E_DbAutoIncrement).(bool)
	dbPrecision, _ := proto.GetExtension(options, dbAn.E_DbPrecision).(int32)
	dbScale, _ := proto.GetExtension(options, dbAn.E_DbScale).(int32)

	// Format precision/scale for DECIMAL type
	precisionScale := ""
	if dbPrecision > 0 {
		precisionScale = fmt.Sprintf("(%d", dbPrecision)
		if dbScale > 0 {
			precisionScale += fmt.Sprintf(",%d", dbScale)
		}
		precisionScale += ")"
	}

	// Parse constraints and foreign key details
	constraints := parseConstraints(dbConstraints, dbDefault, customDefaultValue, dbUpdateAction, dbType)
	// Extract character set and collation
	characterSet, _ := proto.GetExtension(options, dbAn.E_DbCharacterSet).(string)
	collation, _ := proto.GetExtension(options, dbAn.E_DbCollate).(string)
	defaultFunction, _ := proto.GetExtension(options, dbAn.E_DbDefaultFunction).(dbAn.DbDefaultFunction)

	var defaultFunc string
	switch defaultFunction {
	case dbAn.DbDefaultFunction_DB_DEFAULT_FUNCTION_UUID:
		defaultFunc = "UUID()"
	case dbAn.DbDefaultFunction_DB_DEFAULT_FUNCTION_NOW:
		switch dbType {
		case DatabaseTypeSQLite:
			defaultFunc = "CURRENT_TIMESTAMP"
		default:
			defaultFunc = "CURRENT_TIMESTAMP"
		}
	}

	column = ColumnSchema{
		Name:             dbColumn,
		Type:             dbColumnTypeToMySQLType(dbColumnType),
		Constraints:      constraints,
		IsPrimaryKey:     dbPrimaryKey,
		ForeignKeyTable:  foreignKeyTable,
		ForeignKeyColumn: foreignKeyColumn,
		OnDelete:         parseForeignKeyAction(onDelete),
		OnUpdate:         parseForeignKeyAction(onUpdate),
		AutoIncrement:    dbAutoIncrement,
		Precision:        dbPrecision,
		Scale:            dbScale,
		CharacterSet:     characterSet,
		Collation:        collation,
		DefaultFunction:  defaultFunc,
	}

	return column, nil
}

func parseCompositePrimaryKeys(md protoreflect.MessageDescriptor) string {
	options, ok := md.Options().(*descriptorpb.MessageOptions)
	if !ok || options == nil {
		return ""
	}

	// fmt.Printf("Options: %+v\n", options)

	// Check if the extension is present before accessing it
	if !proto.HasExtension(options, dbAn.E_DbCompositePrimaryKey) {
		fmt.Println("No composite primary key extension found in options")
		return ""
	}

	// Safely extract the composite primary keys
	compositeKeys, ok := proto.GetExtension(options, dbAn.E_DbCompositePrimaryKey).(string)
	if !ok {
		fmt.Println("Failed to extract composite primary keys")
		return ""
	}

	// fmt.Printf("Composite Primary Keys Parsed: %v\n", compositeKeys)
	return compositeKeys
}

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

	// Extract the composite index string
	compositeIndexStr, ok := proto.GetExtension(options, dbAn.E_DbCompositeIndex).(string)
	if !ok || compositeIndexStr == "" {
		return nil
	}

	// Split by semicolons to get individual indexes
	rawIndexes := strings.Split(compositeIndexStr, ";")
	// log.Println("Found composite raw indexes: ", rawIndexes)
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
		if check, ok := proto.GetExtension(options, dbAn.E_DbCheckConstraint).([]string); ok {
			checkConstraints = check
		}
	}

	return
}

func generateIndexName(index string, tableName string) string {
	sanitizedIndex := strings.ReplaceAll(index, ",", "_") // Replace commas with underscores
	return fmt.Sprintf("%s_%s_idx", tableName, sanitizedIndex)
}

// Helper: Check if a column exists in the schema
func findColumnInSchema(columnName string, columns []ColumnSchema) (ColumnSchema, bool) {
	for _, col := range columns {
		if col.Name == columnName {
			return col, true
		}
	}
	return ColumnSchema{}, false
}

// Helper function to check if a slice contains a speci fic value
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

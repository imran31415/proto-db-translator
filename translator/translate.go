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

// DatabaseType represents the type of database (SQLite, MySQL, PostgreSQL, etc.)
type DatabaseType int

const (
	DatabaseTypeUnknown DatabaseType = iota
	DatabaseTypeSQLite
	DatabaseTypeMySQL
	DatabaseTypePostgreSQL
)

type DbConnection struct {
	DbType DatabaseType
	DbName string
	DbHost string
	DbPort string
	DbUser string
	DbPass string
}

func DefaultMysqlConnection() DbConnection {
	return DbConnection{
		DbType: DatabaseTypeMySQL,
		DbName: "prototestdb123",
		DbHost: "127.0.0.1",
		DbPort: "3306",
		DbUser: "root",
		// Just a default value obv don't use this locally
		DbPass: "Password123!",
	}
}

func DefaultSqliteConnection() DbConnection {
	return DbConnection{
		DbType: DatabaseTypeSQLite,
	}
}

type Translator struct {
	dbConnection DbConnection
}

func NewTranslator(in DbConnection) Translator {
	return Translator{
		dbConnection: in,
	}
}

func NewSqliteTranslator() Translator {
	return Translator{
		dbConnection: DefaultSqliteConnection(),
	}
}

// Schema represents the structure of a table for versioning
type Schema struct {
	TableName            string         `json:"table_name"`
	Columns              []ColumnSchema `json:"columns"`
	CompositePrimaryKeys []string       `json:"composite_primary_keys"` // Composite primary keys
	Indexes              []string       `json:"indexes"`                // Index definitions
	UniqueConstraints    []string       `json:"unique_constraints,omitempty"`
	CheckConstraints     []string       `json:"check_constraints,omitempty"`
	CompositIndexes      []string       `json:"composite_indexes,omitempty"`
}

// ColumnSchema represents the definition of a table column
type ColumnSchema struct {
	Name             string   `json:"name"`
	Type             string   `json:"type"`
	Constraints      []string `json:"constraints"`
	IsPrimaryKey     bool     `json:"is_primary_key"`
	ForeignKeyTable  string   `json:"foreign_key_table,omitempty"`  // Referenced table
	ForeignKeyColumn string   `json:"foreign_key_column,omitempty"` // Referenced column
	OnDelete         string   `json:"on_delete,omitempty"`          // Action on delete
	OnUpdate         string   `json:"on_update,omitempty"`          // Action on update
	AutoIncrement    bool     `json:"auto_increment,omitempty"`
	Precision        int32    `json:"precision,omitempty"`
	Scale            int32    `json:"scale,omitempty"`
	CharacterSet     string   `json:"character_set,omitempty"`    // New field
	Collation        string   `json:"collation,omitempty"`        // New field
	DefaultFunction  string   `json:"default_function,omitempty"` // New field for default functions

}

func (t Translator) GenerateSchema(message proto.Message) (Schema, error) {
	md := message.ProtoReflect().Descriptor()
	tableName := string(md.Name())

	var columns []ColumnSchema
	var indexes []string
	for i := 0; i < md.Fields().Len(); i++ {
		field := md.Fields().Get(i)
		c, err := extractFieldSchema(field, t.dbConnection.DbType)
		if err != nil {
			return Schema{}, err
		}

		// Parse index type for individual fields
		index, err := parseIndexes(field)
		if err != nil {
			return Schema{}, err
		}
		if index != "" {
			indexes = append(indexes, fmt.Sprintf("%s (%s)", index, c.Name))
		}
		columns = append(columns, c)
	}

	// Parse composite indexes
	compositeIndexes := parseCompositeIndexes(md)
	uniqueConstraints, checkConstraints := parseTableLevelConstraints(md)

	return Schema{
		TableName:            tableName,
		Columns:              columns,
		Indexes:              indexes,
		CompositIndexes:      compositeIndexes,
		CompositePrimaryKeys: parseCompositePrimaryKeys(md),
		UniqueConstraints:    uniqueConstraints,
		CheckConstraints:     checkConstraints,
	}, nil
}

func (t Translator) GenerateCreateTableSQL(schema Schema) string {
	var createStmt strings.Builder
	createStmt.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", schema.TableName))

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
		createStmt.WriteString(fmt.Sprintf(",\n  PRIMARY KEY (%s)", strings.Join(schema.CompositePrimaryKeys, ", ")))
	}

	// Add unique constraints
	for _, unique := range schema.UniqueConstraints {
		createStmt.WriteString(fmt.Sprintf(",\n  UNIQUE (%s)", unique))
	}

	// Add check constraints (optional)
	for _, check := range schema.CheckConstraints {
		createStmt.WriteString(fmt.Sprintf(",\n  CHECK (%s)", check))
	}

	// Add foreign key constraints as table-level constraints
	for _, col := range schema.Columns {
		if col.ForeignKeyTable != "" && col.ForeignKeyColumn != "" {
			createStmt.WriteString(fmt.Sprintf(",\n  FOREIGN KEY (%s) REFERENCES %s (%s)", col.Name, col.ForeignKeyTable, col.ForeignKeyColumn))
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

	for _, compositeIndex := range schema.CompositIndexes {
		createStmt.WriteString(fmt.Sprintf("\nCREATE INDEX %s ON %s (%s);",
			generateIndexName(compositeIndex, schema.TableName),
			schema.TableName,
			compositeIndex))
	}
	// Add composite index definitions

	log.Printf("----Table: %s, Statement: %s", schema.TableName, createStmt.String())
	return createStmt.String()
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

func parseCompositePrimaryKeys(md protoreflect.MessageDescriptor) []string {
	options, ok := md.Options().(*descriptorpb.MessageOptions)
	if !ok || options == nil {
		fmt.Println("No MessageOptions found for descriptor")
		return nil
	}

	fmt.Printf("Options: %+v\n", options)

	// Check if the extension is present before accessing it
	if !proto.HasExtension(options, dbAn.E_DbCompositePrimaryKey) {
		fmt.Println("No composite primary key extension found in options")
		return nil
	}

	// Safely extract the composite primary keys
	compositeKeys, ok := proto.GetExtension(options, dbAn.E_DbCompositePrimaryKey).([]string)
	if !ok {
		fmt.Println("Failed to extract composite primary keys")
		return nil
	}

	fmt.Printf("Composite Primary Keys Parsed: %v\n", compositeKeys)
	return compositeKeys
}

// Helper: Check if there are foreign keys in the schema
func hasForeignKeys(schema Schema) bool {
	for _, col := range schema.Columns {
		if col.ForeignKeyTable != "" && col.ForeignKeyColumn != "" {
			return true
		}
	}
	return false
}

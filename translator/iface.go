package proto_db

import "google.golang.org/protobuf/proto"

// TranslatorInterface defines the methods for the Translator struct.
type TranslatorInterface interface {
	GenerateSchema(message proto.Message) (Schema, error)          //converts the message along with annotations into a "Schema" representation
	GenerateCreateTableSQL(schema Schema) string                   // Generates the Create Table statement based on the schema
	ValidateSchema(protoMessage []proto.Message, dsn string) error // Validates the schema by applying the Create table statement to an actual database instance to validate the annotations
	GenerateModel(dbConnStr, tableName, outputDir string) error    // Leverages the Xo library to generate the database CRUD
	GenerateMigration(oldSchema, newSchema Schema) string          // Diffs 2 proto messages and determines the SQL migration to apply
}

// Validates existing Translator implementation against
var _ TranslatorInterface = (*Translator)(nil)

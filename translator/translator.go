package proto_db

import "google.golang.org/protobuf/proto"

// TranslatorInterface defines the methods for the Translator struct.
type TranslatorInterface interface {
	GenerateSchema(message proto.Message) (Schema, error)                 // Converts the message along with annotations into a "Schema" representation
	GenerateCreateTableSQL(schema Schema) string                          // Generates the Create Table statement based on the schema
	ValidateSchema(protoMessage []proto.Message, dsn string) error        // Validates the schema by applying the Create table statement to an actual database instance to validate the annotations
	GenerateModels(outputDir string, protoMessages []proto.Message) error // Leverages the Xo library to generate the database CRUD
	GenerateMigration(oldSchema, newSchema Schema) string                 // Diffs 2 proto messages and determines the SQL migration to apply

	// TODO:
	// ValidateMigration  // Validate the migration file produced by running the full series of migrations in a test database
	// GenerateApi // Automatically generate the basic Get/Update/Delete APIs based on the database schema and annotations.
}

// Validates existing Translator implementation against
var _ TranslatorInterface = (*Translator)(nil)

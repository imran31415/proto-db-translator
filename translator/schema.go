package proto_db

import "github.com/imran31415/proto-db-translator/translator/db"

type Translator struct {
	dbConnection db.DbConnection
}

func NewTranslator(in db.DbConnection) Translator {
	return Translator{
		dbConnection: in,
	}
}

func NewSqliteTranslator() Translator {
	return Translator{
		dbConnection: db.DefaultSqliteConnection(),
	}
}

// Schema represents the structure of a table for versioning
type Schema struct {
	TableName            string         `json:"table_name"`
	Columns              []ColumnSchema `json:"columns"`
	CompositePrimaryKeys string         `json:"composite_primary_keys"` // Composite primary keys
	Indexes              []string       `json:"indexes"`                // Index definitions
	UniqueConstraints    []string       `json:"unique_constraints,omitempty"`
	CheckConstraints     []string       `json:"check_constraints,omitempty"`
	CompositeIndexes     []string       `json:"composite_indexes,omitempty"`
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

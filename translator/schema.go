package proto_db

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
		DbName: "proto_db_default",
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

package db

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

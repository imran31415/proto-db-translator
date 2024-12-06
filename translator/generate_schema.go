package proto_db

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

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
		CompositeIndexes:     compositeIndexes,
		CompositePrimaryKeys: parseCompositePrimaryKeys(md),
		UniqueConstraints:    uniqueConstraints,
		CheckConstraints:     checkConstraints,
	}, nil
}

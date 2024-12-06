package proto_db

import (
	"testing"

	dbAn "github.com/imran31415/protobuf-db/db-annotations"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Helper to create a FieldDescriptorProto with extensions
func createFieldDescriptorProto(name, dbColumn string, dbColumnType dbAn.DbColumnType, constraints []dbAn.DbConstraint) *descriptorpb.FieldDescriptorProto {
	field := &descriptorpb.FieldDescriptorProto{
		Name:    proto.String(name),
		Type:    descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), // Ensures a valid type
		Options: &descriptorpb.FieldOptions{},
		Number:  proto.Int32(1), // Assign a valid field number

	}
	// Set extensions for the field
	proto.SetExtension(field.Options, dbAn.E_DbColumn, dbColumn)
	proto.SetExtension(field.Options, dbAn.E_DbColumnType, dbColumnType)
	if len(constraints) > 0 {
		proto.SetExtension(field.Options, dbAn.E_DbConstraints, constraints)
	}
	return field
}

func TestExtractFieldSchema(t *testing.T) {
	tests := []struct {
		name           string
		setupField     func() *descriptorpb.FieldDescriptorProto
		dbType         DatabaseType
		expectedResult ColumnSchema
		expectedError  string
	}{
		{
			name: "Valid field with all attributes",
			setupField: func() *descriptorpb.FieldDescriptorProto {
				return createFieldDescriptorProto(
					"username",
					"username",
					dbAn.DbColumnType_DB_TYPE_VARCHAR,
					[]dbAn.DbConstraint{dbAn.DbConstraint_DB_CONSTRAINT_NOT_NULL},
				)
			},
			dbType: DatabaseTypeMySQL,
			expectedResult: ColumnSchema{
				Name:        "username",
				Type:        "VARCHAR(255)",
				Constraints: []string{"NOT NULL"},
			},
			expectedError: "",
		},
		{
			name: "Missing db_column annotation",
			setupField: func() *descriptorpb.FieldDescriptorProto {
				return createFieldDescriptorProto(
					"missing_column",
					"",
					dbAn.DbColumnType_DB_TYPE_INT,
					nil,
				)
			},
			dbType:         DatabaseTypeMySQL,
			expectedResult: ColumnSchema{},
			expectedError:  "missing or invalid db_column annotation",
		},
		{
			name: "Field with default constraints",
			setupField: func() *descriptorpb.FieldDescriptorProto {
				return createFieldDescriptorProto(
					"balance",
					"balance",
					dbAn.DbColumnType_DB_TYPE_FLOAT,
					[]dbAn.DbConstraint{dbAn.DbConstraint_DB_CONSTRAINT_NOT_NULL},
				)
			},
			dbType: DatabaseTypeMySQL,
			expectedResult: ColumnSchema{
				Name:        "balance",
				Type:        "FLOAT",
				Constraints: []string{"NOT NULL"},
			},
			expectedError: "",
		},
		// {
		// 	name: "Invalid column type",
		// 	setupField: func() *descriptorpb.FieldDescriptorProto {
		// 		return createFieldDescriptorProto(
		// 			"unknown_type",
		// 			"unknown_type",
		// 			dbAn.DbColumnType_DB_TYPE_UNSPECIFIED,
		// 			nil,
		// 		)
		// 	},
		// 	dbType:         DatabaseTypeMySQL,
		// 	expectedResult: ColumnSchema{},
		// 	expectedError:  "missing or invalid db_type annotation",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the field descriptor
			field := tt.setupField()

			// Convert FieldDescriptorProto to protoreflect.FieldDescriptor
			fileDesc := &descriptorpb.FileDescriptorProto{
				Name:   proto.String("test_file.proto"),
				Syntax: proto.String("proto3"),
				MessageType: []*descriptorpb.DescriptorProto{
					{
						Name: proto.String("TestMessage"),
						Field: []*descriptorpb.FieldDescriptorProto{
							field,
						},
					},
				},
			}
			// Parse the file descriptor into protoreflect.FileDescriptor
			fd, err := protodesc.NewFile(fileDesc, nil)
			assert.NoError(t, err)

			// Retrieve the field descriptor
			fieldDesc := fd.Messages().ByName("TestMessage").Fields().ByName(protoreflect.Name(field.GetName()))

			// Call the function under test
			result, err := extractFieldSchema(fieldDesc, tt.dbType)

			// Assertions
			if tt.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			} else {
				assert.ErrorContains(t, err, tt.expectedError)
			}
		})
	}
}

// func TestParseCompositePrimaryKeys(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		setup          func() *descriptorpb.DescriptorProto
// 		expectedResult string
// 	}{
// 		{
// 			name: "Valid composite primary key",
// 			setup: func() *descriptorpb.DescriptorProto {
// 				message := &descriptorpb.DescriptorProto{
// 					Options: &descriptorpb.MessageOptions{},
// 				}
// 				// Set the extension manually
// 				message.Options.ProtoReflect().Set(
// 					dbAn.E_DbCompositePrimaryKey.TypeDescriptor(),
// 					protoreflect.ValueOfString("id,username"),
// 				)
// 				return message
// 			},
// 			expectedResult: "id,username",
// 		},
// 		{
// 			name: "No composite primary key",
// 			setup: func() *descriptorpb.DescriptorProto {
// 				return &descriptorpb.DescriptorProto{
// 					Options: &descriptorpb.MessageOptions{},
// 				}
// 			},
// 			expectedResult: "",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			message := tt.setup()
// 			result := parseCompositePrimaryKeys(message.ProtoReflect().Descriptor())
// 			assert.Equal(t, tt.expectedResult, result)
// 		})
// 	}
// }

// func TestParseCompositeIndexes(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		setup          func() protoreflect.MessageDescriptor
// 		expectedResult []string
// 	}{
// 		{
// 			name: "Valid composite indexes",
// 			setup: func() protoreflect.MessageDescriptor {
// 				message := &descriptorpb.DescriptorProto{
// 					Options: &descriptorpb.MessageOptions{},
// 				}
// 				proto.SetExtension(message.Options, dbAn.E_DbCompositeIndex, "id,username;email,phone")
// 				return message.ProtoReflect().Descriptor()
// 			},
// 			expectedResult: []string{"id,username", "email,phone"},
// 		},
// 		{
// 			name: "No composite indexes",
// 			setup: func() protoreflect.MessageDescriptor {
// 				return (&descriptorpb.DescriptorProto{}).ProtoReflect().Descriptor()
// 			},
// 			expectedResult: nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := parseCompositeIndexes(tt.setup())
// 			assert.ElementsMatch(t, tt.expectedResult, result)
// 		})
// 	}
// }

package main

import (
	"fmt"
	"os"

	proto_db "github.com/imran31415/proto-db-translator/translator"
	user_proto "github.com/imran31415/proto-db-translator/user"
	"google.golang.org/protobuf/proto"
	// "google.golang.org/protobuf/proto"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go-cli <output_dir> <proto_files...>")

		translator := proto_db.NewTranslator(proto_db.DefaultMysqlConnection())

		inputProtos := []proto.Message{
			&user_proto.User{},
			&user_proto.Role{},
			&user_proto.RoleHierarchy{}, // RoleHierarchy has a fk dependency on Role so must come after
		}

		translator.ProcessProtoMessages("../generated_models", inputProtos)

		fmt.Println("Successfully created trasnlator")
		os.Exit(1)
	}
}

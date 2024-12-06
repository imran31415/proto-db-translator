package main

import (
	"fmt"
	"log"

	proto_db "github.com/imran31415/proto-db-translator/translator"
	user_proto "github.com/imran31415/proto-db-translator/user"
	"google.golang.org/protobuf/proto"
	// "google.golang.org/protobuf/proto"
)

func main() {

	translator := proto_db.NewTranslator(proto_db.DefaultMysqlConnection())
	log.Println("successfully initialized translator")

	inputProtos := []proto.Message{
		&user_proto.User{},
		&user_proto.Role{},
		&user_proto.RoleHierarchy{},
		&user_proto.Customer{},
		&user_proto.Product{},
		&user_proto.Orders{},
		&user_proto.OrderDetails{},
		&user_proto.OrderItems{},
	}

	err := translator.GenerateModels("../generated_models", inputProtos)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Successfully created tables and modelsr")

}

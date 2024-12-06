package main

import (
	"fmt"
	"log"

	proto_db "github.com/imran31415/proto-db-translator/translator"
	"github.com/imran31415/proto-db-translator/translator/db"

	user_proto "github.com/imran31415/proto-db-translator/user"
	"google.golang.org/protobuf/proto"

	config_generator "github.com/imran31415/proto-db-translator/config_generator"
	// "google.golang.org/protobuf/proto"
)

func main() {

	conn := db.DefaultMysqlConnection()
	conn.DbName = "protodbtranslatortestdb"
	translator := proto_db.NewTranslator(conn)

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

	config_generator.GenerateConfig("../config")
	fmt.Println("Successfully created tables and modelsr")

}

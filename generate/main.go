package main

import (
	"fmt"
	"log"

	grpc_generator "github.com/imran31415/proto-db-translator/grpc_generator"
	proto_db "github.com/imran31415/proto-db-translator/translator"
	"github.com/imran31415/proto-db-translator/translator/db"
	"google.golang.org/protobuf/proto"

	user_proto "github.com/imran31415/proto-db-translator/user"

	config_generator "github.com/imran31415/proto-db-translator/config_generator"
)

func main() {

	// The following is example usage of this package to show what it can do

	// Step 1: Initialize the translator
	conn := db.DefaultMysqlConnection()
	conn.DbName = "protodbtranslatortestdb"
	translator := proto_db.NewTranslator(conn)

	log.Println("successfully initialized translator")
	// Pass in Protobuf Messages which each should represent a SQL "table" with appropriate annotations configured
	// These should be passed in strict order in which they would be created in SQL for FK dependencies.
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
	// Generate validated Create table statements that were validated by applying to an actual database
	statements, err := translator.ValidateSchema(inputProtos)
	log.Printf("Statements are: %v", statements)
	if err != nil {
		log.Println(err)
		return
	}

	// Generate go DB models for CRUD operations and even advanced queries like pagination
	// This will create a package in the outer directory
	err = translator.GenerateModels("../generated_models", inputProtos)
	if err != nil {
		log.Println(err)
		return
	}

	// Generate a config object which will be used to provision a GRPC server
	config_generator.GenerateConfig("../config")

	config := grpc_generator.GRPCServerConfig{
		ModuleName:      "github.com/imran31415/proto-db-translator",
		ProtobufPackage: "user",
		ProtobufService: "UserAuthService",
		DatabaseDriver:  "github.com/go-sql-driver/mysql",
		InsecurePort:    "50052",
	}

	// Generate the GRPC server base code
	err = grpc_generator.GenerateGRPCServer(config, "../grpcserver")
	if err != nil {
		fmt.Printf("Error generating gRPC server: %v\n", err)
	} else {
		fmt.Println("gRPC server generated successfully.")
	}
	fmt.Println("Successfully created tables and modelsr")

}

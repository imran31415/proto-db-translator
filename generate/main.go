package main

import (
	"fmt"
	"log"

	grpc_generator "github.com/imran31415/proto-db-translator/grpc_generator"
	proto_db "github.com/imran31415/proto-db-translator/translator"
	"github.com/imran31415/proto-db-translator/translator/db"
	user_proto "github.com/imran31415/proto-db-translator/user"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	config_generator "github.com/imran31415/proto-db-translator/config_generator"
)

func main() {
	// Step 1: Initialize the translator
	conn := db.DefaultMysqlConnection()
	conn.DbName = "protodbtranslatortestdb"
	translator := proto_db.NewTranslator(conn)

	log.Println("Successfully initialized translator")
	// Pass in Protobuf Messages which each should represent a SQL "table" with appropriate annotations configured
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

	descriptorPath := "../user/desc.pb"
	plugin, err := grpc_generator.ParseDescriptorFromFile(descriptorPath, protogen.Options{})
	if err != nil {
		fmt.Printf("Failed to parse descriptor: %v\n", err)
		return
	}

	outputDir := "./grpcimpl"
	if err := grpc_generator.GenerateGRPCImpl(plugin, outputDir); err != nil {
		fmt.Printf("Error generating gRPC implementation: %v\n", err)
	} else {
		fmt.Println("gRPC implementation generated successfully.")
	}

	fmt.Println("Successfully created tables, models, and gRPC server/implementation.")
}

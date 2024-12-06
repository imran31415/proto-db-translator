package main

import (
	"fmt"
	"log"

	grpcservergenerator "github.com/imran31415/proto-db-translator/grpc_generator"
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

	config := grpcservergenerator.GRPCServerConfig{
		ModuleName:      "github.com/imran31415/proto-db-translator",
		ProtobufPackage: "user",
		ProtobufService: "UserAuthService",
		DatabaseDriver:  "github.com/go-sql-driver/mysql",
		DatabaseDSN:     "root:Password123!@tcp(localhost:3306)/protodbtranslatortestdb",
		InsecurePort:    "50052",
	}

	err = grpcservergenerator.GenerateGRPCServer(config, "../grpcserver")
	if err != nil {
		fmt.Printf("Error generating gRPC server: %v\n", err)
	} else {
		fmt.Println("gRPC server generated successfully.")
	}
	fmt.Println("Successfully created tables and modelsr")

}

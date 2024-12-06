package grpcservergenerator

import (
	"fmt"
	"os"
	"path/filepath"
)

type GRPCServerConfig struct {
	ModuleName      string
	ProtobufPackage string
	ProtobufService string
	DatabaseDriver  string
	InsecurePort    string
}

func GenerateGRPCServer(config GRPCServerConfig, outputDir string) error {
	serverContent := fmt.Sprintf(`package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	"%[1]s/config"
	"%[1]s/%[2]s"

	_ "%[4]s" // Database driver
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Db *sql.DB
	%[2]s.Unimplemented%[3]sServer
}

func connectToDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%%s:%%s@tcp(%%s:%%s)/%%s?parseTime=true&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %%w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping the database: %%w", err)
	}

	log.Println("Successfully connected to the database.")
	return db, nil
}

func startGRPCServers(server *Server) {
	go func() {
		log.Println("Starting insecure gRPC server on port %[5]s...")
		insecureServer := grpc.NewServer()
		%[2]s.Register%[3]sServer(insecureServer, server)
		grpc_health_v1.RegisterHealthServer(insecureServer, health.NewServer())
		reflection.Register(insecureServer)

		listenAndServe(insecureServer, ":%[5]s")
	}()

	select {}
}

func listenAndServe(server *grpc.Server, port string) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %%s: %%v", port, err)
	}

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server on port %%s: %%v", port, err)
	}
}

func main() {
	cfg := config.LoadConfig()

	db, err := connectToDB(cfg.Database)
	if err != nil {
		log.Fatalf("Error connecting to the database: %%v", err)
	}
	defer db.Close()

	server := &Server{Db: db}

	startGRPCServers(server)
}
`,
		config.ModuleName,
		config.ProtobufPackage,
		config.ProtobufService,
		config.DatabaseDriver,
		config.InsecurePort,
	)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write to `main.go` file
	filePath := filepath.Join(outputDir, "main.go")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create main.go file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(serverContent)
	if err != nil {
		return fmt.Errorf("failed to write server content: %w", err)
	}

	fmt.Printf("gRPC server code generated successfully at %s\n", filePath)
	return nil
}

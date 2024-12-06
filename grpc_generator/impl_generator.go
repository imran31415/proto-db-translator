package grpcservergenerator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

// RPCMetadata contains metadata for generating gRPC methods.
type RPCMetadata struct {
	ServiceName  string
	MethodName   string
	Operation    string
	ModelName    string
	RequestType  string
	ResponseType string
}

func ParseDescriptorFromFile(descriptorPath string, options protogen.Options) (*protogen.Plugin, error) {
	// Read descriptor file
	descriptorData, err := os.ReadFile(descriptorPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read descriptor file: %w", err)
	}

	// Unmarshal into CodeGeneratorRequest
	codeGenRequest := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(descriptorData, codeGenRequest); err != nil {
		return nil, fmt.Errorf("failed to parse descriptor: %w", err)
	}

	// Log ProtoFile and FileToGenerate contents
	for _, file := range codeGenRequest.ProtoFile {
		fmt.Printf("ProtoFile: %s\n", file.GetName())
	}
	for _, fileToGenerate := range codeGenRequest.FileToGenerate {
		fmt.Printf("FileToGenerate: %s\n", fileToGenerate)
	}

	// Create protogen.Plugin
	plugin, err := options.New(codeGenRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create protogen.Plugin: %w", err)
	}

	return plugin, nil
}

// GenerateGRPCImplFromFiles reads proto files and generates the gRPC implementation.
func GenerateGRPCImplFromFiles(protoFiles []string, outputDir string) error {
	// Set up protogen options
	options := protogen.Options{}

	for _, pbGoFile := range protoFiles {
		// Parse the .pb.go file
		plugin, err := ParseDescriptorFromFile(pbGoFile, options)
		if err != nil {
			return fmt.Errorf("failed to parse .pb.go file %s: %w", pbGoFile, err)
		}

		// Generate the gRPC implementation
		if err := GenerateGRPCImpl(plugin, outputDir); err != nil {
			return fmt.Errorf("failed to generate gRPC implementation for file %s: %w", pbGoFile, err)
		}
	}

	return nil
}

// GenerateGRPCImpl generates the gRPC server implementation using a protogen.Plugin.
func GenerateGRPCImpl(plugin *protogen.Plugin, outputDir string) error {
	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}

		for _, service := range file.Services {
			// Generate service code
			generatedCode, err := generateServiceCode(service)
			if err != nil {
				return fmt.Errorf("failed to generate service code for %s: %w", service.Desc.Name(), err)
			}

			// Write the generated code to a file in the output directory
			outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_server_impl.go", service.GoName))
			if err := os.WriteFile(outputPath, []byte(generatedCode), 0644); err != nil {
				return fmt.Errorf("failed to write generated code to file: %w", err)
			}
		}
	}
	return nil
}

// generateServiceCode generates code for a single gRPC service.
func generateServiceCode(service *protogen.Service) (string, error) {
	// Read the template file
	templateContent, err := os.ReadFile("service_impl.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse the template
	tmpl, err := template.New("service").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Gather metadata for the service
	serviceMeta := struct {
		PackageName string
		ServiceName string
		Methods     []RPCMetadata
	}{
		PackageName: string(service.GoName),
		ServiceName: string(service.Desc.Name()),
		Methods:     extractRPCMetadata(service),
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, serviceMeta); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// extractRPCMetadata extracts metadata for the RPC methods of a service.
func extractRPCMetadata(service *protogen.Service) []RPCMetadata {
	var methods []RPCMetadata
	for _, method := range service.Methods {
		methods = append(methods, RPCMetadata{
			ServiceName:  string(service.Desc.Name()),
			MethodName:   string(method.Desc.Name()),
			Operation:    "UNKNOWN",                    // Placeholder, extend with custom annotations if needed.
			ModelName:    inferModelName(method.Input), // Assumes the input type matches the model.
			RequestType:  method.Input.GoIdent.String(),
			ResponseType: method.Output.GoIdent.String(),
		})
	}
	return methods
}

// inferModelName infers the model name from the input message.
func inferModelName(inputMessage *protogen.Message) string {
	return string(inputMessage.Desc.Name()) // Simple inference, extend as needed.
}

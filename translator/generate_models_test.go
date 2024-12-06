package proto_db

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	userauth "github.com/imran31415/proto-db-translator/user"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestProcessProtoMessages(t *testing.T) {

	outputDir := "../generated_models"

	if err := clearDirectory(outputDir); err != nil {
		fmt.Println("Error clearing dir:", err)
		t.Fail()
	}
	translator := NewTranslator(DefaultMysqlConnection())

	protoMessages := []proto.Message{
		&userauth.User{}, // Replace with your actual proto message types
		&userauth.Role{},
		&userauth.RoleHierarchy{},
		&userauth.Customer{},
		&userauth.Product{},
		&userauth.Orders{},
		&userauth.OrderDetails{},
		&userauth.OrderItems{},
	}

	err := translator.GenerateModels(outputDir, protoMessages)
	if err != nil {
		fmt.Println("err is", err)
	}
	require.NoError(t, err, "ProcessProtoMessages failed")
	filenames := []string{"db.xo.go", "orderdetail.xo.go", "rolehierarchy.xo.go", "user.xo.go"} // Replace with actual filenames

	err = checkFilesExist(outputDir, filenames)
	require.NoError(t, err, "Models were not generatedd")

}

// Check if all files exist in the directory
func checkFilesExist(outputDir string, filenames []string) error {
	for _, filename := range filenames {
		fullPath := filepath.Join(outputDir, filename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", fullPath)
		}
	}
	return nil
}

func clearDirectory(outputDir string) error {
	files, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	for _, file := range files {
		fullPath := filepath.Join(outputDir, file.Name())
		if err := os.Remove(fullPath); err != nil {
			return fmt.Errorf("failed to remove file %s: %v", fullPath, err)
		}
	}
	return nil
}

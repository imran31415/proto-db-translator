package grpcservergenerator

import "fmt"

// OperationType defines the type of database operation.
type OperationType int

const (
	OperationCreate OperationType = iota
	OperationRead
	OperationUpdate
	OperationDelete
	OperationList
)

// String converts the OperationType to its string representation.
func (op OperationType) String() string {
	switch op {
	case OperationCreate:
		return "CREATE"
	case OperationRead:
		return "READ"
	case OperationUpdate:
		return "UPDATE"
	case OperationDelete:
		return "DELETE"
	case OperationList:
		return "LIST"
	default:
		return "UNKNOWN"
	}
}

// ParseOperationType parses a string into an OperationType.
func ParseOperationType(value string) (OperationType, error) {
	switch value {
	case "CREATE":
		return OperationCreate, nil
	case "READ":
		return OperationRead, nil
	case "UPDATE":
		return OperationUpdate, nil
	case "DELETE":
		return OperationDelete, nil
	case "LIST":
		return OperationList, nil
	default:
		return -1, fmt.Errorf("unknown operation type: %s", value)
	}
}

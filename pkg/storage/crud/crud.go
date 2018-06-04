package crud

type Operation int

const (
	OperationCreate Operation = iota
	OperationUpdate           = iota
)

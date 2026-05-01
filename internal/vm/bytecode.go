package vm

import "slug/internal/object"

type Opcode byte

const (
	OpConstant Opcode = iota
	OpNil
	OpTrue
	OpFalse
	OpGetGlobal
	OpSetGlobalConst
	OpSetGlobalVar
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpLessThan
	OpBang
	OpNegate
	OpPop
	OpJump
	OpJumpIfFalse
	OpReturn
)

type Instruction struct {
	Op       Opcode
	IntArg   int
	StrArg   string
	Position int
}

type Chunk struct {
	Instructions []Instruction
	Constants    []object.Object
}

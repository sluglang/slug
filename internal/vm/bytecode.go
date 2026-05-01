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
	OpAssignGlobal
	OpArray
	OpHash
	OpIndex
	OpIndexDot
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
	OpCall
	OpReturn
)

type Instruction struct {
	Op       Opcode
	IntArg   int
	StrArg   string
	CallPlan []CallArgSpec
	Position int
}

type CallArgKind byte

const (
	CallArgPositional CallArgKind = iota
	CallArgNamed
	CallArgSpread
)

type CallArgSpec struct {
	Kind CallArgKind
	Name string
}

type Chunk struct {
	Instructions []Instruction
	Constants    []object.Object
}

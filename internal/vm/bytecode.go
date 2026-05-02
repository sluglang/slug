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
	OpBindMapAllConst
	OpBindMapAllVar
	OpAssignGlobal
	OpArray
	OpHash
	OpSlice
	OpIndex
	OpIndexDot
	OpDup
	OpSpawn
	OpAwait
	OpListPrepend
	OpListAppend
	OpMatchListEmpty
	OpMatchListHeadTail
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
	StrArg2  string
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

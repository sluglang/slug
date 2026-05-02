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
	OpMapHasKey
	OpDup
	OpSpawn
	OpAwait
	OpListPrepend
	OpListAppend
	OpMatchListEmpty
	OpMatchListHeadTail
	OpMatchSeqLenEq
	OpMatchSeqLenGte
	OpMatchSeqTail
	OpMatchMapLenEq
	OpMatchMapLenGte
	OpMatchMapBindRemainder
	OpPushScope
	OpPopScope
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpGreaterThanEqual
	OpLessThan
	OpLessThanEqual
	OpBitAnd
	OpBitOr
	OpBitXor
	OpShiftLeft
	OpShiftRight
	OpBang
	OpNegate
	OpBitNot
	OpPop
	OpJump
	OpJumpIfFalse
	OpCall
	OpRecur
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

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
	OpSetDoc
	OpBindMapAllConst
	OpBindMapAllVar
	OpAssignGlobal
	OpArray
	OpHash
	OpStructSchema
	OpStructInit
	OpStructCopy
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
	OpMatchStructSchema
	OpDefer
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
	OpApplyTags
	OpPop
	OpJump
	OpJumpIfFalse
	OpCall
	OpSelect
	OpRecur
	OpThrow
	OpReturn
)

type Instruction struct {
	Op       Opcode
	IntArg   int
	StrArg   string
	StrArg2  string
	CallPlan []CallArgSpec
	Select   []SelectCaseSpec
	Position int
}

type SelectCaseSpec struct {
	Kind      int
	TokenPos  int
	ChannelFn int
	ValueFn   int
	AfterFn   int
	AwaitFn   int
	HandlerFn int
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

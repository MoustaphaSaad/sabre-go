package spirv

import (
	"fmt"
	"strings"
)

// Module represents the top-level container for IR objects in the compilation unit.
// It manages ID generation and tracks all objects by their IDs.
type Module struct {
	idGenerator int
	objectsByID map[ID]Object
	typesByKey  map[string]Type
}

// ID is a unique identifier for IR objects within a Module.
type ID int

type Object interface {
	ID() ID
	Name() string
}

func NewModule() *Module {
	return &Module{
		idGenerator: 0,
	}
}

func (m *Module) NewID() ID {
	m.idGenerator++
	return ID(m.idGenerator)
}

func (m *Module) GetObject(id ID) Object {
	if obj, ok := m.objectsByID[id]; ok {
		return obj
	}
	return nil
}

func (m *Module) NewFunction(name string) *Func {
	f := &Func{
		ObjectID:   m.NewID(),
		ObjectName: name,
		Module:     m,
		Blocks:     make([]Block, 0),
	}
	m.objectsByID[f.ObjectID] = f
	return f
}

func (m *Module) InternVoid() *VoidType {
	t := &VoidType{}
	if t, ok := m.typesByKey[t.HashKey()]; ok {
		return t.(*VoidType)
	}
	t.ObjectID = m.NewID()
	t.ObjectName = t.HashKey()
	t.Module = m
	m.objectsByID[t.ObjectID] = t
	m.typesByKey[t.HashKey()] = t
	return t
}

func (m *Module) InternBool() *BoolType {
	t := &BoolType{}
	if t, ok := m.typesByKey[t.HashKey()]; ok {
		return t.(*BoolType)
	}
	t.ObjectID = m.NewID()
	t.ObjectName = t.HashKey()
	t.Module = m
	m.objectsByID[t.ObjectID] = t
	m.typesByKey[t.HashKey()] = t
	return t
}

func (m *Module) InternInt(bitWidth int, isSigned bool) *IntType {
	t := &IntType{
		BitWidth: bitWidth,
		IsSigned: isSigned,
	}
	if t, ok := m.typesByKey[t.HashKey()]; ok {
		return t.(*IntType)
	}
	t.ObjectID = m.NewID()
	t.ObjectName = t.HashKey()
	t.Module = m
	m.objectsByID[t.ObjectID] = t
	m.typesByKey[t.HashKey()] = t
	return t
}

func (m *Module) InternPtr(to Type, sc StorageClass) *PtrType {
	t := &PtrType{
		To:           to,
		StorageClass: sc,
	}
	if t, ok := m.typesByKey[t.HashKey()]; ok {
		return t.(*PtrType)
	}
	t.ObjectID = m.NewID()
	t.ObjectName = t.HashKey()
	t.Module = m
	m.objectsByID[t.ObjectID] = t
	m.typesByKey[t.HashKey()] = t
	return t
}

func (m *Module) InternFunc(returnType Type, args []Type) *FuncType {
	t := &FuncType{
		ReturnType: returnType,
		ArgTypes:   args,
	}
	if t, ok := m.typesByKey[t.HashKey()]; ok {
		return t.(*FuncType)
	}
	t.ObjectID = m.NewID()
	t.ObjectName = t.HashKey()
	t.Module = m
	m.objectsByID[t.ObjectID] = t
	m.typesByKey[t.HashKey()] = t
	return t
}

type StorageClass int

const (
	// Shared externally, visible across all invocations. Graphics uniform memory.
	// OpenCL constant memory. Variables declared with this storage class are read-only.
	// They may have initializers, as allowed by the client API.
	StorageClassUniformConstant StorageClass = 0
	// Input from pipeline. Visible only by the current invocation. Variables
	// declared with this storage class are read-only, and must not have initializers.
	StorageClassInput StorageClass = 1
	// Shared externally, visible across all invocations. Composite objects in this
	// storage class must have a type with an explicit layout.
	StorageClassUniform StorageClass = 2
	// Output to pipeline. Visible only by the current invocation.
	StorageClassOutput StorageClass = 3
	// Visible across all invocations within a workgroup.
	StorageClassWorkgroup StorageClass = 4
	// Visible across all invocations.
	StorageClassCrossWorkgroup StorageClass = 5
	// Visible only by the current invocation.
	StorageClassPrivate StorageClass = 6
	// Visible only by the current invocation. For memory allocation within
	// a function with specific lifetime. See OpVariable for more information.
	StorageClassFunction StorageClass = 7
	// For generic pointers, which overload the Function, Workgroup, and CrossWorkgroup Storage Classes.
	StorageClassGeneric StorageClass = 8
	// For holding push-constant memory, visible across all invocations. Intended to
	// contain a small bank of values pushed from the client API. Variables declared
	// with this storage class are read-only, and must not have initializers.
	// Composite objects in this storage class must have a type with an explicit layout.
	StorageClassPushConstant StorageClass = 9
	// For holding atomic counters. Visible only by the current invocation.
	StorageClassAtomicCounter StorageClass = 10
	// For holding image memory.
	StorageClassImage StorageClass = 11
	// Shared externally, readable and writable, visible across all invocations.
	// Composite objects in this storage class must have a type with an explicit layout.
	StorageClassStorageBuffer StorageClass = 12
)

type Type interface {
	Object
	aType()
	HashKey() string
}

type VoidType struct {
	ObjectID   ID
	ObjectName string
	Module     *Module
}

func (t VoidType) ID() ID {
	return t.ObjectID
}
func (t VoidType) Name() string {
	return t.ObjectName
}
func (VoidType) aType() {}
func (t VoidType) HashKey() string {
	return "void"
}

type BoolType struct {
	ObjectID   ID
	ObjectName string
	Module     *Module
}

func (t BoolType) ID() ID {
	return t.ObjectID
}
func (t BoolType) Name() string {
	return t.ObjectName
}
func (BoolType) aType() {}
func (t BoolType) HashKey() string {
	return "bool"
}

type IntType struct {
	ObjectID   ID
	ObjectName string
	Module     *Module
	BitWidth   int
	IsSigned   bool
}

func (t IntType) ID() ID {
	return t.ObjectID
}
func (t IntType) Name() string {
	return t.ObjectName
}
func (IntType) aType() {}
func (t IntType) HashKey() string {
	if t.IsSigned {
		return fmt.Sprintf("int%d", t.BitWidth)
	} else {
		return fmt.Sprintf("uint%d", t.BitWidth)
	}
}

type PtrType struct {
	ObjectID     ID
	ObjectName   string
	Module       *Module
	To           Type
	StorageClass StorageClass
}

func (t PtrType) ID() ID {
	return t.ObjectID
}
func (t PtrType) Name() string {
	return t.ObjectName
}
func (PtrType) aType() {}
func (t PtrType) HashKey() string {
	return fmt.Sprintf("ptr(%s,%d)", t.To.HashKey(), t.StorageClass)
}

type FuncType struct {
	ObjectID   ID
	ObjectName string
	Module     *Module
	ReturnType Type
	ArgTypes   []Type
}

func (t FuncType) ID() ID {
	return t.ObjectID
}
func (t FuncType) Name() string {
	return t.ObjectName
}
func (FuncType) aType() {}
func (t FuncType) HashKey() string {
	var b strings.Builder
	b.WriteString("func(")
	for i, arg := range t.ArgTypes {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(arg.HashKey())
	}
	b.WriteString(")")
	b.WriteString(t.ReturnType.HashKey())
	return b.String()
}

// Func represents a function in the IR, containing a sequence of basic blocks.
// Functions are the primary unit of executable code in the shader program.
type Func struct {
	ObjectID   ID
	ObjectName string
	Module     *Module
	Blocks     []Block
}

func (f Func) ID() ID {
	return f.ObjectID
}

func (f Func) Name() string {
	return f.ObjectName
}

func (f *Func) NewBlock(name string) *Block {
	b := &Block{
		ObjectID:     f.Module.NewID(),
		ObjectName:   name,
		Func:         f,
		Instructions: make([]Instr, 0),
		Terminated:   false,
	}
	f.Module.objectsByID[b.ObjectID] = b
	return b
}

// Block represents a basic block in the control flow graph.
// It contains a sequence of instructions that execute sequentially without branching.
type Block struct {
	ObjectID     ID
	ObjectName   string
	Func         *Func
	Instructions []Instr
	Terminated   bool
}

func (f Block) ID() ID {
	return f.ObjectID
}

func (f Block) Name() string {
	return f.ObjectName
}

func (b *Block) Push(ins Instr) {
	if b.Terminated {
		panic("New instructions can not be added to a terminated block")
	}

	b.Instructions = append(b.Instructions, ins)
}

// Op represents an operation code for IR instructions.
type Op int

const (
	OpNone        = 0
	OpReturn      = 253
	OpReturnValue = 254
)

// Instr represents an instruction in a basic block.
type Instr interface {
	Op() Op
}

type ReturnInstr struct{}

func (ReturnInstr) Op() Op { return OpReturn }

type ReturnValueInstr struct{ Value ID }

func (ReturnValueInstr) Op() Op { return OpReturnValue }

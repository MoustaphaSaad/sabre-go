package spirv

import (
	"fmt"
	"strings"
)

// ID represents a unique identifier for SPIR-V objects.
type ID int32

// Object is the base interface for all SPIR-V IR objects like functions and basic blocks, etc...
type Object interface {
	ID() ID
	Name() string
}

// BaseObject provides a basic implementation of the Object interface.
type BaseObject struct {
	ObjectID   ID
	ObjectName string
}

func (o BaseObject) ID() ID {
	return o.ObjectID
}
func (o BaseObject) Name() string {
	return o.ObjectName
}

// Constant represents a SPIR-V constant value.
type Constant interface {
	isConstant()
}

type BoolConstant struct {
	BaseObject
	Type  *BoolType
	Value bool
}

func (c *BoolConstant) isConstant() {}

type IntConstant struct {
	BaseObject
	Type  *IntType
	Value int64
}

func (c *IntConstant) isConstant() {}

type FloatConstant struct {
	BaseObject
	Type  *FloatType
	Value float64
}

func (c *FloatConstant) isConstant() {}

// RuntimeValue represents a value produced by an instruction at runtime.
type RuntimeValue struct {
	BaseObject
	Type Type
}

func (v *RuntimeValue) GetType() Type {
	return v.Type
}

// Module represents a SPIR-V module containing functions.
type Module struct {
	idGenerator     int
	objectsByID     map[ID]Object
	typesByKey      map[string]Type
	constantsByKey  map[string]Constant
	capabilities    []Capability
	AddressingModel AddressingModel
	MemoryModel     MemoryModel
}

func NewModule(addressingModel AddressingModel, memoryModel MemoryModel) *Module {
	return &Module{
		idGenerator:     0,
		objectsByID:     make(map[ID]Object),
		typesByKey:      make(map[string]Type),
		constantsByKey:  make(map[string]Constant),
		capabilities:    make([]Capability, 0),
		AddressingModel: addressingModel,
		MemoryModel:     memoryModel,
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

func (m *Module) NewFunction(name string, functionType *FuncType) *Function {
	f := &Function{
		BaseObject: BaseObject{
			ObjectID:   m.NewID(),
			ObjectName: name,
		},
		Module: m,
		Type:   functionType,
		Blocks: make([]*Block, 0),
	}
	m.objectsByID[f.ObjectID] = f
	return f
}

func (m *Module) InternVoid() *VoidType {
	t := &VoidType{}
	if existingType, ok := m.typesByKey[t.HashKey()]; ok {
		return existingType.(*VoidType)
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
	if existingType, ok := m.typesByKey[t.HashKey()]; ok {
		return existingType.(*BoolType)
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
	if existingType, ok := m.typesByKey[t.HashKey()]; ok {
		return existingType.(*IntType)
	}
	t.ObjectID = m.NewID()
	t.ObjectName = t.HashKey()
	t.Module = m
	m.objectsByID[t.ObjectID] = t
	m.typesByKey[t.HashKey()] = t
	return t
}

func (m *Module) InternFloat(bitWidth int) *FloatType {
	t := &FloatType{
		BitWidth: bitWidth,
	}
	if existingType, ok := m.typesByKey[t.HashKey()]; ok {
		return existingType.(*FloatType)
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
	if existingType, ok := m.typesByKey[t.HashKey()]; ok {
		return existingType.(*PtrType)
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
	if existingType, ok := m.typesByKey[t.HashKey()]; ok {
		return existingType.(*FuncType)
	}
	t.ObjectID = m.NewID()
	t.ObjectName = t.TypeName()
	t.Module = m
	m.objectsByID[t.ObjectID] = t
	m.typesByKey[t.HashKey()] = t
	return t
}

func (m *Module) AddCapability(cap Capability) {
	for _, c := range m.capabilities {
		if c == cap {
			return
		}
	}
	m.capabilities = append(m.capabilities, cap)
}

func (m *Module) Capabilities() []Capability {
	return m.capabilities
}

func (m *Module) InternBoolConstant(value bool, t *BoolType) *BoolConstant {
	key := fmt.Sprintf("const_%v_%v", t.HashKey(), value)
	if existing, ok := m.constantsByKey[key]; ok {
		return existing.(*BoolConstant)
	}

	id := m.NewID()
	constant := &BoolConstant{
		BaseObject: BaseObject{
			ObjectID:   id,
			ObjectName: key,
		},
		Type:  t,
		Value: value,
	}

	m.objectsByID[id] = constant
	m.constantsByKey[key] = constant
	return constant
}

func (m *Module) InternIntConstant(value int64, t *IntType) *IntConstant {
	key := fmt.Sprintf("const_%v_%v", t.HashKey(), value)
	if existing, ok := m.constantsByKey[key]; ok {
		return existing.(*IntConstant)
	}
	id := m.NewID()
	constant := &IntConstant{
		BaseObject: BaseObject{
			ObjectID:   id,
			ObjectName: key,
		},
		Type:  t,
		Value: value,
	}
	m.objectsByID[id] = constant
	m.constantsByKey[key] = constant
	return constant
}

func (m *Module) InternFloatConstant(value float64, t *FloatType) *FloatConstant {
	valueFmt := fmt.Sprintf("%f", value)
	valueName := strings.ReplaceAll(valueFmt, ".", "_")

	key := fmt.Sprintf("const_%v_%v", t.HashKey(), valueName)
	if existing, ok := m.constantsByKey[key]; ok {
		return existing.(*FloatConstant)
	}
	id := m.NewID()
	constant := &FloatConstant{
		BaseObject: BaseObject{
			ObjectID:   id,
			ObjectName: key,
		},
		Type:  t,
		Value: value,
	}
	m.objectsByID[id] = constant
	m.constantsByKey[key] = constant
	return constant
}

// NewNamedValue creates a new runtime value with the given name and type.
func (m *Module) NewNamedValue(name string, valueType Type) *RuntimeValue {
	id := m.NewID()
	value := &RuntimeValue{
		BaseObject: BaseObject{
			ObjectID:   id,
			ObjectName: name,
		},
		Type: valueType,
	}
	m.objectsByID[id] = value
	return value
}

// NewValue creates a new runtime value with the given type.
func (m *Module) NewValue(valueType Type) *RuntimeValue {
	return m.NewNamedValue("", valueType)
}

// Function represents a SPIR-V function containing a sequence of basic blocks.
type Function struct {
	BaseObject
	Module *Module
	Type   *FuncType
	Blocks []*Block
}

func (f *Function) NewBlock(name string) *Block {
	b := &Block{
		BaseObject: BaseObject{
			ObjectID:   f.Module.NewID(),
			ObjectName: name,
		},
		Function:     f,
		Instructions: make([]Instruction, 0),
	}
	f.Blocks = append(f.Blocks, b)
	f.Module.objectsByID[b.ObjectID] = b
	return b
}

// Block represents a basic block in a SPIR-V function containing a sequence of instructions.
type Block struct {
	BaseObject
	Function     *Function
	Instructions []Instruction
}

func (b *Block) Push(instr Instruction) {
	b.Instructions = append(b.Instructions, instr)
}

// Instruction represents a single SPIR-V instruction with an opcode.
type Instruction interface {
	Opcode() Opcode
}

type ConstantTrueInstruction struct {
	ResultType ID
	ResultID   ID
}

func (i *ConstantTrueInstruction) Opcode() Opcode {
	return OpConstantTrue
}

type ConstantFalseInstruction struct {
	ResultType ID
	ResultID   ID
}

func (i *ConstantFalseInstruction) Opcode() Opcode {
	return OpConstantFalse
}

type ReturnInstruction struct{}

func (r *ReturnInstruction) Opcode() Opcode {
	return OpReturn
}

type ReturnValueInstruction struct {
	Value ID
}

func (r *ReturnValueInstruction) Opcode() Opcode {
	return OpReturnValue
}

type SNegateInstruction struct {
	ResultType ID
	ResultID   ID
	Operand    ID
}

func (i *SNegateInstruction) Opcode() Opcode {
	return OpSNegate
}

type FNegateInstruction struct {
	ResultType ID
	ResultID   ID
	Operand    ID
}

func (i *FNegateInstruction) Opcode() Opcode {
	return OpFNegate
}

type LogicalOrInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *LogicalOrInstruction) Opcode() Opcode {
	return OpLogicalOr
}

type LogicalNotInstruction struct {
	ResultType ID
	ResultID   ID
	Operand    ID
}

func (i *LogicalNotInstruction) Opcode() Opcode {
	return OpLogicalNot
}

type NotInstruction struct {
	ResultType ID
	ResultID   ID
	Operand    ID
}

func (i *NotInstruction) Opcode() Opcode {
	return OpNot
}

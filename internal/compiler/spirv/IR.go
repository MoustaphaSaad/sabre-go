package spirv

import (
	"fmt"
	"slices"
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

type Value interface {
	Object
	GetType() Type
}

// ConstantValue represents a SPIR-V constant value.
type ConstantValue interface {
	Value
	isConstantValue()
}

type BoolConstant struct {
	BaseObject
	Type  *BoolType
	Value bool
}

func (c *BoolConstant) GetType() Type    { return c.Type }
func (c *BoolConstant) isConstantValue() {}

type IntConstant struct {
	BaseObject
	Type  *IntType
	Value int64
}

func (c *IntConstant) GetType() Type    { return c.Type }
func (c *IntConstant) isConstantValue() {}

type FloatConstant struct {
	BaseObject
	Type  *FloatType
	Value float64
}

func (c *FloatConstant) GetType() Type    { return c.Type }
func (c *FloatConstant) isConstantValue() {}

// RuntimeValue represents a value produced by an instruction at runtime.
type RuntimeValue struct {
	BaseObject
	Type Type
}

func (v *RuntimeValue) GetType() Type { return v.Type }

// Module represents a SPIR-V module containing functions.
type Module struct {
	idGenerator     int
	Objects         []Object
	objectsByID     map[ID]int
	typesByKey      map[string]int
	constantsByKey  map[string]int
	capabilities    []Capability
	AddressingModel AddressingModel
	MemoryModel     MemoryModel
}

func NewModule(addressingModel AddressingModel, memoryModel MemoryModel) *Module {
	return &Module{
		idGenerator:     0,
		Objects:         make([]Object, 0),
		objectsByID:     make(map[ID]int),
		typesByKey:      make(map[string]int),
		constantsByKey:  make(map[string]int),
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
	if index, ok := m.objectsByID[id]; ok {
		return m.Objects[index]
	}
	return nil
}

func (m *Module) RemoveObjects(ids []ID) {
	for _, id := range ids {
		m.Objects[m.objectsByID[id]] = nil
	}

	m.Objects = slices.DeleteFunc(m.Objects, func(obj Object) bool {
		return obj == nil
	})

	newObjectsByID := make(map[ID]int, len(m.Objects))
	for i, obj := range m.Objects {
		newObjectsByID[obj.ID()] = i
	}
	m.objectsByID = newObjectsByID
}

func (m *Module) addObject(obj Object) {
	index := len(m.Objects)
	m.objectsByID[obj.ID()] = index
	if t, ok := obj.(Type); ok {
		m.typesByKey[t.HashKey()] = index
	}
	if c, ok := obj.(ConstantValue); ok {
		m.constantsByKey[c.Name()] = index
	}
	m.Objects = append(m.Objects, obj)
}

func (m *Module) NewFuncParam(name string, paramType Type) *FuncParam {
	p := &FuncParam{
		BaseObject: BaseObject{
			ObjectID:   m.NewID(),
			ObjectName: name,
		},
		Type: paramType,
	}
	m.addObject(p)
	return p
}

func (m *Module) NewFunction(name string, functionType *FuncType, params []*FuncParam) *Function {
	f := &Function{
		BaseObject: BaseObject{
			ObjectID:   m.NewID(),
			ObjectName: name,
		},
		Module: m,
		Type:   functionType,
		Params: params,
		Blocks: make([]*Block, 0),
	}
	m.addObject(f)
	return f
}

func (m *Module) NewVariable(name string, ptrType *PtrType, sc StorageClass) *Variable {
	v := &Variable{
		BaseObject: BaseObject{
			ObjectID:   m.NewID(),
			ObjectName: name,
		},
		Type:         ptrType,
		StorageClass: sc,
	}
	m.addObject(v)
	return v
}

func (m *Module) InternVoid() *VoidType {
	t := &VoidType{}
	if index, ok := m.typesByKey[t.HashKey()]; ok {
		return m.Objects[index].(*VoidType)
	}
	t.ObjectID = m.NewID()
	t.ObjectName = t.HashKey()
	t.Module = m
	m.addObject(t)
	return t
}

func (m *Module) InternBool() *BoolType {
	t := &BoolType{}
	if index, ok := m.typesByKey[t.HashKey()]; ok {
		return m.Objects[index].(*BoolType)
	}
	t.ObjectID = m.NewID()
	t.ObjectName = t.HashKey()
	t.Module = m
	m.addObject(t)
	return t
}

func (m *Module) InternInt(bitWidth int, isSigned bool) *IntType {
	t := &IntType{
		BitWidth: bitWidth,
		IsSigned: isSigned,
	}
	if index, ok := m.typesByKey[t.HashKey()]; ok {
		return m.Objects[index].(*IntType)
	}
	t.ObjectID = m.NewID()
	t.ObjectName = t.HashKey()
	t.Module = m
	m.addObject(t)
	return t
}

func (m *Module) InternFloat(bitWidth int) *FloatType {
	t := &FloatType{
		BitWidth: bitWidth,
	}
	if index, ok := m.typesByKey[t.HashKey()]; ok {
		return m.Objects[index].(*FloatType)
	}
	t.ObjectID = m.NewID()
	t.ObjectName = t.HashKey()
	t.Module = m
	m.addObject(t)
	return t
}

func (m *Module) InternPtr(to Type, sc StorageClass) *PtrType {
	t := &PtrType{
		To:           to,
		StorageClass: sc,
	}
	if index, ok := m.typesByKey[t.HashKey()]; ok {
		return m.Objects[index].(*PtrType)
	}
	t.ObjectID = m.NewID()
	t.ObjectName = t.TypeName()
	t.Module = m
	m.addObject(t)
	return t
}

func (m *Module) InternFunc(returnType Type, args []Type) *FuncType {
	t := &FuncType{
		ReturnType: returnType,
		ArgTypes:   args,
	}
	if index, ok := m.typesByKey[t.HashKey()]; ok {
		return m.Objects[index].(*FuncType)
	}
	t.ObjectID = m.NewID()
	t.ObjectName = t.TypeName()
	t.Module = m
	m.addObject(t)
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
	if index, ok := m.constantsByKey[key]; ok {
		return m.Objects[index].(*BoolConstant)
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

	m.addObject(constant)
	return constant
}

func (m *Module) InternIntConstant(value int64, t *IntType) *IntConstant {
	key := fmt.Sprintf("const_%v_%v", t.HashKey(), value)
	if index, ok := m.constantsByKey[key]; ok {
		return m.Objects[index].(*IntConstant)
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
	m.addObject(constant)
	return constant
}

func (m *Module) InternFloatConstant(value float64, t *FloatType) *FloatConstant {
	valueFmt := fmt.Sprintf("%f", value)
	valueName := strings.ReplaceAll(valueFmt, ".", "_")

	key := fmt.Sprintf("const_%v_%v", t.HashKey(), valueName)
	if index, ok := m.constantsByKey[key]; ok {
		return m.Objects[index].(*FloatConstant)
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
	m.addObject(constant)
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
	m.addObject(value)
	return value
}

// NewValue creates a new runtime value with the given type.
func (m *Module) NewValue(valueType Type) *RuntimeValue {
	return m.NewNamedValue("", valueType)
}

type FuncParam struct {
	BaseObject
	Type Type
}

// Function represents a SPIR-V function containing a sequence of basic blocks.
type Function struct {
	BaseObject
	Module *Module
	Type   *FuncType
	Params []*FuncParam
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
	f.Module.addObject(b)
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

type Variable struct {
	BaseObject
	Type         *PtrType
	StorageClass StorageClass
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

type LogicalAndInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *LogicalAndInstruction) Opcode() Opcode {
	return OpLogicalAnd
}

type LogicalNotInstruction struct {
	ResultType ID
	ResultID   ID
	Operand    ID
}

func (i *LogicalNotInstruction) Opcode() Opcode {
	return OpLogicalNot
}

type LogicalEqualInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *LogicalEqualInstruction) Opcode() Opcode {
	return OpLogicalEqual
}

type LogicalNotEqualInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *LogicalNotEqualInstruction) Opcode() Opcode {
	return OpLogicalNotEqual
}

type UGreaterThanInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *UGreaterThanInstruction) Opcode() Opcode {
	return OpUGreaterThan
}

type SGreaterThanInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *SGreaterThanInstruction) Opcode() Opcode {
	return OpSGreaterThan
}

type ULessThanInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *ULessThanInstruction) Opcode() Opcode {
	return OpULessThan
}

type SLessThanInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *SLessThanInstruction) Opcode() Opcode {
	return OpSLessThan
}

type ULessThanEqualInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *ULessThanEqualInstruction) Opcode() Opcode {
	return OpULessThanEqual
}

type SLessThanEqualInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *SLessThanEqualInstruction) Opcode() Opcode {
	return OpSLessThanEqual
}

type FOrdLessThanInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *FOrdLessThanInstruction) Opcode() Opcode {
	return OpFOrdLessThan
}

type FOrdGreaterThanInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *FOrdGreaterThanInstruction) Opcode() Opcode {
	return OpFOrdGreaterThan
}

type FOrdLessThanEqualInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *FOrdLessThanEqualInstruction) Opcode() Opcode {
	return OpFOrdLessThanEqual
}

type UGreaterThanEqualInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *UGreaterThanEqualInstruction) Opcode() Opcode {
	return OpUGreaterThanEqual
}

type SGreaterThanEqualInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *SGreaterThanEqualInstruction) Opcode() Opcode {
	return OpSGreaterThanEqual
}

type FOrdGreaterThanEqualInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *FOrdGreaterThanEqualInstruction) Opcode() Opcode {
	return OpFOrdGreaterThanEqual
}

type IEqualInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *IEqualInstruction) Opcode() Opcode {
	return OpIEqual
}

type FOrdEqualInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *FOrdEqualInstruction) Opcode() Opcode {
	return OpFOrdEqual
}

type INotEqualInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *INotEqualInstruction) Opcode() Opcode {
	return OpINotEqual
}

type FOrdNotEqualInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *FOrdNotEqualInstruction) Opcode() Opcode {
	return OpFOrdNotEqual
}

type IAddInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *IAddInstruction) Opcode() Opcode {
	return OpIAdd
}

type FAddInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *FAddInstruction) Opcode() Opcode {
	return OpFAdd
}

type ISubInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *ISubInstruction) Opcode() Opcode {
	return OpISub
}

type FSubInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *FSubInstruction) Opcode() Opcode {
	return OpFSub
}

type IMulInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *IMulInstruction) Opcode() Opcode {
	return OpIMul
}

type FMulInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *FMulInstruction) Opcode() Opcode {
	return OpFMul
}

type UDivInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *UDivInstruction) Opcode() Opcode {
	return OpUDiv
}

type SDivInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *SDivInstruction) Opcode() Opcode {
	return OpSDiv
}

type FDivInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *FDivInstruction) Opcode() Opcode {
	return OpFDiv
}

type UModInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *UModInstruction) Opcode() Opcode {
	return OpUMod
}

type SRemInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *SRemInstruction) Opcode() Opcode {
	return OpSRem
}

type FRemInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *FRemInstruction) Opcode() Opcode {
	return OpFRem
}

type BitwiseXorInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *BitwiseXorInstruction) Opcode() Opcode {
	return OpBitwiseXor
}

type BitwiseOrInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *BitwiseOrInstruction) Opcode() Opcode {
	return OpBitwiseOr
}

type BitwiseAndInstruction struct {
	ResultType ID
	ResultID   ID
	Operand1   ID
	Operand2   ID
}

func (i *BitwiseAndInstruction) Opcode() Opcode {
	return OpBitwiseAnd
}

type NotInstruction struct {
	ResultType ID
	ResultID   ID
	Operand    ID
}

func (i *NotInstruction) Opcode() Opcode {
	return OpNot
}

type ShiftLeftLogicalInstruction struct {
	ResultType ID
	ResultID   ID
	Base       ID
	Shift      ID
}

func (i *ShiftLeftLogicalInstruction) Opcode() Opcode {
	return OpShiftLeftLogical
}

type ShiftRightLogicalInstruction struct {
	ResultType ID
	ResultID   ID
	Base       ID
	Shift      ID
}

func (i *ShiftRightLogicalInstruction) Opcode() Opcode {
	return OpShiftRightLogical
}

type ShiftRightArithmeticInstruction struct {
	ResultType ID
	ResultID   ID
	Base       ID
	Shift      ID
}

func (i *ShiftRightArithmeticInstruction) Opcode() Opcode {
	return OpShiftRightArithmetic
}

type FunctionCallInstruction struct {
	ResultType ID
	ResultID   ID
	FunctionID ID
	Args       []ID
}

func (i *FunctionCallInstruction) Opcode() Opcode {
	return OpFunctionCall
}

type VariableInstruction struct {
	ResultType   ID
	ResultID     ID
	StorageClass StorageClass
	Initializer  ID
}

func (i *VariableInstruction) Opcode() Opcode {
	return OpVariable
}

type LoadInstruction struct {
	ResultType ID
	ResultID   ID
	Pointer    ID
}

func (i *LoadInstruction) Opcode() Opcode {
	return OpLoad
}

type StoreInstruction struct {
	Pointer ID
	Object  ID
}

func (i *StoreInstruction) Opcode() Opcode {
	return OpStore
}

type UnreachableInstruction struct{}

func (r *UnreachableInstruction) Opcode() Opcode {
	return OpUnreachable
}

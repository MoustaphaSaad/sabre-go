package spirv

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

// Module represents a SPIR-V module containing functions.
type Module struct {
	idGenerator     int
	objectsByID     map[ID]Object
	capabilities    []Capability
	AddressingModel AddressingModel
	MemoryModel     MemoryModel
}

func NewModule(addressingModel AddressingModel, memoryModel MemoryModel) *Module {
	return &Module{
		idGenerator:     0,
		objectsByID:     make(map[ID]Object),
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

func (m *Module) NewFunction(name string) *Function {
	f := &Function{
		BaseObject: BaseObject{
			ObjectID:   m.NewID(),
			ObjectName: name,
		},
		Module: m,
		Blocks: make([]Block, 0),
	}
	m.objectsByID[f.ObjectID] = f
	return f
}

// Function represents a SPIR-V function containing a sequence of basic blocks.
type Function struct {
	BaseObject
	Module *Module
	Blocks []Block
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
	f.Blocks = append(f.Blocks, *b)
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

type ReturnInstruction struct{}

func (r ReturnInstruction) Opcode() Opcode {
	return OpReturn
}

type ReturnValueInstruction struct {
	Value ID
}

func (r ReturnValueInstruction) Opcode() Opcode {
	return OpReturnValue
}

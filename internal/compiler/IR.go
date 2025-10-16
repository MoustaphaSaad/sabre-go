package compiler

// Module represents the top-level container for IR objects in the compilation unit.
// It manages ID generation and tracks all objects by their IDs.
type Module struct {
	idGenerator int
	objectsByID map[ID]Object
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

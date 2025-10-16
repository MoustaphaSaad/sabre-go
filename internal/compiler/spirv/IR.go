package spirv

import (
	"fmt"
	"strings"
)

// Capability represents capabilities a module can declare it uses.
// All used capabilities need to be declared, either explicitly with OpCapability
// or implicitly through the Implicitly Declares column.
type Capability int

const (
	CapabilityMatrix                            Capability = 0
	CapabilityShader                            Capability = 1
	CapabilityGeometry                          Capability = 2
	CapabilityTessellation                      Capability = 3
	CapabilityAddresses                         Capability = 4
	CapabilityLinkage                           Capability = 5
	CapabilityKernel                            Capability = 6
	CapabilityVector16                          Capability = 7
	CapabilityFloat16Buffer                     Capability = 8
	CapabilityFloat16                           Capability = 9
	CapabilityFloat64                           Capability = 10
	CapabilityInt64                             Capability = 11
	CapabilityInt64Atomics                      Capability = 12
	CapabilityImageBasic                        Capability = 13
	CapabilityImageReadWrite                    Capability = 14
	CapabilityImageMipmap                       Capability = 15
	CapabilityPipes                             Capability = 17
	CapabilityGroups                            Capability = 18
	CapabilityDeviceEnqueue                     Capability = 19
	CapabilityLiteralSampler                    Capability = 20
	CapabilityAtomicStorage                     Capability = 21
	CapabilityInt16                             Capability = 22
	CapabilityTessellationPointSize             Capability = 23
	CapabilityGeometryPointSize                 Capability = 24
	CapabilityImageGatherExtended               Capability = 25
	CapabilityStorageImageMultisample           Capability = 27
	CapabilityUniformBufferArrayDynamicIndexing Capability = 28
	CapabilitySampledImageArrayDynamicIndexing  Capability = 29
	CapabilityStorageBufferArrayDynamicIndexing Capability = 30
	CapabilityStorageImageArrayDynamicIndexing  Capability = 31
	CapabilityClipDistance                      Capability = 32
	CapabilityCullDistance                      Capability = 33
	CapabilityImageCubeArray                    Capability = 34
	CapabilitySampleRateShading                 Capability = 35
	CapabilityImageRect                         Capability = 36
	CapabilitySampledRect                       Capability = 37
	CapabilityGenericPointer                    Capability = 38
	CapabilityInt8                              Capability = 39
	CapabilityInputAttachment                   Capability = 40
	CapabilitySparseResidency                   Capability = 41
	CapabilityMinLod                            Capability = 42
	CapabilitySampled1D                         Capability = 43
	CapabilityImage1D                           Capability = 44
	CapabilitySampledCubeArray                  Capability = 45
	CapabilitySampledBuffer                     Capability = 46
	CapabilityImageBuffer                       Capability = 47
	CapabilityImageMSArray                      Capability = 48
	CapabilityStorageImageExtendedFormats       Capability = 49
	CapabilityImageQuery                        Capability = 50
	CapabilityDerivativeControl                 Capability = 51
	CapabilityInterpolationFunction             Capability = 52
	CapabilityTransformFeedback                 Capability = 53
	CapabilityGeometryStreams                   Capability = 54
	CapabilityStorageImageReadWithoutFormat     Capability = 55
	CapabilityStorageImageWriteWithoutFormat    Capability = 56
	CapabilityMultiViewport                     Capability = 57
	CapabilitySubgroupDispatch                  Capability = 58
	CapabilityNamedBarrier                      Capability = 59
	CapabilityPipeStorage                       Capability = 60
	CapabilityGroupNonUniform                   Capability = 61
	CapabilityGroupNonUniformVote               Capability = 62
	CapabilityGroupNonUniformArithmetic         Capability = 63
	CapabilityGroupNonUniformBallot             Capability = 64
	CapabilityGroupNonUniformShuffle            Capability = 65
	CapabilityGroupNonUniformShuffleRelative    Capability = 66
	CapabilityGroupNonUniformClustered          Capability = 67
	CapabilityGroupNonUniformQuad               Capability = 68
	CapabilityShaderLayer                       Capability = 69
	CapabilityShaderViewportIndex               Capability = 70
	CapabilityUniformDecoration                 Capability = 71
)

func (c Capability) String() string {
	switch c {
	case CapabilityMatrix:
		return "Matrix"
	case CapabilityShader:
		return "Shader"
	case CapabilityGeometry:
		return "Geometry"
	case CapabilityTessellation:
		return "Tessellation"
	case CapabilityAddresses:
		return "Addresses"
	case CapabilityLinkage:
		return "Linkage"
	case CapabilityKernel:
		return "Kernel"
	case CapabilityVector16:
		return "Vector16"
	case CapabilityFloat16Buffer:
		return "Float16Buffer"
	case CapabilityFloat16:
		return "Float16"
	case CapabilityFloat64:
		return "Float64"
	case CapabilityInt64:
		return "Int64"
	case CapabilityInt64Atomics:
		return "Int64Atomics"
	case CapabilityImageBasic:
		return "ImageBasic"
	case CapabilityImageReadWrite:
		return "ImageReadWrite"
	case CapabilityImageMipmap:
		return "ImageMipmap"
	case CapabilityPipes:
		return "Pipes"
	case CapabilityGroups:
		return "Groups"
	case CapabilityDeviceEnqueue:
		return "DeviceEnqueue"
	case CapabilityLiteralSampler:
		return "LiteralSampler"
	case CapabilityAtomicStorage:
		return "AtomicStorage"
	case CapabilityInt16:
		return "Int16"
	case CapabilityTessellationPointSize:
		return "TessellationPointSize"
	case CapabilityGeometryPointSize:
		return "GeometryPointSize"
	case CapabilityImageGatherExtended:
		return "ImageGatherExtended"
	case CapabilityStorageImageMultisample:
		return "StorageImageMultisample"
	case CapabilityUniformBufferArrayDynamicIndexing:
		return "UniformBufferArrayDynamicIndexing"
	case CapabilitySampledImageArrayDynamicIndexing:
		return "SampledImageArrayDynamicIndexing"
	case CapabilityStorageBufferArrayDynamicIndexing:
		return "StorageBufferArrayDynamicIndexing"
	case CapabilityStorageImageArrayDynamicIndexing:
		return "StorageImageArrayDynamicIndexing"
	case CapabilityClipDistance:
		return "ClipDistance"
	case CapabilityCullDistance:
		return "CullDistance"
	case CapabilityImageCubeArray:
		return "ImageCubeArray"
	case CapabilitySampleRateShading:
		return "SampleRateShading"
	case CapabilityImageRect:
		return "ImageRect"
	case CapabilitySampledRect:
		return "SampledRect"
	case CapabilityGenericPointer:
		return "GenericPointer"
	case CapabilityInt8:
		return "Int8"
	case CapabilityInputAttachment:
		return "InputAttachment"
	case CapabilitySparseResidency:
		return "SparseResidency"
	case CapabilityMinLod:
		return "MinLod"
	case CapabilitySampled1D:
		return "Sampled1D"
	case CapabilityImage1D:
		return "Image1D"
	case CapabilitySampledCubeArray:
		return "SampledCubeArray"
	case CapabilitySampledBuffer:
		return "SampledBuffer"
	case CapabilityImageBuffer:
		return "ImageBuffer"
	case CapabilityImageMSArray:
		return "ImageMSArray"
	case CapabilityStorageImageExtendedFormats:
		return "StorageImageExtendedFormats"
	case CapabilityImageQuery:
		return "ImageQuery"
	case CapabilityDerivativeControl:
		return "DerivativeControl"
	case CapabilityInterpolationFunction:
		return "InterpolationFunction"
	case CapabilityTransformFeedback:
		return "TransformFeedback"
	case CapabilityGeometryStreams:
		return "GeometryStreams"
	case CapabilityStorageImageReadWithoutFormat:
		return "StorageImageReadWithoutFormat"
	case CapabilityStorageImageWriteWithoutFormat:
		return "StorageImageWriteWithoutFormat"
	case CapabilityMultiViewport:
		return "MultiViewport"
	case CapabilitySubgroupDispatch:
		return "SubgroupDispatch"
	case CapabilityNamedBarrier:
		return "NamedBarrier"
	case CapabilityPipeStorage:
		return "PipeStorage"
	case CapabilityGroupNonUniform:
		return "GroupNonUniform"
	case CapabilityGroupNonUniformVote:
		return "GroupNonUniformVote"
	case CapabilityGroupNonUniformArithmetic:
		return "GroupNonUniformArithmetic"
	case CapabilityGroupNonUniformBallot:
		return "GroupNonUniformBallot"
	case CapabilityGroupNonUniformShuffle:
		return "GroupNonUniformShuffle"
	case CapabilityGroupNonUniformShuffleRelative:
		return "GroupNonUniformShuffleRelative"
	case CapabilityGroupNonUniformClustered:
		return "GroupNonUniformClustered"
	case CapabilityGroupNonUniformQuad:
		return "GroupNonUniformQuad"
	case CapabilityShaderLayer:
		return "ShaderLayer"
	case CapabilityShaderViewportIndex:
		return "ShaderViewportIndex"
	case CapabilityUniformDecoration:
		return "UniformDecoration"
	default:
		panic("unknown capability")
	}
}

// AddressingModel specifies the addressing model used by the module.
// Used by OpMemoryModel.
type AddressingModel int

const (
	AddressingModelLogical                 AddressingModel = 0
	AddressingModelPhysical32              AddressingModel = 1
	AddressingModelPhysical64              AddressingModel = 2
	AddressingModelPhysicalStorageBuffer64 AddressingModel = 5348
)

func (a AddressingModel) String() string {
	switch a {
	case AddressingModelLogical:
		return "Logical"
	case AddressingModelPhysical32:
		return "Physical32"
	case AddressingModelPhysical64:
		return "Physical64"
	case AddressingModelPhysicalStorageBuffer64:
		return "PhysicalStorageBuffer64"
	default:
		return fmt.Sprintf("AddressingModel(%d)", a)
	}
}

// MemoryModel specifies the memory model used by the module.
// Used by OpMemoryModel.
type MemoryModel int

const (
	MemoryModelSimple  MemoryModel = 0
	MemoryModelGLSL450 MemoryModel = 1
	MemoryModelOpenCL  MemoryModel = 2
	MemoryModelVulkan  MemoryModel = 3
)

func (m MemoryModel) String() string {
	switch m {
	case MemoryModelSimple:
		return "Simple"
	case MemoryModelGLSL450:
		return "GLSL450"
	case MemoryModelOpenCL:
		return "OpenCL"
	case MemoryModelVulkan:
		return "Vulkan"
	default:
		return fmt.Sprintf("MemoryModel(%d)", m)
	}
}

// Module represents the top-level container for IR objects in the compilation unit.
// It manages ID generation and tracks all objects by their IDs.
type Module struct {
	idGenerator     int
	objectsByID     map[ID]Object
	typesByKey      map[string]Type
	capabilities    []Capability
	AddressingModel AddressingModel
	MemoryModel     MemoryModel
}

// ID is a unique identifier for IR objects within a Module.
type ID int

type Object interface {
	ID() ID
	Name() string
}

func NewModule(addressingModel AddressingModel, memoryModel MemoryModel) *Module {
	return &Module{
		idGenerator:     0,
		objectsByID:     make(map[ID]Object),
		typesByKey:      make(map[string]Type),
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
	OpNone         = 0
	OpMemoryModel  = 14
	OpCapability   = 17
	OpTypeVoid     = 19
	OpTypeBool     = 20
	OpTypeInt      = 21
	OpTypePointer  = 32
	OpTypeFunction = 33
	OpReturn       = 253
	OpReturnValue  = 254
)

func (o Op) String() string {
	switch o {
	case OpNone:
		return "OpNone"
	case OpMemoryModel:
		return "OpMemoryModel"
	case OpCapability:
		return "OpCapability"
	case OpTypeVoid:
		return "OpTypeVoid"
	case OpTypeBool:
		return "OpTypeBool"
	case OpTypeInt:
		return "OpTypeInt"
	case OpTypePointer:
		return "OpTypePointer"
	case OpTypeFunction:
		return "OpTypeFunction"
	case OpReturn:
		return "OpReturn"
	case OpReturnValue:
		return "OpReturnValue"
	default:
		return fmt.Sprintf("Op(%d)", o)
	}
}

// Instr represents an instruction in a basic block.
type Instr interface {
	Op() Op
}

type ReturnInstr struct{}

func (ReturnInstr) Op() Op { return OpReturn }

type ReturnValueInstr struct{ Value ID }

func (ReturnValueInstr) Op() Op { return OpReturnValue }

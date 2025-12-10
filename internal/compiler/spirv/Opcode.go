package spirv

import "strings"

// Opcode represents a SPIR-V instruction opcode.
type Opcode int

const (
	OpNone                 Opcode = 0
	OpMemoryModel          Opcode = 14
	OpCapability           Opcode = 17
	OpTypeVoid             Opcode = 19
	OpTypeBool             Opcode = 20
	OpTypeInt              Opcode = 21
	OpTypeFloat            Opcode = 22
	OpTypePointer          Opcode = 32
	OpTypeFunction         Opcode = 33
	OpConstantTrue         Opcode = 41
	OpConstantFalse        Opcode = 42
	OpConstant             Opcode = 43
	OpFunction             Opcode = 54
	OpFunctionParameter    Opcode = 55
	OpFunctionEnd          Opcode = 56
	OpFunctionCall         Opcode = 57
	OpVariable             Opcode = 59
	OpLoad                 Opcode = 61
	OpStore                Opcode = 62
	OpSNegate              Opcode = 126
	OpFNegate              Opcode = 127
	OpIAdd                 Opcode = 128
	OpFAdd                 Opcode = 129
	OpISub                 Opcode = 130
	OpFSub                 Opcode = 131
	OpIMul                 Opcode = 132
	OpFMul                 Opcode = 133
	OpUDiv                 Opcode = 134
	OpSDiv                 Opcode = 135
	OpFDiv                 Opcode = 136
	OpUMod                 Opcode = 137
	OpSRem                 Opcode = 139
	OpFRem                 Opcode = 141
	OpLogicalEqual         Opcode = 164
	OpLogicalNotEqual      Opcode = 165
	OpLogicalOr            Opcode = 166
	OpLogicalAnd           Opcode = 167
	OpLogicalNot           Opcode = 168
	OpIEqual               Opcode = 170
	OpINotEqual            Opcode = 171
	OpUGreaterThan         Opcode = 172
	OpSGreaterThan         Opcode = 173
	OpUGreaterThanEqual    Opcode = 174
	OpSGreaterThanEqual    Opcode = 175
	OpULessThan            Opcode = 176
	OpSLessThan            Opcode = 177
	OpULessThanEqual       Opcode = 178
	OpSLessThanEqual       Opcode = 179
	OpFOrdEqual            Opcode = 180
	OpFOrdNotEqual         Opcode = 182
	OpFOrdLessThan         Opcode = 184
	OpFOrdGreaterThan      Opcode = 186
	OpFOrdLessThanEqual    Opcode = 188
	OpFOrdGreaterThanEqual Opcode = 190
	OpShiftRightLogical    Opcode = 194
	OpShiftRightArithmetic Opcode = 195
	OpShiftLeftLogical     Opcode = 196
	OpBitwiseOr            Opcode = 197
	OpBitwiseXor           Opcode = 198
	OpBitwiseAnd           Opcode = 199
	OpNot                  Opcode = 200
	OpLabel                Opcode = 248
	OpReturn               Opcode = 253
	OpReturnValue          Opcode = 254
	OpUnreachable          Opcode = 255
)

func (op Opcode) String() string {
	switch op {
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
	case OpTypeFloat:
		return "OpTypeFloat"
	case OpTypePointer:
		return "OpTypePointer"
	case OpTypeFunction:
		return "OpTypeFunction"
	case OpConstantTrue:
		return "OpConstantTrue"
	case OpConstantFalse:
		return "OpConstantFalse"
	case OpConstant:
		return "OpConstant"
	case OpFunction:
		return "OpFunction"
	case OpFunctionParameter:
		return "OpFunctionParameter"
	case OpFunctionEnd:
		return "OpFunctionEnd"
	case OpFunctionCall:
		return "OpFunctionCall"
	case OpVariable:
		return "OpVariable"
	case OpLoad:
		return "OpLoad"
	case OpStore:
		return "OpStore"
	case OpSNegate:
		return "OpSNegate"
	case OpFNegate:
		return "OpFNegate"
	case OpIAdd:
		return "OpIAdd"
	case OpFAdd:
		return "OpFAdd"
	case OpISub:
		return "OpISub"
	case OpFSub:
		return "OpFSub"
	case OpIMul:
		return "OpIMul"
	case OpFMul:
		return "OpFMul"
	case OpUDiv:
		return "OpUDiv"
	case OpSDiv:
		return "OpSDiv"
	case OpFDiv:
		return "OpFDiv"
	case OpUMod:
		return "OpUMod"
	case OpSRem:
		return "OpSRem"
	case OpFRem:
		return "OpFRem"
	case OpLogicalEqual:
		return "OpLogicalEqual"
	case OpLogicalNotEqual:
		return "OpLogicalNotEqual"
	case OpLogicalOr:
		return "OpLogicalOr"
	case OpLogicalAnd:
		return "OpLogicalAnd"
	case OpLogicalNot:
		return "OpLogicalNot"
	case OpIEqual:
		return "OpIEqual"
	case OpINotEqual:
		return "OpINotEqual"
	case OpUGreaterThan:
		return "OpUGreaterThan"
	case OpSGreaterThan:
		return "OpSGreaterThan"
	case OpUGreaterThanEqual:
		return "OpUGreaterThanEqual"
	case OpSGreaterThanEqual:
		return "OpSGreaterThanEqual"
	case OpULessThan:
		return "OpULessThan"
	case OpSLessThan:
		return "OpSLessThan"
	case OpULessThanEqual:
		return "OpULessThanEqual"
	case OpSLessThanEqual:
		return "OpSLessThanEqual"
	case OpFOrdEqual:
		return "OpFOrdEqual"
	case OpFOrdNotEqual:
		return "OpFOrdNotEqual"
	case OpFOrdLessThan:
		return "OpFOrdLessThan"
	case OpFOrdGreaterThan:
		return "OpFOrdGreaterThan"
	case OpFOrdLessThanEqual:
		return "OpFOrdLessThanEqual"
	case OpFOrdGreaterThanEqual:
		return "OpFOrdGreaterThanEqual"
	case OpShiftRightLogical:
		return "OpShiftRightLogical"
	case OpShiftRightArithmetic:
		return "OpShiftRightArithmetic"
	case OpBitwiseOr:
		return "OpBitwiseOr"
	case OpBitwiseXor:
		return "OpBitwiseXor"
	case OpBitwiseAnd:
		return "OpBitwiseAnd"
	case OpNot:
		return "OpNot"
	case OpShiftLeftLogical:
		return "OpShiftLeftLogical"
	case OpLabel:
		return "OpLabel"
	case OpReturn:
		return "OpReturn"
	case OpReturnValue:
		return "OpReturnValue"
	case OpUnreachable:
		return "OpUnreachable"
	default:
		panic("unknown opcode")
	}
}

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
		panic("unknown addressing model")
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
		panic("unknown memory model")
	}
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

func (s StorageClass) String() string {
	switch s {
	case StorageClassUniformConstant:
		return "UniformConstant"
	case StorageClassInput:
		return "Input"
	case StorageClassUniform:
		return "Uniform"
	case StorageClassOutput:
		return "Output"
	case StorageClassWorkgroup:
		return "Workgroup"
	case StorageClassCrossWorkgroup:
		return "CrossWorkgroup"
	case StorageClassPrivate:
		return "Private"
	case StorageClassFunction:
		return "Function"
	case StorageClassGeneric:
		return "Generic"
	case StorageClassPushConstant:
		return "PushConstant"
	case StorageClassAtomicCounter:
		return "AtomicCounter"
	case StorageClassImage:
		return "Image"
	case StorageClassStorageBuffer:
		return "StorageBuffer"
	default:
		panic("unknown storage class")
	}
}

type FunctionControl int

const (
	FunctionControlNone FunctionControl = 0
	// Performance hint. Strong request to inline the function.
	FunctionControlInline FunctionControl = 1
	// Performance hint. Strong request to not inline the function.
	FunctionControlDontInline FunctionControl = 2
	// Compiler can assume this function has no side effect, but might read global memory or
	// read through dereferenced function parameters.
	// Always computes the same result when called with the same argument values and
	// the same global state.
	FunctionControlPure FunctionControl = 4
	// Compiler assumes this function has no side effects, and does not access global
	// memory or dereference function parameters. Always computes the same result for the
	// same argument values.
	FunctionControlConst FunctionControl = 8
)

func (v FunctionControl) String() string {
	if v == FunctionControlNone {
		return "None"
	}

	var flags []string
	if v&FunctionControlInline != 0 {
		flags = append(flags, "Inline")
	}
	if v&FunctionControlDontInline != 0 {
		flags = append(flags, "DontInline")
	}
	if v&FunctionControlPure != 0 {
		flags = append(flags, "Pure")
	}
	if v&FunctionControlConst != 0 {
		flags = append(flags, "Const")
	}
	return strings.Join(flags, "|")
}

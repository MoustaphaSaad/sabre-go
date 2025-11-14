package spirv

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sort"
)

type Word = uint32

type BinaryPrinter struct {
	out    io.Writer
	module *Module
}

func NewBinaryPrinter(out io.Writer, module *Module) *BinaryPrinter {
	return &BinaryPrinter{
		out:    out,
		module: module,
	}
}

func (bp *BinaryPrinter) Emit() {
	bp.emitHeader()
	bp.emitCapabilities()
	bp.emitMemoryModel()

	objs := make([]Object, 0, len(bp.module.objectsByID))
	for _, obj := range bp.module.objectsByID {
		objs = append(objs, obj)
	}
	sort.Slice(objs, func(i, j int) bool { return objs[i].ID() < objs[j].ID() })

	for _, obj := range objs {
		if _, isType := obj.(Type); isType {
			bp.emitObject(obj)
		}
	}

	for _, obj := range objs {
		switch obj.(type) {
		case Constant:
			bp.emitObject(obj)
		}
	}

	for _, obj := range objs {
		if _, isFunction := obj.(*Function); isFunction {
			bp.emitObject(obj)
		}
	}
}

func (bp *BinaryPrinter) emitObject(obj Object) {
	switch v := obj.(type) {
	case *Function:
		bp.emitFunction(v)
	case Type:
		bp.emitType(v)
	case Constant:
		bp.emitConstant(v)
	}
}

func (bp *BinaryPrinter) emitFunction(f *Function) {
	bp.emitOp(
		Word(OpFunction),
		Word(f.Type.ReturnType.ID()),
		Word(f.ID()),
		Word(FunctionControlNone),
		Word(f.Type.ID()),
	)
	for _, block := range f.Blocks {
		bp.emitBlock(block)
	}
	bp.emitOp(Word(OpFunctionEnd))
}

func (bp *BinaryPrinter) emitBlock(block *Block) {
	bp.emitOp(Word(OpLabel), Word(block.ID()))
	for _, inst := range block.Instructions {
		bp.emitInstruction(inst)
	}
}

func (bp *BinaryPrinter) emitInstruction(inst Instruction) {
	switch i := inst.(type) {
	case *ConstantTrueInstruction:
		bp.emitOp(Word(OpConstantTrue), Word(i.ResultType), Word(i.ResultID))
	case *ConstantFalseInstruction:
		bp.emitOp(Word(OpConstantFalse), Word(i.ResultType), Word(i.ResultID))
	case *ReturnInstruction:
		bp.emitOp(Word(OpReturn))
	case *ReturnValueInstruction:
		bp.emitOp(Word(OpReturnValue), Word(i.Value))
	case *SNegateInstruction:
		bp.emitOp(Word(OpSNegate), Word(i.ResultType), Word(i.ResultID), Word(i.Operand))
	case *FNegateInstruction:
		bp.emitOp(Word(OpFNegate), Word(i.ResultType), Word(i.ResultID), Word(i.Operand))
	case *LogicalOrInstruction:
		bp.emitOp(Word(OpLogicalOr), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *LogicalAndInstruction:
		bp.emitOp(Word(OpLogicalAnd), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *LogicalNotInstruction:
		bp.emitOp(Word(OpLogicalNot), Word(i.ResultType), Word(i.ResultID), Word(i.Operand))
	case *UGreaterThanInstruction:
		bp.emitOp(Word(OpUGreaterThan), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *SGreaterThanInstruction:
		bp.emitOp(Word(OpSGreaterThan), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *ULessThanInstruction:
		bp.emitOp(Word(OpULessThan), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *SLessThanInstruction:
		bp.emitOp(Word(OpSLessThan), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *ULessThanEqualInstruction:
		bp.emitOp(Word(OpULessThanEqual), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *SLessThanEqualInstruction:
		bp.emitOp(Word(OpSLessThanEqual), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *FOrdLessThanInstruction:
		bp.emitOp(Word(OpFOrdLessThan), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *FOrdGreaterThanInstruction:
		bp.emitOp(Word(OpFOrdGreaterThan), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *FOrdLessThanEqualInstruction:
		bp.emitOp(Word(OpFOrdLessThanEqual), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *UGreaterThanEqualInstruction:
		bp.emitOp(Word(OpUGreaterThanEqual), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *SGreaterThanEqualInstruction:
		bp.emitOp(Word(OpSGreaterThanEqual), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *FOrdGreaterThanEqualInstruction:
		bp.emitOp(Word(OpFOrdGreaterThanEqual), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *IEqualInstruction:
		bp.emitOp(Word(OpIEqual), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *FOrdEqualInstruction:
		bp.emitOp(Word(OpFOrdEqual), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *INotEqualInstruction:
		bp.emitOp(Word(OpINotEqual), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *FOrdNotEqualInstruction:
		bp.emitOp(Word(OpFOrdNotEqual), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *IAddInstruction:
		bp.emitOp(Word(OpIAdd), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *FAddInstruction:
		bp.emitOp(Word(OpFAdd), Word(i.ResultType), Word(i.ResultID), Word(i.Operand1), Word(i.Operand2))
	case *NotInstruction:
		bp.emitOp(Word(OpNot), Word(i.ResultType), Word(i.ResultID), Word(i.Operand))
	default:
		panic("unsupported instruction")
	}
}

func (bp *BinaryPrinter) emitType(abstractType Type) {
	switch t := abstractType.(type) {
	case *VoidType:
		bp.emitVoidType(t)
	case *BoolType:
		bp.emitBoolType(t)
	case *IntType:
		bp.emitIntType(t)
	case *FloatType:
		bp.emitFloatType(t)
	case *PtrType:
		bp.emitPtrType(t)
	case *FuncType:
		bp.emitFuncType(t)
	default:
		panic("unsupported type")
	}
}

func (bp *BinaryPrinter) emitConstant(constant Constant) {
	switch c := constant.(type) {
	case *BoolConstant:
		bp.emitBoolConstant(c)
	case *IntConstant:
		bp.emitIntConstant(c)
	case *FloatConstant:
		bp.emitFloatConstant(c)
	default:
		panic(fmt.Sprintf("unsupported constant: %T", c))
	}
}

func (bp *BinaryPrinter) emitBoolConstant(c *BoolConstant) {
	if c.Value {
		bp.emitOp(Word(OpConstantTrue), Word(c.Type.ID()), Word(c.ID()))
	} else {
		bp.emitOp(Word(OpConstantFalse), Word(c.Type.ID()), Word(c.ID()))
	}
}

func (bp *BinaryPrinter) emitIntConstant(c *IntConstant) {
	bp.emitOp(Word(OpConstant), Word(c.Type.ID()), Word(c.ID()), Word(uint32(c.Value)))
}

func (bp *BinaryPrinter) emitFloatConstant(c *FloatConstant) {
	switch c.Type.BitWidth {
	case 32:
		bp.emitOp(Word(OpConstant), Word(c.Type.ID()), Word(c.ID()), Word(math.Float32bits(float32(c.Value))))
	case 64:
		bits := math.Float64bits(c.Value)
		lowWord := Word(bits & 0xFFFFFFFF)
		highWord := Word(bits >> 32)
		bp.emitOp(Word(OpConstant), Word(c.Type.ID()), Word(c.ID()), lowWord, highWord)
	default:
		panic("unsupported float bit width")
	}
}

func (bp *BinaryPrinter) emitVoidType(t *VoidType) {
	bp.emitOp(Word(OpTypeVoid), Word(t.ID()))
}

func (bp *BinaryPrinter) emitBoolType(t *BoolType) {
	bp.emitOp(Word(OpTypeBool), Word(t.ID()))
}

func (bp *BinaryPrinter) emitIntType(t *IntType) {
	bp.emitOp(Word(OpTypeInt), Word(t.ID()), Word(t.BitWidth), Word(boolToWord(t.IsSigned)))
}

func (bp *BinaryPrinter) emitFloatType(t *FloatType) {
	bp.emitOp(Word(OpTypeFloat), Word(t.ID()), Word(t.BitWidth))
}

func (bp *BinaryPrinter) emitPtrType(t *PtrType) {
	bp.emitOp(Word(OpTypePointer), Word(t.ID()), Word(t.StorageClass), Word(t.To.ID()))
}

func (bp *BinaryPrinter) emitFuncType(t *FuncType) {
	args := make([]Word, 0, len(t.ArgTypes)+2)
	args = append(args, Word(t.ID()))
	args = append(args, Word(t.ReturnType.ID()))
	for _, argTy := range t.ArgTypes {
		args = append(args, Word(argTy.ID()))
	}
	bp.emitOp(Word(OpTypeFunction), args...)
}

func (bp *BinaryPrinter) emitMemoryModel() {
	bp.emitOp(Word(OpMemoryModel), Word(bp.module.AddressingModel), Word(bp.module.MemoryModel))
}

func (bp *BinaryPrinter) emitCapabilities() {
	for _, c := range bp.module.Capabilities() {
		bp.emitOp(Word(OpCapability), Word(c))
	}
}

func (bp *BinaryPrinter) emitHeader() {
	// SPIR-V Magic
	bp.emitMagicNumber()
	// Version 1.3, I chose 1.3 arbitrarily.
	bp.emitVersion(1, 3)
	// Generator's magic number (arbitrary)
	// Generator’s magic number. It is associated with the tool that generated the module.
	// Its value does not affect any semantics, and is allowed to be 0. Using a non-0 value is encouraged,
	// and can be registered with Khronos at https://github.com/KhronosGroup/SPIRV-Headers.
	bp.emitWord(0)
	// Bound: All <id>s in the module are guaranteed to be less than this number.
	// The value of the bound is equal to the maximum <id> assigned plus one.
	// The minimum valid value for the bound is 1.
	bp.emitWord(Word(bp.module.idGenerator + 1))
	// Reserved for instruction schema; must be 0.
	bp.emitWord(0)
}

func (bp *BinaryPrinter) emitVersion(major, minor uint8) {
	version := Word(major)<<16 | Word(minor)<<8
	bp.emitWord(version)
}

func (bp *BinaryPrinter) emitMagicNumber() {
	const magicNumber Word = 0x07230203
	bp.emitWord(magicNumber)
}

func (bp *BinaryPrinter) emitOp(opcode Word, operands ...Word) {
	wordCount := Word(1 + len(operands))
	bp.emitWord((wordCount << 16) | opcode)
	for _, operand := range operands {
		bp.emitWord(operand)
	}
}

func (bp *BinaryPrinter) emitWord(value Word) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	bp.out.Write(buf[:])
}

func boolToWord(b bool) Word {
	if b {
		return 1
	}
	return 0
}

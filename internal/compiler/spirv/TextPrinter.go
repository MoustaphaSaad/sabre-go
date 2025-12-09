package spirv

import (
	"fmt"
	"io"
)

type TextPrinter struct {
	out    io.Writer
	module *Module
}

func NewTextPrinter(out io.Writer, module *Module) *TextPrinter {
	return &TextPrinter{
		out:    out,
		module: module,
	}
}

func (tp *TextPrinter) Emit() {
	tp.emitCapabilities()
	tp.emitMemoryModel()

	for _, obj := range tp.module.Objects {
		if _, isType := obj.(Type); isType {
			tp.emitObject(obj)
		}
	}

	for _, obj := range tp.module.Objects {
		switch obj.(type) {
		case ConstantValue:
			tp.emitObject(obj)
		}
	}

	for _, obj := range tp.module.Objects {
		if _, isFunction := obj.(*Function); isFunction {
			tp.emitObject(obj)
		}
	}
}

func (tp *TextPrinter) emitCapabilities() {
	for _, c := range tp.module.Capabilities() {
		fmt.Fprintf(tp.out, "OpCapability %s\n", c)
	}
}

func (tp *TextPrinter) emitMemoryModel() {
	fmt.Fprintf(tp.out, "OpMemoryModel %s %s\n", tp.module.AddressingModel, tp.module.MemoryModel)
}

func (tp *TextPrinter) emitObject(obj Object) {
	switch v := obj.(type) {
	case *Function:
		tp.emitFunction(v)
	case Type:
		tp.emitType(v)
	case ConstantValue:
		tp.emitConstant(v)
	}
}

func (tp *TextPrinter) emitFunction(f *Function) {
	tp.emitWithObject(
		f,
		OpFunction,
		tp.nameOf(f.Type.ReturnType),
		FunctionControlNone,
		tp.nameOf(f.Type),
	)
	for _, p := range f.Params {
		tp.emitWithObject(p, OpFunctionParameter, tp.nameOf(p.Type))
	}
	for _, bb := range f.Blocks {
		tp.emitBlock(bb)
	}
	tp.emit(OpFunctionEnd)
}

func (tp *TextPrinter) emitBlock(bb *Block) {
	tp.emitWithObject(bb, OpLabel)
	for _, inst := range bb.Instructions {
		tp.emitInstruction(inst)
	}
}

func (tp *TextPrinter) emitInstruction(inst Instruction) {
	switch i := inst.(type) {
	case *ReturnInstruction:
		tp.emit(OpReturn)
	case *ReturnValueInstruction:
		tp.emit(OpReturnValue, tp.nameOfByID(i.Value))
	case *SNegateInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpSNegate, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand))
	case *FNegateInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpFNegate, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand))
	case *LogicalOrInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpLogicalOr, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *LogicalAndInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpLogicalAnd, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *LogicalNotInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpLogicalNot, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand))
	case *LogicalEqualInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpLogicalEqual, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *LogicalNotEqualInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpLogicalNotEqual, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *UGreaterThanInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpUGreaterThan, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *SGreaterThanInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpSGreaterThan, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *ULessThanInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpULessThan, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *SLessThanInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpSLessThan, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *ULessThanEqualInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpULessThanEqual, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *SLessThanEqualInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpSLessThanEqual, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *FOrdLessThanInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpFOrdLessThan, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *FOrdGreaterThanInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpFOrdGreaterThan, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *FOrdLessThanEqualInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpFOrdLessThanEqual, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *UGreaterThanEqualInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpUGreaterThanEqual, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *SGreaterThanEqualInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpSGreaterThanEqual, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *FOrdGreaterThanEqualInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpFOrdGreaterThanEqual, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *IEqualInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpIEqual, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *FOrdEqualInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpFOrdEqual, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *INotEqualInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpINotEqual, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *FOrdNotEqualInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpFOrdNotEqual, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *IAddInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpIAdd, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *FAddInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpFAdd, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *ISubInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpISub, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *FSubInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpFSub, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *IMulInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpIMul, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *FMulInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpFMul, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *UDivInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpUDiv, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *SDivInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpSDiv, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *FDivInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpFDiv, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *UModInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpUMod, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *SRemInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpSRem, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *FRemInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpFRem, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *BitwiseXorInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpBitwiseXor, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *BitwiseOrInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpBitwiseOr, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *BitwiseAndInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpBitwiseAnd, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand1), tp.nameOfByID(i.Operand2))
	case *NotInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpNot, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Operand))
	case *ShiftLeftLogicalInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpShiftLeftLogical, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Base), tp.nameOfByID(i.Shift))
	case *ShiftRightLogicalInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpShiftRightLogical, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Base), tp.nameOfByID(i.Shift))
	case *ShiftRightArithmeticInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpShiftRightArithmetic, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Base), tp.nameOfByID(i.Shift))
	case *FunctionCallInstruction:
		args := make([]any, 0, len(i.Args)+2)
		args = append(args, tp.nameOfByID(i.ResultType))
		args = append(args, tp.nameOfByID(i.FunctionID))
		for _, arg := range i.Args {
			args = append(args, tp.nameOfByID(arg))
		}
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpFunctionCall, args...)
	case *VariableInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		if i.Initializer != 0 {
			tp.emitWithObject(resultObj, OpVariable, tp.nameOfByID(i.ResultType), i.StorageClass, tp.nameOfByID(i.Initializer))
		} else {
			tp.emitWithObject(resultObj, OpVariable, tp.nameOfByID(i.ResultType), i.StorageClass)
		}
	case *LoadInstruction:
		resultObj := tp.module.GetObject(i.ResultID)
		tp.emitWithObject(resultObj, OpLoad, tp.nameOfByID(i.ResultType), tp.nameOfByID(i.Pointer))
	case *StoreInstruction:
		tp.emit(OpStore, tp.nameOfByID(i.Pointer), tp.nameOfByID(i.Object))
	default:
		panic(fmt.Sprintf("unsupported instruction: %T", inst))
	}
}

func (tp *TextPrinter) emitType(abstractType Type) {
	switch t := abstractType.(type) {
	case *VoidType:
		tp.emitVoidType(t)
	case *BoolType:
		tp.emitBoolType(t)
	case *IntType:
		tp.emitIntType(t)
	case *FloatType:
		tp.emitFloatType(t)
	case *PtrType:
		tp.emitPtrType(t)
	case *FuncType:
		tp.emitFuncType(t)
	default:
		panic(fmt.Sprintf("unsupported type: %T", abstractType))
	}
}

func (tp *TextPrinter) emitConstant(constant ConstantValue) {
	switch c := constant.(type) {
	case *BoolConstant:
		tp.emitBoolConstant(c)
	case *IntConstant:
		tp.emitIntConstant(c)
	case *FloatConstant:
		tp.emitFloatConstant(c)
	default:
		panic(fmt.Sprintf("unsupported constant: %T", c))
	}
}

func (tp *TextPrinter) emitBoolConstant(c *BoolConstant) {
	if c.Value {
		tp.emitWithObject(c, OpConstantTrue, tp.nameOf(c.Type))
	} else {
		tp.emitWithObject(c, OpConstantFalse, tp.nameOf(c.Type))
	}
}

func (tp *TextPrinter) emitIntConstant(c *IntConstant) {
	tp.emitWithObject(c, OpConstant, tp.nameOf(c.Type), c.Value)
}

func (tp *TextPrinter) emitFloatConstant(c *FloatConstant) {
	tp.emitWithObject(c, OpConstant, tp.nameOf(c.Type), c.Value)
}

func (tp *TextPrinter) emitVoidType(t *VoidType) {
	tp.emitWithObject(t, OpTypeVoid)
}

func (tp *TextPrinter) emitBoolType(t *BoolType) {
	tp.emitWithObject(t, OpTypeBool)
}

func (tp *TextPrinter) emitIntType(t *IntType) {
	tp.emitWithObject(t, OpTypeInt, t.BitWidth, boolToInt(t.IsSigned))
}

func (tp *TextPrinter) emitFloatType(t *FloatType) {
	tp.emitWithObject(t, OpTypeFloat, t.BitWidth)
}

func (tp *TextPrinter) emitPtrType(t *PtrType) {
	tp.emitWithObject(t, OpTypePointer, t.StorageClass, tp.nameOf(t.To))
}

func (tp *TextPrinter) emitFuncType(t *FuncType) {
	args := make([]any, 0, len(t.ArgTypes)+1)
	args = append(args, tp.nameOf(t.ReturnType))
	for _, argTy := range t.ArgTypes {
		args = append(args, tp.nameOf(argTy))
	}
	tp.emitWithObject(t, OpTypeFunction, args...)
}

func (tp *TextPrinter) emitWithObject(obj Object, op Opcode, args ...any) {
	fmt.Fprintf(tp.out, "%s = %s", tp.nameOf(obj), op)
	for _, arg := range args {
		fmt.Fprintf(tp.out, " %v", arg)
	}
	fmt.Fprintln(tp.out)
}

func (tp *TextPrinter) emit(op Opcode, args ...any) {
	fmt.Fprintf(tp.out, "%s", op)
	for _, arg := range args {
		fmt.Fprintf(tp.out, " %v", arg)
	}
	fmt.Fprintln(tp.out)
}

func (tp *TextPrinter) nameOfByID(id ID) string {
	if obj := tp.module.GetObject(id); obj != nil {
		return tp.nameOf(obj)
	}
	panic(fmt.Sprintf("object with ID %v does not exist", id))
}

func (tp *TextPrinter) nameOf(obj Object) string {
	kind := ""
	switch obj.(type) {
	case *Function:
		kind = "func"
	case *Block:
		kind = "block"
	case Type:
		kind = "type"
	case *BoolConstant:
		kind = "const"
	}
	if len(kind) == 0 {
		return fmt.Sprintf("%%%s_%d", obj.Name(), obj.ID())
	} else {
		return fmt.Sprintf("%%%s_%s_%d", kind, obj.Name(), obj.ID())
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

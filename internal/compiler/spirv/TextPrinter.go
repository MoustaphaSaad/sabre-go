package spirv

import (
	"fmt"
	"io"
	"sort"
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

	objs := make([]Object, 0, len(tp.module.objectsByID))
	for _, obj := range tp.module.objectsByID {
		objs = append(objs, obj)
	}
	sort.Slice(objs, func(i, j int) bool { return objs[i].ID() < objs[j].ID() })

	for _, obj := range objs {
		tp.emitObject(obj)
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
	case ReturnInstruction:
		tp.emit(OpReturn)
	case *ReturnValueInstruction:
		tp.emit(OpReturnValue, tp.nameOfByID(i.Value))
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
	case *PtrType:
		tp.emitPtrType(t)
	case *FuncType:
		tp.emitFuncType(t)
	default:
		panic("unsupported type")
	}
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

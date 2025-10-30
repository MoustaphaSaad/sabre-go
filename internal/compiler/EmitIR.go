package compiler

import (
	"github.com/MoustaphaSaad/sabre-go/internal/compiler/spirv"
)

type IREmitter struct {
	unit   *Unit
	module *spirv.Module
}

func NewIREmitter(u *Unit) *IREmitter {
	return &IREmitter{
		unit: u,
		// we set the addressing and memory model to default values for now
		module: spirv.NewModule(spirv.AddressingModelLogical, spirv.MemoryModelGLSL450),
	}
}

func (ir *IREmitter) Emit() *spirv.Module {
	// we add this hardcoded capabilities for now
	ir.module.AddCapability(spirv.CapabilityShader)
	ir.module.AddCapability(spirv.CapabilityLinkage)

	for _, sym := range ir.unit.semanticInfo.ReachableSymbols {
		ir.emitSymbol(sym)
	}
	return ir.module
}

func (ir *IREmitter) emitSymbol(sym Symbol) {
	switch s := sym.(type) {
	case *FuncSymbol:
		ir.emitFunc(s)
	}
}

func (ir *IREmitter) emitFunc(sym *FuncSymbol) {
	funcType := ir.unit.semanticInfo.TypeOf(sym).Type.(*FuncType)
	spirvFuncType := ir.emitType(funcType).(*spirv.FuncType)
	spirvFunction := ir.module.NewFunction(sym.Name(), spirvFuncType)
	spirvBlock := spirvFunction.NewBlock(sym.Name())
	spirvBlock.Push(&spirv.ReturnInstruction{})
}

func (ir *IREmitter) emitType(Type Type) spirv.Type {
	switch t := Type.(type) {
	case *VoidType:
		return ir.module.InternVoid()
	case *FuncType:
		var spirvReturnType spirv.Type
		if len(t.ReturnTypes) > 0 {
			// TODO: Handle multiple return types
			spirvReturnType = ir.emitType(t.ReturnTypes[0])
		} else {
			spirvReturnType = ir.module.InternVoid()
		}

		var parameterTypes []spirv.Type
		for _, paramType := range t.ParameterTypes {
			parameterTypes = append(parameterTypes, ir.emitType(paramType))
		}

		return ir.module.InternFunc(spirvReturnType, parameterTypes)
	default:
		panic("unexpected type")
	}
}

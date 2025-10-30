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

	funcDecl := sym.Decl().(*FuncDecl)
	if funcDecl.Body == nil {
		return
	}

	spirvBlock := spirvFunction.NewBlock(sym.Name())
	if len(funcDecl.Body.Stmts) == 0 {
		spirvBlock.Push(&spirv.ReturnInstruction{})
		return
	}

	for _, stmt := range funcDecl.Body.Stmts {
		ir.emitStatement(stmt, spirvBlock)
	}
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

func (ir *IREmitter) emitStatement(stmt Stmt, block *spirv.Block) {
	switch s := stmt.(type) {
	case *ReturnStmt:
		ir.emitReturnStmt(s, block)
	default:
		panic("unsupported statement")
	}
}

func (ir *IREmitter) emitReturnStmt(s *ReturnStmt, block *spirv.Block) {
	block.Push(&spirv.ReturnInstruction{})
}

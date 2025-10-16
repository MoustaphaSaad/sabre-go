package compiler

import "github.com/MoustaphaSaad/sabre-go/internal/compiler/spirv"

type IREmitter struct {
	unit   *Unit
	module *spirv.Module
}

func NewIREmitter(u *Unit) *IREmitter {
	return &IREmitter{
		unit: u,
		// we set the addressing and memory model to some defaults for now
		module: spirv.NewModule(spirv.AddressingModelLogical, spirv.MemoryModelGLSL450),
	}
}

func (ir *IREmitter) Emit() *spirv.Module {
	// we add this hardcoded capability for now
	ir.module.AddCapability(spirv.CapabilityShader)
	ir.module.AddCapability(spirv.CapabilityLinkage)

	for _, sym := range ir.unit.semanticInfo.ReachableSymbols {
		ir.emitSymbol(sym)
	}
	return ir.module
}

func (ir *IREmitter) emitSymbol(sym Symbol) {
	// To be implemented later
}

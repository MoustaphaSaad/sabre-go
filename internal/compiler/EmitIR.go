package compiler

import "github.com/MoustaphaSaad/sabre-go/internal/compiler/spirv"

type IREmitter struct {
	unit   *Unit
	module *spirv.Module
}

func NewIREmitter(u *Unit) *IREmitter {
	return &IREmitter{
		unit:   u,
		module: spirv.NewModule(),
	}
}

func (ir *IREmitter) Emit() *spirv.Module {
	for _, sym := range ir.unit.semanticInfo.ReachableSymbols {
		ir.emitSymbol(sym)
	}
	return ir.module
}

func (ir *IREmitter) emitSymbol(sym Symbol) {
	panic("Function not implemented yet")
}

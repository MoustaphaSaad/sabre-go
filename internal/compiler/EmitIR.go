package compiler

type IREmitter struct {
	unit   *Unit
	module *Module
}

func NewIREmitter(u *Unit) *IREmitter {
	return &IREmitter{
		unit:   u,
		module: NewModule(),
	}
}

func (ir *IREmitter) Emit() *Module {
	for _, sym := range ir.unit.semanticInfo.ReachableSymbols {
		ir.emitSymbol(sym)
	}
	return ir.module
}

func (ir *IREmitter) emitSymbol(sym Symbol) {
	panic("Function not implemented yet")
}

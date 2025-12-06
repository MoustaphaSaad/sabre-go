package compiler

type ResolveState byte

const (
	ResolveStateUnresolved ResolveState = iota
	ResolveStateResolving
	ResolveStateResolved
)

type Symbol interface {
	aSymbol()
	Scope() *Scope
	SetScope(scope *Scope)
	Name() string
	Decl() Decl
	SourceRange() SourceRange
	ResolveState() ResolveState
	SetResolveState(r ResolveState)
}

type SymbolBase struct {
	SymScope        *Scope
	SymName         string
	SymDecl         Decl
	SymSourceRange  SourceRange
	SymResolveState ResolveState
}

func (sym SymbolBase) Scope() *Scope {
	return sym.SymScope
}
func (sym *SymbolBase) SetScope(scope *Scope) {
	sym.SymScope = scope
}
func (sym SymbolBase) Name() string {
	return sym.SymName
}
func (sym SymbolBase) Decl() Decl {
	return sym.SymDecl
}
func (sym SymbolBase) SourceRange() SourceRange {
	return sym.SymSourceRange
}
func (sym SymbolBase) ResolveState() ResolveState {
	return sym.SymResolveState
}
func (sym *SymbolBase) SetResolveState(r ResolveState) {
	sym.SymResolveState = r
}

type FuncSymbol struct {
	SymbolBase
}

func (FuncSymbol) aSymbol() {}
func NewFuncSymbol(name Token, decl Decl, sourceRange SourceRange) *FuncSymbol {
	return &FuncSymbol{
		SymbolBase: SymbolBase{
			SymScope:       nil,
			SymName:        name.Value(),
			SymDecl:        decl,
			SymSourceRange: sourceRange,
		},
	}
}

type VarSymbol struct {
	SymbolBase
	SpecIndex        int
	ExprIndex        int
	InitTypeAndValue *TypeAndValue
}

func (VarSymbol) aSymbol() {}
func NewVarSymbol(name Token, decl Decl, sourceRange SourceRange, specIndex, exprIndex int, initTAV *TypeAndValue) *VarSymbol {
	return &VarSymbol{
		SymbolBase: SymbolBase{
			SymScope:       nil,
			SymName:        name.Value(),
			SymDecl:        decl,
			SymSourceRange: sourceRange,
		},
		SpecIndex:        specIndex,
		ExprIndex:        exprIndex,
		InitTypeAndValue: initTAV,
	}
}

type ConstSymbol struct {
	SymbolBase
	SpecIndex int
	ExprIndex int
}

func (ConstSymbol) aSymbol() {}
func NewConstSymbol(name Token, decl Decl, sourceRange SourceRange, specIndex, exprIndex int) *ConstSymbol {
	return &ConstSymbol{
		SymbolBase: SymbolBase{
			SymScope:       nil,
			SymName:        name.Value(),
			SymDecl:        decl,
			SymSourceRange: sourceRange,
		},
		SpecIndex: specIndex,
		ExprIndex: exprIndex,
	}
}

type TypeSymbol struct {
	SymbolBase
	TypeExpr TypeExpr
	IsStrong bool
}

func (TypeSymbol) aSymbol() {}
func NewTypeSymbol(name Token, decl Decl, sourceRange SourceRange, typeExpr TypeExpr, isStrong bool) *TypeSymbol {
	return &TypeSymbol{
		SymbolBase: SymbolBase{
			SymScope:       nil,
			SymName:        name.Value(),
			SymDecl:        decl,
			SymSourceRange: sourceRange,
		},
		TypeExpr: typeExpr,
		IsStrong: isStrong,
	}
}

type Scope struct {
	Parent  *Scope
	Name    string
	Symbols []Symbol
	Table   map[string]int
}

func NewScope(parent *Scope, name string) *Scope {
	return &Scope{
		Parent:  parent,
		Name:    name,
		Symbols: make([]Symbol, 0),
		Table:   make(map[string]int),
	}
}

func (s Scope) ShallowFind(name string) Symbol {
	if index, ok := s.Table[name]; ok {
		return s.Symbols[index]
	}
	return nil
}

func (s *Scope) Add(sym Symbol) bool {
	if s.ShallowFind(sym.Name()) != nil {
		return false
	}

	s.Table[sym.Name()] = len(s.Symbols)
	s.Symbols = append(s.Symbols, sym)
	return true
}

func (s Scope) Find(name string) Symbol {
	for it := &s; it != nil; it = it.Parent {
		if sym := it.ShallowFind(name); sym != nil {
			return sym
		}
	}
	return nil
}

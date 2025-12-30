package compiler

import (
	"go/constant"
	"go/token"
	"strconv"
)

type SemanticInfo struct {
	Types              map[any]*TypeAndValue
	Scopes             map[any]*Scope
	SymbolByIdentifier map[*IdentifierExpr]Symbol
	TypeInterner       *TypeInterner
	ReachableSymbols   []Symbol
}

func NewSemanticInfo() *SemanticInfo {
	return &SemanticInfo{
		Types:              make(map[any]*TypeAndValue),
		Scopes:             make(map[any]*Scope),
		SymbolByIdentifier: make(map[*IdentifierExpr]Symbol),
		TypeInterner:       NewTypeInterner(),
		ReachableSymbols:   make([]Symbol, 0),
	}
}

func (info *SemanticInfo) SetTypeOf(e any, t *TypeAndValue) {
	info.Types[e] = t
}

func (info SemanticInfo) TypeOf(e any) *TypeAndValue {
	if t, ok := info.Types[e]; ok {
		return t
	}
	return nil
}

func (info SemanticInfo) ScopeOf(n any) *Scope {
	if s, ok := info.Scopes[n]; ok {
		return s
	}
	return nil
}

func (info *SemanticInfo) createScopeFor(n any, parent *Scope, name string) *Scope {
	if scope, ok := info.Scopes[n]; ok {
		return scope
	}
	scope := NewScope(parent, name)
	info.Scopes[n] = scope
	return scope
}

func (info *SemanticInfo) SetSymbolOfIdentifier(e *IdentifierExpr, s Symbol) {
	info.SymbolByIdentifier[e] = s
}

func (info *SemanticInfo) SymbolOfIdentifier(e *IdentifierExpr) Symbol {
	if symbol, ok := info.SymbolByIdentifier[e]; ok {
		return symbol
	}
	return nil
}

type ResolveStmtProperties struct {
	acceptsBreak       bool
	acceptsContinue    bool
	acceptsFallthrough bool
	isFinalCaseStmt    bool
}

type AddressMode byte

const (
	AddressModeInvalid AddressMode = iota
	AddressModeNoValue
	AddressModeType
	AddressModeConstant
	AddressModeVariable
	AddressModeComputedValue
)

func (a AddressMode) Combine(b AddressMode) AddressMode {
	switch a {
	case AddressModeInvalid, AddressModeNoValue, AddressModeType:
		return a
	case AddressModeConstant:
		switch b {
		case AddressModeConstant:
			return AddressModeConstant
		case AddressModeVariable, AddressModeComputedValue:
			return AddressModeComputedValue
		default:
			return b
		}
	case AddressModeVariable, AddressModeComputedValue:
		switch b {
		case AddressModeConstant, AddressModeVariable, AddressModeComputedValue:
			return AddressModeComputedValue
		default:
			return b
		}
	default:
		panic("unexpected AddressMode")
	}
}

type TypeAndValue struct {
	Mode  AddressMode
	Type  Type
	Value constant.Value
}

func (v TypeAndValue) IsVoid() bool {
	return v.Mode == AddressModeNoValue
}

func (v TypeAndValue) IsType() bool {
	return v.Mode == AddressModeType
}

func (v TypeAndValue) IsValue() bool {
	return v.Mode == AddressModeConstant || v.Mode == AddressModeVariable || v.Mode == AddressModeComputedValue
}

func (v TypeAndValue) IsAddressable() bool {
	return v.Mode == AddressModeVariable
}

func (v TypeAndValue) IsAssignable() bool {
	return v.Mode == AddressModeVariable
}

func convertTokenToConstantToken(op TokenKind) token.Token {
	switch op {
	case TokenLT:
		return token.LSS
	case TokenGT:
		return token.GTR
	case TokenLE:
		return token.LEQ
	case TokenGE:
		return token.GEQ
	case TokenEQ:
		return token.EQL
	case TokenNE:
		return token.NEQ
	case TokenAdd:
		return token.ADD
	case TokenSub:
		return token.SUB
	case TokenMul:
		return token.MUL
	case TokenDiv:
		return token.QUO
	case TokenMod:
		return token.REM
	case TokenLOr:
		return token.LOR
	case TokenLAnd:
		return token.LAND
	case TokenXor:
		return token.XOR
	case TokenOr:
		return token.OR
	case TokenNot:
		return token.NOT
	case TokenAnd:
		return token.AND
	case TokenAndNot:
		return token.AND_NOT
	case TokenShl:
		return token.SHL
	case TokenShr:
		return token.SHR
	default:
		panic("unexpected binary operator token")
	}
}

func (a *TypeAndValue) UnaryOp(op TokenKind) (res *TypeAndValue) {
	res = &TypeAndValue{
		Mode: a.Mode,
		Type: a.Type,
	}
	if res.Mode == AddressModeConstant {
		res.Value = constant.UnaryOp(convertTokenToConstantToken(op), a.Value, 0)
	} else {
		res.Mode = AddressModeComputedValue
	}
	return
}

func (a *TypeAndValue) BinaryOpWithType(op TokenKind, b *TypeAndValue, t Type) (res *TypeAndValue) {
	res = &TypeAndValue{
		Mode: a.Mode.Combine(b.Mode),
		Type: t,
	}
	if res.Mode == AddressModeConstant {
		res.Value = constant.BinaryOp(a.Value, convertTokenToConstantToken(op), b.Value)
	}
	return
}

func (a *TypeAndValue) CompareWithType(op TokenKind, b *TypeAndValue, t Type) (res *TypeAndValue) {
	res = &TypeAndValue{
		Mode: a.Mode.Combine(b.Mode),
		Type: t,
	}
	if res.Mode == AddressModeConstant {
		res.Value = constant.MakeBool(constant.Compare(a.Value, convertTokenToConstantToken(op), b.Value))
	}
	return
}

func (a *TypeAndValue) ShiftWithType(op TokenKind, b *TypeAndValue, t Type) (res *TypeAndValue) {
	res = &TypeAndValue{
		Mode: a.Mode.Combine(b.Mode),
		Type: t,
	}
	if res.Mode == AddressModeConstant {
		v, ok := constant.Int64Val(b.Value)
		if !ok || v < 0 {
			panic("unexpected shift value")
		}
		res.Value = constant.Shift(a.Value, convertTokenToConstantToken(op), uint(v))
	}
	return
}

type Checker struct {
	DefaultVisitor
	unit          *Unit
	scopeStack    []*Scope
	functionStack []*FuncDecl
}

func NewChecker(u *Unit) *Checker {
	return &Checker{
		unit: u,
	}
}

func (checker *Checker) currentScope() *Scope {
	return checker.scopeStack[len(checker.scopeStack)-1]
}

func (checker *Checker) enterScope(scope *Scope) {
	if scope == nil {
		panic("entering nil scope")
	}
	checker.scopeStack = append(checker.scopeStack, scope)
}

func (checker *Checker) leaveScope() {
	checker.scopeStack = checker.scopeStack[:len(checker.scopeStack)-1]
}

func (checker *Checker) currentFunction() *FuncDecl {
	return checker.functionStack[len(checker.functionStack)-1]
}

func (checker *Checker) enterFunction(function *FuncDecl) {
	if function == nil {
		panic("entering nil function")
	}
	checker.functionStack = append(checker.functionStack, function)
}

func (checker *Checker) leaveFunction() {
	checker.functionStack = checker.functionStack[:len(checker.functionStack)-1]
}

func (checker *Checker) Check() bool {
	checker.unit.semanticInfo = NewSemanticInfo()
	globalScope := checker.unit.semanticInfo.createScopeFor(checker.unit.rootFile, nil, "global")

	checker.enterScope(globalScope)
	defer checker.leaveScope()

	checker.shallowWalk()

	for _, sym := range globalScope.Symbols {
		checker.resolveSymbol(sym)
	}

	return !checker.unit.HasErrors()
}

func (checker *Checker) shallowWalk() {
	for _, d := range checker.unit.rootFile.decls {
		switch decl := d.(type) {
		case *GenericDecl:
			checker.shallowWalkGenericDecl(decl)
		case *FuncDecl:
			checker.shallowWalkFuncDecl(decl)
		default:
			panic("unexpected decl kind")
		}
	}
}

func (checker *Checker) shallowWalkGenericDecl(d *GenericDecl) {
	switch d.DeclToken.Kind() {
	case TokenConst:
		for si, s := range d.Specs {
			spec := s.(*ValueSpec)
			for ei, name := range spec.LHS {
				sym := NewConstSymbol(name.Token, d, d.SourceRange(), si, ei)
				checker.addSymbol(sym)
			}
		}
	case TokenVar:
		for si, s := range d.Specs {
			spec := s.(*ValueSpec)
			for ei, name := range spec.LHS {
				sym := NewVarSymbol(name.Token, d, d.SourceRange(), si, ei, nil)
				checker.addSymbol(sym)
			}
		}
	case TokenType:
		for _, s := range d.Specs {
			spec := s.(*TypeSpec)
			sym := NewTypeSymbol(spec.Name.Token, d, spec.Name.SourceRange(), spec.Type, !spec.Assign.valid())
			checker.addSymbol(sym)
		}
	default:
		panic("unexpected GenericDecl kind")
	}
}

func (checker *Checker) shallowWalkFuncDecl(d *FuncDecl) {
	sym := NewFuncSymbol(d.Name.Token, d, d.SourceRange())
	checker.addSymbol(sym)
}

func (checker *Checker) addSymbol(sym Symbol) Symbol {
	scope := checker.currentScope()
	if oldSym := scope.ShallowFind(sym.Name()); oldSym != nil {
		checker.error(
			NewError(sym.SourceRange(), "symbol '%v' redefinition", sym.Name()).
				Note(oldSym.SourceRange(), "first declared here"),
		)
		return oldSym
	}
	scope.Add(sym)
	sym.SetScope(scope)
	return sym
}

func (checker *Checker) error(e Error) {
	checker.unit.rootFile.error(e)
}

func (checker *Checker) resolveSymbol(sym Symbol) *TypeAndValue {
	if sym.ResolveState() == ResolveStateResolved {
		return checker.unit.semanticInfo.TypeOf(sym)
	} else if sym.ResolveState() == ResolveStateResolving {
		checker.error(
			NewError(sym.SourceRange(), "symbol %v has a cyclic dependency", sym.Name()),
		)
		return &TypeAndValue{
			Mode: AddressModeInvalid,
			Type: BuiltinVoidType,
		}
	}

	var symType *TypeAndValue
	sym.SetResolveState(ResolveStateResolving)
	switch symbol := sym.(type) {
	case *FuncSymbol:
		symType = checker.resolveFuncSymbol(symbol)
	case *TypeSymbol:
		symType = checker.resolveTypeSymbol(symbol)
	case *VarSymbol:
		symType = checker.resolveVarSymbol(symbol)
	case *ConstSymbol:
		symType = checker.resolveConstSymbol(symbol)
	default:
		panic("unexpected symbol type")
	}
	sym.SetResolveState(ResolveStateResolved)

	checker.unit.semanticInfo.SetTypeOf(sym, symType)

	switch symbol := sym.(type) {
	case *FuncSymbol:
		checker.resolveFuncBody(symbol)
	case *TypeSymbol:
		// nothing to do here
	case *VarSymbol:
		// nothing to do
	case *ConstSymbol:
		// nothing to do
	default:
		panic("unexpected symbol type")
	}

	globalScope := checker.unit.semanticInfo.ScopeOf(checker.unit.rootFile)
	if sym.Scope() == globalScope {
		checker.unit.semanticInfo.ReachableSymbols = append(checker.unit.semanticInfo.ReachableSymbols, sym)
	}

	return symType
}

func (checker *Checker) resolveFuncSymbol(sym *FuncSymbol) *TypeAndValue {
	scope := checker.unit.semanticInfo.createScopeFor(sym, checker.currentScope(), sym.Name())

	checker.enterScope(scope)
	defer checker.leaveScope()

	funcDecl := sym.SymDecl.(*FuncDecl)

	checker.enterFunction(funcDecl)
	defer checker.leaveFunction()

	funcType := checker.resolveFuncTypeExpr(funcDecl.Type)

	checker.unit.semanticInfo.SetTypeOf(sym.SymDecl, funcType)
	return funcType
}

func (checker *Checker) resolveTypeSymbol(sym *TypeSymbol) *TypeAndValue {
	t := checker.resolveExpr(sym.TypeExpr)
	if sym.IsStrong {
		t.Type = checker.unit.semanticInfo.TypeInterner.InternStrongTypeAlias(sym.Name(), t.Type)
	} else {
		t.Type = checker.unit.semanticInfo.TypeInterner.InternWeakTypeAlias(sym.Name(), t.Type)
	}
	checker.unit.semanticInfo.SetTypeOf(sym.SymDecl, t)
	return t
}

func (checker *Checker) resolveFuncBody(sym *FuncSymbol) {
	scope := checker.unit.semanticInfo.ScopeOf(sym)
	checker.enterScope(scope)
	defer checker.leaveScope()

	funcDecl := sym.SymDecl.(*FuncDecl)
	checker.enterFunction(funcDecl)
	defer checker.leaveFunction()

	for _, stmt := range funcDecl.Body.Stmts {
		checker.resolveStmt(stmt, ResolveStmtProperties{})
	}
}

func (checker *Checker) resolveVarSymbol(sym *VarSymbol) *TypeAndValue {
	invalidType := &TypeAndValue{
		Mode:  AddressModeInvalid,
		Type:  BuiltinVoidType,
		Value: nil,
	}

	spec := sym.SymDecl.(*GenericDecl).Specs[sym.SpecIndex].(*ValueSpec)

	var varType Type
	if spec.Type != nil {
		varType = checker.resolveExpr(spec.Type).Type
	}

	rhsTypes, sourceRanges := checker.resolveAndUnpackTypesFromExprList(spec.RHS)
	if varType == nil {
		if len(rhsTypes) == 0 {
			checker.error(NewError(sym.SourceRange(), "variable declaration requires type or an initializer"))
			return invalidType
		}

		if sym.ExprIndex >= len(rhsTypes) {
			checker.error(NewError(sym.SourceRange(), "assignment mismatch: %v variables but %v values", sym.ExprIndex+1, len(rhsTypes)))
			return invalidType
		}

		varType = rhsTypes[sym.ExprIndex].Type
	} else {
		if len(rhsTypes) > 0 {
			if sym.ExprIndex >= len(rhsTypes) {
				checker.error(NewError(sym.SourceRange(), "assignment mismatch: %v variables but %v values", sym.ExprIndex+1, len(rhsTypes)))
				return invalidType
			}

			if !rhsTypes[sym.ExprIndex].Type.Equal(varType) {
				checker.error(NewError(sourceRanges[sym.ExprIndex], "type mismatch in variable declaration expected '%v', got '%v'", varType, rhsTypes[sym.ExprIndex].Type))
				return invalidType
			}
		}
	}

	return &TypeAndValue{
		Mode:  AddressModeVariable,
		Type:  varType,
		Value: nil,
	}
}

func (checker *Checker) resolveConstSymbol(sym *ConstSymbol) *TypeAndValue {
	invalidType := &TypeAndValue{
		Mode:  AddressModeInvalid,
		Type:  BuiltinVoidType,
		Value: nil,
	}

	spec := sym.SymDecl.(*GenericDecl).Specs[sym.SpecIndex].(*ValueSpec)
	if len(spec.RHS) == 0 {
		checker.error(NewError(sym.SourceRange(), "constant declaration requires an initializer"))
		return invalidType
	}

	rhsValues, sourceRanges := checker.resolveAndUnpackTypesFromExprList(spec.RHS)

	if sym.ExprIndex >= len(rhsValues) {
		checker.error(NewError(sym.SourceRange(), "assignment mismatch: %v constants but %v values", sym.ExprIndex+1, len(rhsValues)))
		return invalidType
	}

	rhsValue := rhsValues[sym.ExprIndex]
	rhsType := rhsValue.Type
	sourceRange := sourceRanges[sym.ExprIndex]

	if rhsValue.Mode != AddressModeConstant {
		checker.error(NewError(sourceRange, "constant declaration requires a constant expression"))
		return invalidType
	}

	if spec.Type != nil {
		constType := checker.resolveExpr(spec.Type).Type
		if !rhsType.Equal(constType) {
			checker.error(NewError(sourceRange, "type mismatch in constant declaration expected '%v', got '%v'", constType, rhsType))
			return invalidType
		}
	}

	return &TypeAndValue{
		Mode:  AddressModeConstant,
		Type:  rhsType,
		Value: rhsValue.Value,
	}
}

func (checker *Checker) resolveExpr(expr Expr) (t *TypeAndValue) {
	if exprType := checker.unit.semanticInfo.TypeOf(expr); exprType != nil {
		return exprType
	}

	switch e := expr.(type) {
	case *LiteralExpr:
		t = checker.resolveLiteralExpr(e)
	case *IdentifierExpr:
		t = checker.resolveIdentifierExpr(e)
	case *ParenExpr:
		t = checker.resolveParenExpr(e)
	case *SelectorExpr:
		t = checker.resolveSelectorExpr(e)
	case *UnaryExpr:
		t = checker.resolveUnaryExpr(e)
	case *BinaryExpr:
		t = checker.resolveBinaryExpr(e)
	case *NamedTypeExpr:
		t = checker.resolveNamedTypeExpr(e)
	case *CallExpr:
		t = checker.resolveCallExpr(e)
	case *ArrayTypeExpr:
		t = checker.resolveArrayTypeExpr(e)
	case *FuncTypeExpr:
		t = checker.resolveFuncTypeExpr(e)
	case *StructTypeExpr:
		t = checker.resolveStructTypeExpr(e)
	default:
		panic("unexpected expr type")
	}

	checker.unit.semanticInfo.SetTypeOf(expr, t)
	return t
}

func (checker *Checker) resolveAndUnpackTypesFromExprList(exprs []Expr) (types []*TypeAndValue, sourceRanges []SourceRange) {
	if len(exprs) == 1 {
		e := exprs[0]
		tv := checker.resolveExpr(e)
		switch t := tv.Type.(type) {
		case *TupleType:
			for _, tt := range t.Types {
				types = append(types, &TypeAndValue{Mode: tv.Mode, Type: tt, Value: nil})
				sourceRanges = append(sourceRanges, e.SourceRange())
			}
		default:
			types = append(types, tv)
			sourceRanges = append(sourceRanges, e.SourceRange())
		}
	} else {
		for _, e := range exprs {
			types = append(types, checker.resolveExpr(e))
			sourceRanges = append(sourceRanges, e.SourceRange())
		}
	}
	return
}

func (checker *Checker) resolveLiteralExpr(e *LiteralExpr) *TypeAndValue {
	switch e.Token.Kind() {
	case TokenLiteralInt:
		i, err := strconv.ParseInt(e.Token.Value(), 0, 64)
		if err == nil {
			return &TypeAndValue{
				Mode:  AddressModeConstant,
				Type:  BuiltinIntType,
				Value: constant.MakeInt64(i),
			}
		} else {
			checker.error(NewError(e.Token.SourceRange(), "invalid integer value").
				Note(e.Token.SourceRange(), "%v", err),
			)
			return &TypeAndValue{
				Mode:  AddressModeInvalid,
				Type:  BuiltinVoidType,
				Value: nil,
			}
		}
	case TokenLiteralFloat:
		f, err := strconv.ParseFloat(e.Token.Value(), 64)
		if err == nil {
			return &TypeAndValue{
				Mode:  AddressModeConstant,
				Type:  BuiltinFloat32Type,
				Value: constant.MakeFloat64(f),
			}
		} else {
			checker.error(NewError(e.Token.SourceRange(), "invalid float value").
				Note(e.Token.SourceRange(), "%v", err),
			)
			return &TypeAndValue{
				Mode:  AddressModeInvalid,
				Type:  BuiltinVoidType,
				Value: nil,
			}
		}
	case TokenTrue:
		return &TypeAndValue{
			Mode:  AddressModeConstant,
			Type:  BuiltinBoolType,
			Value: constant.MakeBool(true),
		}
	case TokenFalse:
		return &TypeAndValue{
			Mode:  AddressModeConstant,
			Type:  BuiltinBoolType,
			Value: constant.MakeBool(false),
		}
	default:
		return &TypeAndValue{
			Mode:  AddressModeInvalid,
			Type:  BuiltinVoidType,
			Value: nil,
		}
	}
}

func (checker *Checker) resolveIdentifierExpr(e *IdentifierExpr) *TypeAndValue {
	scope := checker.currentScope()
	symbol := scope.Find(e.Token.Value())
	if symbol == nil {
		checker.error(NewError(e.SourceRange(), "undeclared identifier"))
		return &TypeAndValue{
			Mode:  AddressModeInvalid,
			Type:  BuiltinVoidType,
			Value: nil,
		}
	}

	checker.unit.semanticInfo.SetSymbolOfIdentifier(e, symbol)

	return checker.resolveSymbol(symbol)
}

func (checker *Checker) resolveParenExpr(e *ParenExpr) *TypeAndValue {
	return checker.resolveExpr(e.Base)
}

func (checker *Checker) resolveSelectorExpr(e *SelectorExpr) *TypeAndValue {
	baseType := checker.resolveExpr(e.Base)

	invalidResult := &TypeAndValue{
		Mode: AddressModeInvalid,
		Type: BuiltinVoidType,
	}

	switch t := baseType.Type.Resolve(true).(type) {
	case *StructType:
		if structField := t.FindField(e.Selector.Token.Value()); structField != nil {
			return &TypeAndValue{
				Mode: baseType.Mode,
				Type: structField.Type,
			}
		} else {
			checker.error(NewError(
				e.Selector.SourceRange(),
				"field '%v' cannot be found in struct '%v'",
				e.Selector.Token.Value(),
				baseType.Type,
			))
		}
	case *VectorType:
		return checker.resolveVectorSwizzle(e, t)
	default:
		checker.error(NewError(e.SourceRange(), "type '%v' does not support selector expr", baseType.Type))
	}
	return invalidResult
}

func (checker *Checker) resolveVectorSwizzle(e *SelectorExpr, base *VectorType) *TypeAndValue {
	isValidSwizzle := func(swizzle string, numComponents int) bool {
		if len(swizzle) == 0 {
			return false
		}

		getStyle := func(r rune) string {
			switch r {
			case 'x', 'y', 'z', 'w':
				return "xyzw"
			case 'r', 'g', 'b', 'a':
				return "rgba"
			case 's', 't', 'q', 'p':
				return "stqp"
			default:
				return ""
			}
		}

		inRange := func(style string, r rune) bool {
			for i := 0; i < numComponents; i++ {
				if style[i] == byte(r) {
					return true
				}
			}
			return false
		}

		style := getStyle(rune(swizzle[0]))
		for _, char := range swizzle {
			if !inRange(style, char) {
				return false
			}
		}

		return true
	}

	getSwizzleResultType := func(componentType Type, swizzleLength int) Type {
		if swizzleLength == 1 {
			return componentType
		}

		switch componentType {
		case BuiltinFloat32Type:
			switch swizzleLength {
			case 2:
				return BuiltinF32x2Type
			case 3:
				return BuiltinF32x3Type
			case 4:
				return BuiltinF32x4Type
			}
		case BuiltinFloat64Type:
			switch swizzleLength {
			case 2:
				return BuiltinF64x2Type
			case 3:
				return BuiltinF64x3Type
			case 4:
				return BuiltinF64x4Type
			}
		case BuiltinIntType:
			switch swizzleLength {
			case 2:
				return BuiltinI32x2Type
			case 3:
				return BuiltinI32x3Type
			case 4:
				return BuiltinI32x4Type
			}
		case BuiltinUintType:
			switch swizzleLength {
			case 2:
				return BuiltinU32x2Type
			case 3:
				return BuiltinU32x3Type
			case 4:
				return BuiltinU32x4Type
			}
		case BuiltinBoolType:
			switch swizzleLength {
			case 2:
				return BuiltinB32x2Type
			case 3:
				return BuiltinB32x3Type
			case 4:
				return BuiltinB32x4Type
			}
		default:
			panic("unexpected vector component type")
		}

		return BuiltinVoidType
	}

	swizzle := e.Selector.Token.Value()
	if !isValidSwizzle(swizzle, base.Width) {
		checker.error(NewError(e.Selector.Token.SourceRange(), "invalid swizzle '%v' for %v-component vector '%v'", swizzle, base.Width, base))
		return &TypeAndValue{
			Mode:  AddressModeInvalid,
			Type:  BuiltinVoidType,
			Value: nil,
		}
	}

	return &TypeAndValue{
		Mode:  AddressModeComputedValue,
		Type:  getSwizzleResultType(base.UnderlyingType, len(swizzle)),
		Value: nil,
	}
}

func (checker *Checker) resolveBinaryExpr(e *BinaryExpr) *TypeAndValue {
	lhsType := checker.resolveExpr(e.LHS)
	rhsType := checker.resolveExpr(e.RHS)

	invalidResult := &TypeAndValue{
		Mode: AddressModeInvalid,
		Type: BuiltinVoidType,
	}

	isVectorType := func(t Type) (*VectorType, bool) {
		if vt, ok := t.Resolve(true).(*VectorType); ok {
			return vt, true
		}
		return nil, false
	}

	lhsVecType, lhsIsVec := isVectorType(lhsType.Type)
	rhsVecType, rhsIsVec := isVectorType(rhsType.Type)
	vecWidth := 1
	if lhsIsVec && rhsIsVec {
		vecWidth = lhsVecType.Width
	} else if lhsIsVec {
		vecWidth = lhsVecType.Width
	} else if rhsIsVec {
		vecWidth = rhsVecType.Width
	}

	vectorBooleanByWidth := func(width int) Type {
		switch width {
		case 1:
			return BuiltinBoolType
		case 2:
			return BuiltinB32x2Type
		case 3:
			return BuiltinB32x3Type
		case 4:
			return BuiltinB32x4Type
		default:
			panic("unsupported vector width")
		}
	}

	checkIsCompatibleTypes := func(e Expr, lhsType, rhsType Type) bool {
		incompatible := false
		if lhsIsVec && rhsIsVec {
			incompatible = lhsType.Equal(rhsType)
		} else if lhsIsVec || rhsIsVec {
			var scalarType Type
			var vectorType *VectorType
			if lhsIsVec {
				vectorType = lhsVecType
				scalarType = rhsType
			} else {
				scalarType = lhsType
				vectorType = rhsVecType
			}
			incompatible = vectorType.UnderlyingType.Equal(scalarType)
		} else {
			incompatible = lhsType.Equal(rhsType)
		}
		if !incompatible {
			checker.error(NewError(
				e.SourceRange(),
				"type mismatch in binary expression, lhs is '%v' and rhs is '%v'",
				lhsType,
				rhsType,
			))
			return false
		}
		return true
	}

	hasTypeProperty := func(e Expr, t Type, hasFeature bool, capName string) bool {
		if !hasFeature {
			checker.error(NewError(
				e.SourceRange(),
				"type '%v' doesn't support %v",
				t,
				capName,
			))
			return false
		}
		return true
	}

	switch e.Operator.Kind() {
	case TokenOr, TokenAnd, TokenXor, TokenAndNot:
		if checkIsCompatibleTypes(e, lhsType.Type, rhsType.Type) &&
			hasTypeProperty(e.LHS, lhsType.Type, lhsType.Type.Properties().HasBitOps, "bitwise operations") &&
			hasTypeProperty(e.RHS, rhsType.Type, rhsType.Type.Properties().HasBitOps, "bitwise operations") {
			return lhsType.BinaryOpWithType(e.Operator.Kind(), rhsType, lhsType.Type)
		}
	case TokenAdd, TokenSub, TokenMul, TokenDiv:
		if checkIsCompatibleTypes(e, lhsType.Type, rhsType.Type) &&
			hasTypeProperty(e.LHS, lhsType.Type, lhsType.Type.Properties().HasArithmetic, "arithmetic operations") &&
			hasTypeProperty(e.RHS, rhsType.Type, rhsType.Type.Properties().HasArithmetic, "arithmetic operations") {
			return lhsType.BinaryOpWithType(e.Operator.Kind(), rhsType, lhsType.Type)
		}
	case TokenMod:
		if checkIsCompatibleTypes(e, lhsType.Type, rhsType.Type) &&
			hasTypeProperty(e.LHS, lhsType.Type, lhsType.Type.Properties().HasModulus, "modulus operations") &&
			hasTypeProperty(e.RHS, rhsType.Type, rhsType.Type.Properties().HasModulus, "modulus operations") {
			return lhsType.BinaryOpWithType(e.Operator.Kind(), rhsType, lhsType.Type)
		}
	case TokenLOr, TokenLAnd:
		if checkIsCompatibleTypes(e, lhsType.Type, rhsType.Type) &&
			hasTypeProperty(e.LHS, lhsType.Type, lhsType.Type.Properties().HasLogicOps, "logic operations") &&
			hasTypeProperty(e.RHS, rhsType.Type, rhsType.Type.Properties().HasLogicOps, "logic operations") {
			return lhsType.BinaryOpWithType(e.Operator.Kind(), rhsType, lhsType.Type)
		}
	case TokenLT, TokenGT, TokenLE, TokenGE:
		if checkIsCompatibleTypes(e, lhsType.Type, rhsType.Type) &&
			hasTypeProperty(e.LHS, lhsType.Type, lhsType.Type.Properties().HasCompare, "compare operations") &&
			hasTypeProperty(e.RHS, rhsType.Type, rhsType.Type.Properties().HasCompare, "compare operations") {
			return lhsType.CompareWithType(e.Operator.Kind(), rhsType, vectorBooleanByWidth(vecWidth))
		}
	case TokenEQ, TokenNE:
		if checkIsCompatibleTypes(e, lhsType.Type, rhsType.Type) &&
			hasTypeProperty(e.LHS, lhsType.Type, lhsType.Type.Properties().HasEquality, "equality operations") &&
			hasTypeProperty(e.RHS, rhsType.Type, rhsType.Type.Properties().HasEquality, "equality operations") {
			return lhsType.CompareWithType(e.Operator.Kind(), rhsType, vectorBooleanByWidth(vecWidth))
		}
	case TokenShl, TokenShr:
		if !rhsType.Type.Properties().Integral {
			checker.error(NewError(
				e.RHS.SourceRange(),
				"shift operator should be integral type instead of '%v'",
				rhsType.Type,
			))
			return invalidResult
		}

		if rhsType.Mode == AddressModeConstant && constant.Compare(rhsType.Value, token.LSS, constant.MakeInt64(0)) {
			checker.error(NewError(
				e.RHS.SourceRange(),
				"shift operator should not be negative, but it has value '%v'",
				rhsType.Value,
			))
			return invalidResult
		}

		if hasTypeProperty(e.LHS, lhsType.Type, lhsType.Type.Properties().HasBitOps, "bitwise operations") {
			return lhsType.ShiftWithType(e.Operator.Kind(), rhsType, lhsType.Type)
		}
	default:
		panic("unexpected binary operator")
	}
	return invalidResult
}

func (checker *Checker) resolveNamedTypeExpr(e *NamedTypeExpr) *TypeAndValue {
	if e.Package.valid() {
		panic("we don't support packages yet")
	}
	scope := checker.currentScope()
	if typeSym := scope.Find(e.TypeName.Value()); typeSym != nil {
		return checker.resolveSymbol(typeSym)
	}
	return &TypeAndValue{
		Mode:  AddressModeType,
		Type:  typeFromName(e.TypeName),
		Value: nil,
	}
}

func (checker *Checker) resolveUnaryExpr(e *UnaryExpr) *TypeAndValue {
	t := checker.resolveExpr(e.Base)

	invalidResult := &TypeAndValue{
		Mode:  AddressModeInvalid,
		Type:  BuiltinVoidType,
		Value: nil,
	}

	hasTypeProperty := func(e Expr, t Type, hasFeature bool, capName string) bool {
		if !hasFeature {
			checker.error(NewError(
				e.SourceRange(),
				"type '%v' doesn't support %v",
				t,
				capName,
			))
			return false
		}
		return true
	}

	switch e.Operator.Kind() {
	case TokenAdd:
		fallthrough
	case TokenSub:
		if !hasTypeProperty(e.Base, t.Type, t.Type.Properties().HasArithmetic, "arithmetic operations") {
			return invalidResult
		}
	case TokenNot:
		if !hasTypeProperty(e.Base, t.Type, t.Type.Properties().HasLogicOps, "logic operations") {
			return invalidResult
		}
	case TokenXor:
		if !hasTypeProperty(e.Base, t.Type, t.Type.Properties().HasBitOps, "bitwise operations") {
			return invalidResult
		}
	default:
		panic("invalid unary operator")
	}

	return t.UnaryOp(e.Operator.Kind())
}

func (checker *Checker) resolveCallExpr(e *CallExpr) *TypeAndValue {
	t := checker.resolveExpr(e.Base)

	res := &TypeAndValue{
		Mode:  AddressModeInvalid,
		Type:  BuiltinVoidType,
		Value: nil,
	}

	funcType, ok := t.Type.(*FuncType)
	if !ok {
		checker.error(NewError(e.SourceRange(), "invalid call expression, expected function type but found '%v'", t.Type))
		return res
	}

	arguments, sourceRanges := checker.resolveAndUnpackTypesFromExprList(e.Args)
	if len(arguments) != len(funcType.ParameterTypes) {
		argumentTypes := make([]Type, len(arguments))
		for i, a := range arguments {
			argumentTypes[i] = a.Type
		}
		checker.error(NewError(e.SourceRange(), "expected %v arguments, but found %v", len(funcType.ParameterTypes), len(e.Args)).
			Note(e.SourceRange(), "have %v, want %v", TupleType{Types: argumentTypes}, TupleType{Types: funcType.ParameterTypes}),
		)
		return res
	}

	for i, a := range arguments {
		parameterType := funcType.ParameterTypes[i]
		if !a.Type.Equal(parameterType) {
			checker.error(NewError(sourceRanges[i], "incorrect argument type '%v', expected '%v'", a.Type, parameterType))
			return res
		}
	}

	res.Mode = AddressModeComputedValue
	if len(funcType.ReturnTypes) == 1 {
		res.Type = funcType.ReturnTypes[0]
	} else if len(funcType.ReturnTypes) > 1 {
		res.Type = checker.unit.semanticInfo.TypeInterner.InternTupleType(funcType.ReturnTypes)
	}
	return res
}

func (checker *Checker) resolveArrayTypeExpr(e *ArrayTypeExpr) *TypeAndValue {
	elementType := checker.resolveExpr(e.ElementType)

	res := &TypeAndValue{
		Mode:  AddressModeType,
		Type:  BuiltinVoidType,
		Value: nil,
	}

	lengthAsInt := 0
	if e.Length != nil {
		lengthType := checker.resolveExpr(e.Length)
		if lengthType.Mode != AddressModeConstant {
			checker.error(NewError(e.Length.SourceRange(), "array type length should be constant"))
			return res
		}
		if lengthType.Value != nil {
			if lengthType.Value.Kind() != constant.Int {
				checker.error(NewError(e.Length.SourceRange(), "array type length should be integer"))
				return res
			}
			valueAsInt, exact := constant.Int64Val(lengthType.Value)
			if !exact {
				checker.error(NewError(e.Length.SourceRange(), "array type length does not fit in 64bit integer"))
				return res
			}
			lengthAsInt = int(valueAsInt)
		}
	}

	res.Type = checker.unit.semanticInfo.TypeInterner.InternArrayType(lengthAsInt, elementType.Type)
	return res
}

func (checker *Checker) resolveFuncTypeExpr(e *FuncTypeExpr) *TypeAndValue {
	processFields := func(fields []Field) (types []Type) {
		for _, field := range fields {
			fieldType := checker.resolveExpr(field.Type)
			if len(field.Names) > 0 {
				for _, name := range field.Names {
					v := NewVarSymbol(name.Token, nil, name.SourceRange(), -1, -1, nil)
					v.SetResolveState(ResolveStateResolved)
					checker.unit.semanticInfo.SetTypeOf(v, fieldType)
					checker.addSymbol(v)
					checker.unit.semanticInfo.SetSymbolOfIdentifier(name, v)
					types = append(types, fieldType.Type)
				}
			} else {
				types = append(types, fieldType.Type)
			}
		}
		return types
	}

	parameterTypes := processFields(e.Parameters.Fields)
	var returnTypes []Type
	if e.Result != nil {
		returnTypes = processFields(e.Result.Fields)
	}

	return &TypeAndValue{
		Mode:  AddressModeType,
		Type:  checker.unit.semanticInfo.TypeInterner.InternFuncType(parameterTypes, returnTypes),
		Value: nil,
	}
}

func (checker *Checker) resolveStructTypeExpr(e *StructTypeExpr) *TypeAndValue {
	var names []string
	var types []StructTypeField

	type fieldInfo struct {
		name        string
		sourceRange SourceRange
	}
	fieldsMap := make(map[string]fieldInfo)

	checkExistingFields := func(name string, sourceRange SourceRange) bool {
		if existing, ok := fieldsMap[name]; ok {
			checker.error(
				NewError(
					sourceRange,
					"field '%v' redefinition",
					name,
				).Note(
					existing.sourceRange,
					"first declared here",
				),
			)
			return true
		}
		fieldsMap[name] = fieldInfo{name: name, sourceRange: sourceRange}
		return false
	}

	invalidType := &TypeAndValue{
		Mode:  AddressModeType,
		Type:  BuiltinVoidType,
		Value: nil,
	}

	for _, field := range e.FieldList.Fields {
		if len(field.Names) > 0 {
			for _, id := range field.Names {
				if checkExistingFields(id.Token.Value(), id.SourceRange()) {
					return invalidType
				}
				names = append(names, id.Token.Value())
				types = append(types, StructTypeField{
					Identifer: id,
					Type:      checker.resolveExpr(field.Type).Type,
				})
			}
		} else {
			fieldType := checker.resolveExpr(field.Type)
			if strongAlias, ok := fieldType.Type.(*StrongAliasType); ok {
				if checkExistingFields(strongAlias.Name, field.Type.SourceRange()) {
					return invalidType
				}
				names = append(names, strongAlias.Name)
				types = append(types, StructTypeField{
					Identifer: nil,
					Type:      fieldType.Type,
				})
			} else if weakAlias, ok := fieldType.Type.(*WeakAliasType); ok {
				if checkExistingFields(weakAlias.Name, field.Type.SourceRange()) {
					return invalidType
				}
				names = append(names, weakAlias.Name)
				types = append(types, StructTypeField{
					Identifer: nil,
					Type:      fieldType.Type,
				})
			} else {
				checker.error(NewError(field.Type.SourceRange(), "Cannot embed type '%v'", field.Type))
			}
		}
	}

	return &TypeAndValue{
		Mode:  AddressModeType,
		Type:  checker.unit.semanticInfo.TypeInterner.InternStructType(names, types),
		Value: nil,
	}
}

func typeFromName(name Token) Type {
	switch name.Value() {
	case "bool":
		return BuiltinBoolType
	case "int":
		return BuiltinIntType
	case "uint":
		return BuiltinUintType
	case "float32":
		return BuiltinFloat32Type
	case "float64":
		return BuiltinFloat64Type
	case BuiltinF32x2Type.name:
		return BuiltinF32x2Type
	case BuiltinF32x3Type.name:
		return BuiltinF32x3Type
	case BuiltinF32x4Type.name:
		return BuiltinF32x4Type
	case BuiltinF64x2Type.name:
		return BuiltinF64x2Type
	case BuiltinF64x3Type.name:
		return BuiltinF64x3Type
	case BuiltinF64x4Type.name:
		return BuiltinF64x4Type
	case BuiltinI32x2Type.name:
		return BuiltinI32x2Type
	case BuiltinI32x3Type.name:
		return BuiltinI32x3Type
	case BuiltinI32x4Type.name:
		return BuiltinI32x4Type
	case BuiltinU32x2Type.name:
		return BuiltinU32x2Type
	case BuiltinU32x3Type.name:
		return BuiltinU32x3Type
	case BuiltinU32x4Type.name:
		return BuiltinU32x4Type
	case BuiltinB32x2Type.name:
		return BuiltinB32x2Type
	case BuiltinB32x3Type.name:
		return BuiltinB32x3Type
	case BuiltinB32x4Type.name:
		return BuiltinB32x4Type
	default:
		return BuiltinVoidType
	}
}

func (checker *Checker) resolveStmt(stmt Stmt, properties ResolveStmtProperties) {
	switch s := stmt.(type) {
	case *ExprStmt:
		checker.resolveExpr(s.Expr)
	case *IncDecStmt:
		checker.resolveIncDecStmt(s)
	case *ReturnStmt:
		checker.resolveReturnStmt(s)
	case *BreakStmt:
		checker.resolveBreakStmt(s, properties)
	case *FallthroughStmt:
		checker.resolveFallthroughStmt(s, properties)
	case *ContinueStmt:
		checker.resolveContinueStmt(s, properties)
	case *BlockStmt:
		checker.resolveBlockStmt(s, properties)
	case *AssignStmt:
		checker.resolveAssignStmt(s)
	case *IfStmt:
		checker.resolveIfStmt(s, properties)
	case *ForStmt:
		checker.resolveForStmt(s, properties)
	case *SwitchStmt:
		checker.resolveSwitchStmt(s, properties)
	case *DeclStmt:
		checker.resolveDeclStmt(s)
	default:
		panic("unexpected stmt type")
	}
}

func (checker *Checker) resolveIncDecStmt(s *IncDecStmt) {
	t := checker.resolveExpr(s.Expr)

	if !t.IsAssignable() {
		checker.error(NewError(s.SourceRange(), "expression is not assignable"))
		return
	}

	if !t.Type.Properties().HasArithmetic {
		checker.error(NewError(s.SourceRange(), "type '%v' doesn't support arithmetic operations", t.Type))
		return
	}
}

func (checker *Checker) resolveReturnStmt(s *ReturnStmt) {
	funcDecl := checker.currentFunction()
	if funcDecl == nil {
		checker.error(NewError(s.SourceRange(), "unexpected return statement"))
		return
	}

	returnTypes, sourceRanges := checker.resolveAndUnpackTypesFromExprList(s.Exprs)
	expectedReturnTypes := checker.unit.semanticInfo.TypeOf(funcDecl).Type.(*FuncType).ReturnTypes
	if len(returnTypes) == len(expectedReturnTypes) {
		for i, et := range expectedReturnTypes {
			if t := returnTypes[i]; !t.Type.Equal(et) {
				checker.error(NewError(sourceRanges[i], "incorrect return type '%v', expected '%v'", t.Type, et))
			}
		}
	} else {
		named := funcDecl.Type.Result != nil && len(funcDecl.Type.Result.Fields[0].Names) > 0
		if len(returnTypes) != 0 || !named {
			rTypes := make([]Type, len(returnTypes))
			for i, a := range returnTypes {
				rTypes[i] = a.Type
			}
			checker.error(NewError(s.SourceRange(), "expected %v return values, but found %v", len(expectedReturnTypes), len(rTypes)).
				Note(s.SourceRange(), "have %v, want %v", TupleType{Types: rTypes}, TupleType{Types: expectedReturnTypes}),
			)
		}
	}
}

func (checker *Checker) resolveBreakStmt(s *BreakStmt, properties ResolveStmtProperties) {
	if s.Label.valid() {
		panic("labeled break not supported yet")
	}

	if !properties.acceptsBreak {
		checker.error(NewError(s.SourceRange(), "break statement not within loop or switch"))
	}
}

func (checker *Checker) resolveFallthroughStmt(s *FallthroughStmt, properties ResolveStmtProperties) {
	if properties.isFinalCaseStmt {
		checker.error(NewError(s.SourceRange(), "cannot fallthrough from the final case in a switch"))
	}

	if !properties.acceptsFallthrough {
		checker.error(NewError(s.SourceRange(), "fallthrough statement not within switch"))
	}
}

func (checker *Checker) resolveContinueStmt(s *ContinueStmt, properties ResolveStmtProperties) {
	if s.Label.valid() {
		panic("labeled continue not supported yet")
	}

	if !properties.acceptsContinue {
		checker.error(NewError(s.SourceRange(), "continue statement not within for loop"))
	}
}

func (checker *Checker) resolveBlockStmt(s *BlockStmt, properties ResolveStmtProperties) {
	scope := checker.unit.semanticInfo.createScopeFor(s, checker.currentScope(), "block")
	checker.enterScope(scope)
	defer checker.leaveScope()

	for _, stmt := range s.Stmts {
		checker.resolveStmt(stmt, properties)
	}
}

func (checker *Checker) resolveAssignStmt(s *AssignStmt) {
	hasMultiValue := func(s *AssignStmt, lhsValuesLen, rhsValuesLen int) bool {
		if lhsValuesLen != rhsValuesLen {
			checker.error(NewError(
				s.SourceRange(),
				"assignment mismatch: %v variables but %v values",
				lhsValuesLen,
				rhsValuesLen,
			))
			return false
		}
		return true
	}

	hasSingleValue := func(s *AssignStmt) bool {
		if len(s.LHS) != 1 || len(s.RHS) != 1 {
			checker.error(NewError(
				s.SourceRange(),
				"assignment operator %v requires single value expressions",
				s.Operator.Value(),
			))
			return false
		}
		return true
	}

	checkIsAssignable := func(e Expr, eType *TypeAndValue) {
		if !eType.IsAssignable() {
			checker.error(NewError(
				e.SourceRange(),
				"expression is not assignable",
			))
		}
	}

	checkTypeProperty := func(sourceRange SourceRange, t Type, hasFeature bool, capName string) {
		if !hasFeature {
			checker.error(NewError(
				sourceRange,
				"type '%v' doesn't support %v",
				t,
				capName,
			))
		}
	}

	checkTypeEqual := func(lhsType, rhsType Type, lhsSourceRange, rhsSourceRange SourceRange) {
		if !lhsType.Equal(rhsType) {
			checker.error(
				NewError(
					s.SourceRange(),
					"type mistmatch in assignment",
				).Note(
					lhsSourceRange,
					"LHS type is '%v'",
					lhsType,
				).Note(
					rhsSourceRange,
					"RHS type is '%v'",
					rhsType,
				),
			)
		}
	}

	switch s.Operator.Kind() {
	case TokenColonAssign:
		rhsTypes, _ := checker.resolveAndUnpackTypesFromExprList(s.RHS)
		if !hasMultiValue(s, len(s.LHS), len(rhsTypes)) {
			return
		}

		for _, varName := range s.LHS {
			if _, ok := varName.(*IdentifierExpr); !ok {
				checker.error(NewError(
					varName.SourceRange(),
					"expression can not be used as variable name",
				))
				return
			}
		}

		for i := range s.LHS {
			lhs := s.LHS[i]
			name := lhs.(*IdentifierExpr).Token
			v := NewVarSymbol(name, nil, name.SourceRange(), -1, -1, rhsTypes[i])
			v.SetResolveState(ResolveStateResolved)
			checker.unit.semanticInfo.SetTypeOf(v, &TypeAndValue{Mode: AddressModeVariable, Type: rhsTypes[i].Type})
			checker.addSymbol(v)
			checker.unit.semanticInfo.SetSymbolOfIdentifier(lhs.(*IdentifierExpr), v)
		}
	case TokenAssign:
		rhsTypes, rhsSourceRanges := checker.resolveAndUnpackTypesFromExprList(s.RHS)
		if !hasMultiValue(s, len(s.LHS), len(rhsTypes)) {
			return
		}

		for i := range s.LHS {
			lhs := s.LHS[i]
			lhsType := checker.resolveExpr(lhs)
			checkIsAssignable(lhs, lhsType)
			checkTypeEqual(lhsType.Type, rhsTypes[i].Type, lhs.SourceRange(), rhsSourceRanges[i])
		}
	case TokenAddAssign, TokenSubAssign, TokenMulAssign, TokenDivAssign, TokenModAssign:
		if !hasSingleValue(s) {
			return
		}
		lhs := s.LHS[0]
		lhsType := checker.resolveExpr(lhs)
		checkIsAssignable(lhs, lhsType)
		checkTypeProperty(
			lhs.SourceRange(),
			lhsType.Type,
			lhsType.Type.Properties().HasArithmetic,
			"arithmetic operations",
		)
		rhs := s.RHS[0]
		rhsType := checker.resolveExpr(rhs)
		checkTypeProperty(
			rhs.SourceRange(),
			rhsType.Type,
			rhsType.Type.Properties().HasArithmetic,
			"arithmetic operations",
		)
		checkTypeEqual(lhsType.Type, rhsType.Type, lhs.SourceRange(), rhs.SourceRange())
	case TokenAndAssign, TokenAndNotAssign, TokenOrAssign, TokenXorAssign:
		if !hasSingleValue(s) {
			return
		}
		lhs := s.LHS[0]
		lhsType := checker.resolveExpr(lhs)
		checkIsAssignable(lhs, lhsType)
		checkTypeProperty(
			lhs.SourceRange(),
			lhsType.Type,
			lhsType.Type.Properties().HasBitOps,
			"bitwise operations",
		)
		rhs := s.RHS[0]
		rhsType := checker.resolveExpr(rhs)
		checkTypeProperty(
			rhs.SourceRange(),
			rhsType.Type,
			rhsType.Type.Properties().HasBitOps,
			"bitwise operations",
		)
		checkTypeEqual(lhsType.Type, rhsType.Type, lhs.SourceRange(), rhs.SourceRange())
	case TokenShlAssign, TokenShrAssign:
		if !hasSingleValue(s) {
			return
		}
		lhs := s.LHS[0]
		lhsType := checker.resolveExpr(lhs)
		checkIsAssignable(lhs, lhsType)
		checkTypeProperty(
			lhs.SourceRange(),
			lhsType.Type,
			lhsType.Type.Properties().HasBitOps,
			"bitwise operations",
		)
		rhs := s.RHS[0]
		rhsType := checker.resolveExpr(rhs)
		if !rhsType.Type.Properties().Integral {
			checker.error(NewError(
				rhs.SourceRange(),
				"shift operator should be integral type instead of '%v'",
				rhsType.Type,
			))
		} else {
			if rhsType.Mode == AddressModeConstant && constant.Compare(rhsType.Value, token.LSS, constant.MakeInt64(0)) {
				checker.error(NewError(
					rhs.SourceRange(),
					"shift operator should not be negative, but it has value '%v'",
					rhsType.Value,
				))
			}
		}
	}
}

func (checker *Checker) resolveIfStmt(s *IfStmt, properties ResolveStmtProperties) {
	scope := checker.unit.semanticInfo.createScopeFor(s, checker.currentScope(), "if")
	checker.enterScope(scope)
	defer checker.leaveScope()

	if s.Init != nil {
		checker.resolveStmt(s.Init, properties)
	}

	condType := checker.resolveExpr(s.Cond)
	if !condType.Type.Equal(BuiltinBoolType) {
		checker.error(NewError(
			s.Cond.SourceRange(),
			"if condition should be boolean, but found '%v'",
			condType.Type,
		))
		return
	}

	checker.resolveBlockStmt(s.Body, properties)

	if s.Else != nil {
		checker.resolveStmt(s.Else, properties)
	}
}

func (checker *Checker) resolveForStmt(s *ForStmt, properties ResolveStmtProperties) {
	scope := checker.unit.semanticInfo.createScopeFor(s, checker.currentScope(), "for")
	checker.enterScope(scope)
	defer checker.leaveScope()

	if s.Init != nil {
		checker.resolveStmt(s.Init, properties)
	}

	if s.Cond != nil {
		condType := checker.resolveExpr(s.Cond)
		if !condType.Type.Equal(BuiltinBoolType) {
			checker.error(NewError(
				s.Cond.SourceRange(),
				"for condition should be boolean, but found '%v'",
				condType.Type,
			))
			return
		}
	}

	if s.Post != nil {
		checker.resolveStmt(s.Post, properties)
	}

	properties.acceptsBreak = true
	properties.acceptsContinue = true

	checker.resolveBlockStmt(s.Body, properties)
}

func (checker *Checker) resolveSwitchStmt(s *SwitchStmt, properties ResolveStmtProperties) {
	scope := checker.unit.semanticInfo.createScopeFor(s, checker.currentScope(), "switch")
	checker.enterScope(scope)
	defer checker.leaveScope()

	if s.Init != nil {
		checker.resolveStmt(s.Init, properties)
	}

	var tag *TypeAndValue
	if s.Tag != nil {
		tag = checker.resolveExpr(s.Tag)
		if !tag.Type.Properties().Integral && !tag.Type.Properties().Floating && !tag.Type.Equal(BuiltinBoolType) {
			checker.error(NewError(
				s.Tag.SourceRange(),
				"invalid switch tag type '%v'",
				tag.Type,
			))
		}
	} else {
		tag = &TypeAndValue{
			Mode:  AddressModeConstant,
			Type:  BuiltinBoolType,
			Value: constant.MakeBool(true),
		}
	}

	caseValueMap := make(map[constant.Value]SourceRange)
	for i, stmt := range s.Body.Stmts {
		caseStmt, ok := stmt.(*SwitchCaseStmt)
		if !ok {
			checker.error(NewError(caseStmt.SourceRange(), "only switch case statements are allowed in switch body"))
		}

		properties.acceptsBreak = true
		properties.acceptsFallthrough = true
		properties.isFinalCaseStmt = i+1 == len(s.Body.Stmts)
		checker.resolveSwitchCaseStmt(caseStmt, caseValueMap, tag.Type, properties)
	}
}

func (checker *Checker) resolveSwitchCaseStmt(
	s *SwitchCaseStmt,
	caseValueMap map[constant.Value]SourceRange,
	tagType Type,
	properties ResolveStmtProperties,
) {
	scope := checker.unit.semanticInfo.createScopeFor(s, checker.currentScope(), "case")
	checker.enterScope(scope)
	defer checker.leaveScope()

	for _, expr := range s.LHS {
		t := checker.resolveExpr(expr)

		if !t.Type.Equal(tagType) {
			checker.error(NewError(expr.SourceRange(),
				"case value type '%v' is not comparable to switch tag type '%v'",
				t.Type, tagType,
			))
		}

		if t.Mode == AddressModeConstant && t.Value != nil {
			key := t.Value
			if prev, exists := caseValueMap[key]; exists {
				checker.error(NewError(expr.SourceRange(), "duplicate case value '%v'", key).
					Note(prev, "first case value declared here"))
			} else {
				caseValueMap[key] = expr.SourceRange()
			}
		}
	}

	for i, stmt := range s.RHS {
		fs, ok := stmt.(*FallthroughStmt)
		if ok && i != len(s.RHS)-1 {
			checker.error(NewError(fs.SourceRange(), "fallthrough statement must be the last statement in a case"))
		}
		checker.resolveStmt(stmt, properties)
	}
}

func (checker *Checker) resolveDeclStmt(s *DeclStmt) {
	hasMultiValue := func(s *DeclStmt, lhsValuesLen, rhsValuesLen int) bool {
		if lhsValuesLen != rhsValuesLen {
			checker.error(NewError(
				s.SourceRange(),
				"assignment mismatch: %v variables but %v values",
				lhsValuesLen,
				rhsValuesLen,
			))
			return false
		}
		return true
	}

	type SymbolFunc func(name Token, decl Decl, sourceRange SourceRange, specIndex, exprIndex int, initTAV *TypeAndValue) Symbol
	resolveValueSymbol := func(d *GenericDecl, symbolFunc SymbolFunc) {
		for si, spec := range d.Specs {
			spec := spec.(*ValueSpec)
			rhs, _ := checker.resolveAndUnpackTypesFromExprList(spec.RHS)
			if spec.Assign.valid() {
				if !hasMultiValue(s, len(spec.LHS), len(rhs)) {
					return
				}
			}

			for ei, name := range spec.LHS {
				var initTAV *TypeAndValue = nil
				if ei < len(rhs) {
					initTAV = rhs[ei]
				}
				sym := symbolFunc(name.Token, d, d.SourceRange(), si, ei, initTAV)
				checker.unit.semanticInfo.SetSymbolOfIdentifier(name, sym)
			}
		}
	}

	switch d := s.Decl.(*GenericDecl); d.DeclToken.Kind() {
	case TokenVar:
		resolveValueSymbol(d, func(name Token, decl Decl, sourceRange SourceRange, specIndex, exprIndex int, initTAV *TypeAndValue) Symbol {
			sym := NewVarSymbol(name, decl, sourceRange, specIndex, exprIndex, initTAV)
			checker.addSymbol(sym)
			checker.resolveSymbol(sym)
			return sym
		})
	case TokenConst:
		resolveValueSymbol(d, func(name Token, decl Decl, sourceRange SourceRange, specIndex, exprIndex int, _ *TypeAndValue) Symbol {
			sym := NewConstSymbol(name, decl, sourceRange, specIndex, exprIndex)
			checker.addSymbol(sym)
			checker.resolveSymbol(sym)
			return sym
		})
	case TokenType:
		for _, s := range d.Specs {
			spec := s.(*TypeSpec)
			sym := NewTypeSymbol(spec.Name.Token, d, spec.Name.SourceRange(), spec.Type, !spec.Assign.valid())
			checker.addSymbol(sym)
			checker.resolveSymbol(sym)
			checker.unit.semanticInfo.SetSymbolOfIdentifier(spec.Name, sym)
		}
	default:
		panic("unexpected decl type")
	}
}

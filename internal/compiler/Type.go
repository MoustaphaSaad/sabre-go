package compiler

import (
	"fmt"
	"strings"
)

type TypeProperties struct {
	Size, Align                           int
	Signed, Integral, Floating            bool
	HasBitOps, HasArithmetic, HasLogicOps bool
	HasCompare, HasEquality               bool
}

type Type interface {
	aType()
	Properties() TypeProperties
	String() string
	HashKey() string
	Resolve(resolveStrongAlias bool) Type
	Equal(rhs Type) bool
}

type VoidType struct{}

var BuiltinVoidType = &VoidType{}

func (VoidType) aType() {}
func (VoidType) Properties() TypeProperties {
	return TypeProperties{}
}
func (VoidType) String() string    { return "void" }
func (t VoidType) HashKey() string { return t.String() }
func (t *VoidType) Resolve(bool) Type {
	return t
}
func (lhs *VoidType) Equal(rhs Type) bool {
	return lhs == rhs.Resolve(false)
}

type BoolType struct{}

var BuiltinBoolType = &BoolType{}

func (BoolType) aType() {}
func (BoolType) Properties() TypeProperties {
	return TypeProperties{
		Size:        1,
		Align:       4,
		HasLogicOps: true,
		HasEquality: true,
	}
}
func (BoolType) String() string    { return "bool" }
func (t BoolType) HashKey() string { return t.String() }
func (t *BoolType) Resolve(bool) Type {
	return t
}
func (lhs *BoolType) Equal(rhs Type) bool {
	return lhs == rhs.Resolve(false)
}

type IntType struct{}

var BuiltinIntType = &IntType{}

func (IntType) aType() {}
func (IntType) Properties() TypeProperties {
	return TypeProperties{
		Size:          4,
		Align:         4,
		Signed:        true,
		Integral:      true,
		HasBitOps:     true,
		HasArithmetic: true,
		HasCompare:    true,
		HasEquality:   true,
	}
}
func (IntType) String() string    { return "int" }
func (t IntType) HashKey() string { return t.String() }
func (t *IntType) Resolve(bool) Type {
	return t
}
func (lhs *IntType) Equal(rhs Type) bool {
	return lhs == rhs.Resolve(false)
}

type UintType struct{}

var BuiltinUintType = &UintType{}

func (UintType) aType() {}
func (UintType) Properties() TypeProperties {
	return TypeProperties{
		Size:          4,
		Align:         4,
		Integral:      true,
		HasBitOps:     true,
		HasArithmetic: true,
		HasCompare:    true,
		HasEquality:   true,
	}
}
func (UintType) String() string    { return "uint" }
func (t UintType) HashKey() string { return t.String() }
func (t *UintType) Resolve(bool) Type {
	return t
}
func (lhs *UintType) Equal(rhs Type) bool {
	return lhs == rhs.Resolve(false)
}

type Float32Type struct{}

var BuiltinFloat32Type = &Float32Type{}

func (Float32Type) aType() {}
func (Float32Type) Properties() TypeProperties {
	return TypeProperties{
		Size:          4,
		Align:         4,
		Signed:        true,
		Floating:      true,
		HasArithmetic: true,
		HasCompare:    true,
		HasEquality:   true,
	}
}
func (Float32Type) String() string    { return "float32" }
func (t Float32Type) HashKey() string { return t.String() }
func (t *Float32Type) Resolve(bool) Type {
	return t
}
func (lhs *Float32Type) Equal(rhs Type) bool {
	return lhs == rhs.Resolve(false)
}

type Float64Type struct{}

var BuiltinFloat64Type = &Float64Type{}

func (Float64Type) aType() {}
func (Float64Type) Properties() TypeProperties {
	return TypeProperties{
		Size:          8,
		Align:         8,
		Signed:        true,
		Floating:      true,
		HasArithmetic: true,
		HasCompare:    true,
		HasEquality:   true,
	}
}
func (Float64Type) String() string    { return "float64" }
func (t Float64Type) HashKey() string { return t.String() }
func (t *Float64Type) Resolve(bool) Type {
	return t
}
func (lhs *Float64Type) Equal(rhs Type) bool {
	return lhs == rhs.Resolve(false)
}

type StringType struct{}

var BuiltinStringType = &StringType{}

func (StringType) aType() {}
func (StringType) Properties() TypeProperties {
	return TypeProperties{}
}
func (StringType) String() string    { return "string" }
func (t StringType) HashKey() string { return t.String() }
func (t *StringType) Resolve(bool) Type {
	return t
}
func (lhs *StringType) Equal(rhs Type) bool {
	return lhs == rhs.Resolve(false)
}

type VectorType struct {
	UnderlyingType Type
	Width          int
	properties     TypeProperties
	name           string
}

var (
	BuiltinF32x2Type = &VectorType{
		UnderlyingType: BuiltinFloat32Type,
		Width:          2,
		properties: TypeProperties{
			Size:          8,
			Align:         8,
			Signed:        true,
			Floating:      true,
			HasArithmetic: true,
			HasCompare:    true,
			HasEquality:   true,
		},
		name: "f32x2",
	}
	BuiltinF32x3Type = &VectorType{
		UnderlyingType: BuiltinFloat32Type,
		Width:          3,
		properties: TypeProperties{
			Size:          12,
			Align:         16,
			Signed:        true,
			Floating:      true,
			HasArithmetic: true,
			HasCompare:    true,
			HasEquality:   true,
		},
		name: "f32x3",
	}
	BuiltinF32x4Type = &VectorType{
		UnderlyingType: BuiltinFloat32Type,
		Width:          4,
		properties: TypeProperties{
			Size:          16,
			Align:         16,
			Signed:        true,
			Floating:      true,
			HasArithmetic: true,
			HasCompare:    true,
			HasEquality:   true,
		},
		name: "f32x4",
	}

	BuiltinF64x2Type = &VectorType{
		UnderlyingType: BuiltinFloat32Type,
		Width:          2,
		properties: TypeProperties{
			Size:          16,
			Align:         16,
			Signed:        true,
			Floating:      true,
			HasArithmetic: true,
			HasCompare:    true,
			HasEquality:   true,
		},
		name: "f64x2",
	}
	BuiltinF64x3Type = &VectorType{
		UnderlyingType: BuiltinFloat32Type,
		Width:          3,
		properties: TypeProperties{
			Size:          24,
			Align:         16,
			Signed:        true,
			Floating:      true,
			HasArithmetic: true,
			HasCompare:    true,
			HasEquality:   true,
		},
		name: "f64x3",
	}
	BuiltinF64x4Type = &VectorType{
		UnderlyingType: BuiltinFloat32Type,
		Width:          4,
		properties: TypeProperties{
			Size:          32,
			Align:         16,
			Signed:        true,
			Floating:      true,
			HasArithmetic: true,
			HasCompare:    true,
			HasEquality:   true,
		},
		name: "f64x4",
	}

	BuiltinI32x2Type = &VectorType{
		UnderlyingType: BuiltinIntType,
		Width:          2,
		properties: TypeProperties{
			Size:          8,
			Align:         8,
			Signed:        true,
			Integral:      true,
			HasBitOps:     true,
			HasArithmetic: true,
			HasCompare:    true,
			HasEquality:   true,
		},
		name: "i32x2",
	}
	BuiltinI32x3Type = &VectorType{
		UnderlyingType: BuiltinIntType,
		Width:          3,
		properties: TypeProperties{
			Size:          12,
			Align:         16,
			Signed:        true,
			Integral:      true,
			HasBitOps:     true,
			HasArithmetic: true,
			HasCompare:    true,
			HasEquality:   true,
		},
		name: "i32x3",
	}
	BuiltinI32x4Type = &VectorType{
		UnderlyingType: BuiltinIntType,
		Width:          4,
		properties: TypeProperties{
			Size:          16,
			Align:         16,
			Signed:        true,
			Integral:      true,
			HasBitOps:     true,
			HasArithmetic: true,
			HasCompare:    true,
			HasEquality:   true,
		},
		name: "i32x4",
	}

	BuiltinU32x2Type = &VectorType{
		UnderlyingType: BuiltinUintType,
		Width:          2,
		properties: TypeProperties{
			Size:          8,
			Align:         8,
			Integral:      true,
			HasBitOps:     true,
			HasArithmetic: true,
			HasCompare:    true,
			HasEquality:   true,
		},
		name: "u32x2",
	}
	BuiltinU32x3Type = &VectorType{
		UnderlyingType: BuiltinUintType,
		Width:          3,
		properties: TypeProperties{
			Size:          12,
			Align:         16,
			Integral:      true,
			HasBitOps:     true,
			HasArithmetic: true,
			HasCompare:    true,
			HasEquality:   true,
		},
		name: "u32x3",
	}
	BuiltinU32x4Type = &VectorType{
		UnderlyingType: BuiltinUintType,
		Width:          4,
		properties: TypeProperties{
			Size:          16,
			Align:         16,
			Integral:      true,
			HasBitOps:     true,
			HasArithmetic: true,
			HasCompare:    true,
			HasEquality:   true,
		},
		name: "u32x4",
	}

	BuiltinB32x2Type = &VectorType{
		UnderlyingType: BuiltinFloat32Type,
		Width:          2,
		properties: TypeProperties{
			Size:        8,
			Align:       8,
			HasCompare:  true,
			HasEquality: true,
		},
		name: "b32x2",
	}
	BuiltinB32x3Type = &VectorType{
		UnderlyingType: BuiltinFloat32Type,
		Width:          3,
		properties: TypeProperties{
			Size:        12,
			Align:       16,
			HasCompare:  true,
			HasEquality: true,
		},
		name: "b32x3",
	}
	BuiltinB32x4Type = &VectorType{
		UnderlyingType: BuiltinFloat32Type,
		Width:          4,
		properties: TypeProperties{
			Size:        16,
			Align:       16,
			HasCompare:  true,
			HasEquality: true,
		},
		name: "b32x4",
	}
)

func (VectorType) aType() {}
func (t VectorType) Properties() TypeProperties {
	return t.properties
}
func (t VectorType) String() string  { return t.name }
func (t VectorType) HashKey() string { return t.String() }
func (t *VectorType) Resolve(bool) Type {
	return t
}
func (lhs *VectorType) Equal(rhs Type) bool {
	return lhs == rhs.Resolve(false)
}

type FuncType struct {
	ParameterTypes []Type
	ReturnTypes    []Type
}

func (FuncType) aType() {}
func (FuncType) Properties() TypeProperties {
	return TypeProperties{}
}
func (t FuncType) String() string {
	var b strings.Builder
	b.WriteString("func(")
	for i, a := range t.ParameterTypes {
		if i > 0 {
			b.WriteRune(',')
		}
		b.WriteString(a.String())
	}
	b.WriteRune(')')
	if len(t.ReturnTypes) > 0 {
		b.WriteRune('(')
		for i, a := range t.ReturnTypes {
			if i > 0 {
				b.WriteRune(',')
			}
			b.WriteString(a.String())
		}
		b.WriteRune(')')
	}
	return b.String()
}
func (t FuncType) HashKey() string {
	var b strings.Builder
	b.WriteString("func(")
	for i, a := range t.ParameterTypes {
		if i > 0 {
			b.WriteRune(',')
		}
		b.WriteString(a.HashKey())
	}
	b.WriteRune(')')
	if len(t.ReturnTypes) > 0 {
		b.WriteRune('(')
		for i, a := range t.ReturnTypes {
			if i > 0 {
				b.WriteRune(',')
			}
			b.WriteString(a.HashKey())
		}
		b.WriteRune(')')
	}
	return b.String()
}
func (t *FuncType) Resolve(bool) Type {
	return t
}
func (lhs *FuncType) Equal(rhs Type) bool {
	return lhs == rhs.Resolve(false)
}

type ArrayType struct {
	Length      int
	ElementType Type
}

func (ArrayType) aType() {}
func (t ArrayType) Properties() TypeProperties {
	return TypeProperties{
		Size:  t.ElementType.Properties().Size * t.Length,
		Align: t.ElementType.Properties().Align,
	}
}
func (t ArrayType) String() string {
	return fmt.Sprintf("[%v]%v", t.Length, t.ElementType.String())
}
func (t ArrayType) HashKey() string {
	return fmt.Sprintf("[%v]%v", t.Length, t.ElementType.HashKey())
}
func (t *ArrayType) Resolve(bool) Type {
	return t
}
func (lhs *ArrayType) Equal(rhs Type) bool {
	return lhs == rhs.Resolve(false)
}

type TupleType struct {
	Types []Type
}

func (TupleType) aType() {}
func (t TupleType) Properties() TypeProperties {
	size := 0
	align := 0
	for _, typ := range t.Types {
		size += typ.Properties().Size
		if typ.Properties().Align > align {
			align = typ.Properties().Align
		}
	}
	return TypeProperties{
		Size:  size,
		Align: align,
	}
}
func (t TupleType) String() string {
	var b strings.Builder
	b.WriteRune('(')
	for i, typ := range t.Types {
		if i > 0 {
			b.WriteRune(',')
		}
		b.WriteString(typ.String())
	}
	b.WriteRune(')')
	return b.String()
}
func (t TupleType) HashKey() string {
	var b strings.Builder
	b.WriteRune('(')
	for i, typ := range t.Types {
		if i > 0 {
			b.WriteRune(',')
		}
		b.WriteString(typ.HashKey())
	}
	b.WriteRune(')')
	return b.String()
}
func (t *TupleType) Resolve(bool) Type {
	return t
}
func (lhs *TupleType) Equal(rhs Type) bool {
	return lhs == rhs.Resolve(false)
}

type StructTypeField struct {
	Identifer *IdentifierExpr
	Type      Type
}

type StructType struct {
	Fields       []StructTypeField
	FieldsByName map[string]int
}

func (StructType) aType() {}
func (t StructType) Properties() TypeProperties {
	size, align := 0, 0
	for _, field := range t.Fields {
		fieldProperties := field.Type.Properties()
		size += fieldProperties.Size
		if align < fieldProperties.Align {
			align = fieldProperties.Align
		}
	}
	return TypeProperties{
		Size:  size,
		Align: align,
	}
}
func (t StructType) String() string {
	var b strings.Builder
	b.WriteString("struct{")
	for i, field := range t.Fields {
		if i > 0 {
			b.WriteRune(',')
		}
		b.WriteString(field.Type.String())
	}
	b.WriteRune('}')
	return b.String()
}
func (t StructType) HashKey() string {
	var b strings.Builder
	b.WriteString("struct{")
	for i, field := range t.Fields {
		if i > 0 {
			b.WriteRune(',')
		}
		b.WriteString(field.Type.HashKey())
	}
	b.WriteRune('}')
	return b.String()
}
func (t StructType) FindField(name string) *StructTypeField {
	if index, ok := t.FieldsByName[name]; ok {
		return &t.Fields[index]
	}
	for _, field := range t.Fields {
		if field.Identifer != nil {
			continue
		}

		strongAlias, ok := field.Type.(*StrongAliasType)
		if !ok {
			continue
		}

		underlyingType := strongAlias.Resolve(true)
		structType, ok := underlyingType.(*StructType)
		if !ok {
			continue
		}

		if field := structType.FindField(name); field != nil {
			return field
		}
	}
	return nil
}
func (t *StructType) Resolve(bool) Type {
	return t
}
func (lhs *StructType) Equal(rhs Type) bool {
	return lhs == rhs.Resolve(false)
}

type StrongAliasType struct {
	Name           string
	UnderlyingType Type
}

func (StrongAliasType) aType() {}
func (t StrongAliasType) Properties() TypeProperties {
	return t.UnderlyingType.Properties()
}
func (t StrongAliasType) String() string  { return t.Name }
func (t StrongAliasType) HashKey() string { return t.String() }
func (t *StrongAliasType) Resolve(resolveStrongAlias bool) Type {
	return resolveAlias(t, resolveStrongAlias)
}
func (lhs *StrongAliasType) Equal(rhs Type) bool {
	return lhs == rhs.Resolve(false)
}

type WeakAliasType struct {
	Name           string
	UnderlyingType Type
}

func (WeakAliasType) aType() {}
func (t WeakAliasType) Properties() TypeProperties {
	return t.UnderlyingType.Properties()
}
func (t WeakAliasType) String() string  { return fmt.Sprintf("%v=%v", t.Name, t.UnderlyingType) }
func (t WeakAliasType) HashKey() string { return t.String() }
func (t *WeakAliasType) Resolve(resolveStrongAlias bool) Type {
	return resolveAlias(t, resolveStrongAlias)
}
func (lhs *WeakAliasType) Equal(rhs Type) bool {
	return lhs.Resolve(false) == rhs.Resolve(false)
}

func resolveAlias(alias Type, resolveStrongAlias bool) Type {
	switch t := alias.(type) {
	case *StrongAliasType:
		if resolveStrongAlias {
			return resolveAlias(t.UnderlyingType, resolveStrongAlias)
		} else {
			return t
		}
	case *WeakAliasType:
		return resolveAlias(t.UnderlyingType, resolveStrongAlias)
	default:
		return t
	}
}

type TypeInterner struct {
	types map[string]Type
}

func NewTypeInterner() *TypeInterner {
	return &TypeInterner{
		types: make(map[string]Type),
	}
}

func (t *TypeInterner) InternFuncType(args []Type, returns []Type) Type {
	funcType := FuncType{
		ParameterTypes: args,
		ReturnTypes:    returns,
	}
	key := funcType.HashKey()

	if v, ok := t.types[key]; ok {
		return v
	}

	t.types[key] = &funcType
	return &funcType
}

func (t *TypeInterner) InternArrayType(length int, elementType Type) Type {
	arrayType := ArrayType{
		Length:      length,
		ElementType: elementType,
	}
	key := arrayType.HashKey()

	if v, ok := t.types[key]; ok {
		return v
	}

	t.types[key] = &arrayType
	return &arrayType
}

func (t *TypeInterner) InternTupleType(types []Type) Type {
	tupleType := TupleType{
		Types: types,
	}
	key := tupleType.HashKey()

	if v, ok := t.types[key]; ok {
		return v
	}

	t.types[key] = &tupleType
	return &tupleType
}

func (t *TypeInterner) InternStructType(names []string, types []StructTypeField) Type {
	structType := StructType{
		Fields:       types,
		FieldsByName: make(map[string]int),
	}
	for i := 0; i < len(names); i++ {
		structType.FieldsByName[names[i]] = i
	}
	key := structType.HashKey()

	if v, ok := t.types[key]; ok {
		return v
	}

	t.types[key] = &structType
	return &structType
}

func (t *TypeInterner) InternStrongTypeAlias(name string, underlyingType Type) *StrongAliasType {
	return &StrongAliasType{
		Name:           name,
		UnderlyingType: underlyingType,
	}
}

func (t *TypeInterner) InternWeakTypeAlias(name string, underlyingType Type) *WeakAliasType {
	return &WeakAliasType{
		Name:           name,
		UnderlyingType: underlyingType,
	}
}

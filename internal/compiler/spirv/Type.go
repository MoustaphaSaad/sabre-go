package spirv

import (
	"fmt"
	"strings"
)

type Type interface {
	Object
	aType()
	TypeName() string
	HashKey() string
}

type VoidType struct {
	ObjectID   ID
	ObjectName string
	Module     *Module
}

func (t VoidType) ID() ID {
	return t.ObjectID
}
func (t VoidType) Name() string {
	return t.ObjectName
}
func (VoidType) aType() {}
func (t VoidType) TypeName() string {
	return "void"
}
func (t VoidType) HashKey() string {
	return t.TypeName()
}

type BoolType struct {
	ObjectID   ID
	ObjectName string
	Module     *Module
}

func (t BoolType) ID() ID {
	return t.ObjectID
}
func (t BoolType) Name() string {
	return t.ObjectName
}
func (BoolType) aType() {}
func (t BoolType) TypeName() string {
	return "bool"
}
func (t BoolType) HashKey() string {
	return t.TypeName()
}

type IntType struct {
	ObjectID   ID
	ObjectName string
	Module     *Module
	BitWidth   int
	IsSigned   bool
}

func (t IntType) ID() ID {
	return t.ObjectID
}
func (t IntType) Name() string {
	return t.ObjectName
}
func (IntType) aType() {}
func (t IntType) TypeName() string {
	if t.IsSigned {
		return fmt.Sprintf("int%d", t.BitWidth)
	} else {
		return fmt.Sprintf("uint%d", t.BitWidth)
	}
}
func (t IntType) HashKey() string {
	return t.TypeName()
}

type FloatType struct {
	ObjectID   ID
	ObjectName string
	Module     *Module
	BitWidth   int
}

func (t FloatType) ID() ID {
	return t.ObjectID
}
func (t FloatType) Name() string {
	return t.ObjectName
}
func (FloatType) aType() {}
func (t FloatType) TypeName() string {
	return fmt.Sprintf("float%d", t.BitWidth)
}
func (t FloatType) HashKey() string {
	return t.TypeName()
}

type PtrType struct {
	ObjectID     ID
	ObjectName   string
	Module       *Module
	To           Type
	StorageClass StorageClass
}

func (t PtrType) ID() ID {
	return t.ObjectID
}
func (t PtrType) Name() string {
	return t.ObjectName
}
func (PtrType) aType() {}
func (t PtrType) TypeName() string {
	return fmt.Sprintf("ptr_%s_%d", t.To.TypeName(), t.StorageClass)
}
func (t PtrType) HashKey() string {
	return fmt.Sprintf("ptr(%s,%d)", t.To.HashKey(), t.StorageClass)
}

type FuncType struct {
	ObjectID   ID
	ObjectName string
	Module     *Module
	ReturnType Type
	ArgTypes   []Type
}

func (t FuncType) ID() ID {
	return t.ObjectID
}
func (t FuncType) Name() string {
	return t.ObjectName
}
func (FuncType) aType() {}
func (t FuncType) TypeName() string {
	var b strings.Builder
	b.WriteString("func")
	for _, arg := range t.ArgTypes {
		b.WriteString("_")
		b.WriteString(arg.TypeName())
	}
	b.WriteString("_ret_")
	b.WriteString(t.ReturnType.TypeName())
	return b.String()
}
func (t FuncType) HashKey() string {
	var b strings.Builder
	b.WriteString("func(")
	for i, arg := range t.ArgTypes {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(arg.HashKey())
	}
	b.WriteString(")")
	b.WriteString(t.ReturnType.HashKey())
	return b.String()
}

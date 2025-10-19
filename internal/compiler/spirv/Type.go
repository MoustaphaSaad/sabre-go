package spirv

import (
	"fmt"
	"strings"
)

type Type interface {
	Object
	aType()
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
func (t VoidType) HashKey() string {
	return "void"
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
func (t BoolType) HashKey() string {
	return "bool"
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
func (t IntType) HashKey() string {
	if t.IsSigned {
		return fmt.Sprintf("int%d", t.BitWidth)
	} else {
		return fmt.Sprintf("uint%d", t.BitWidth)
	}
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

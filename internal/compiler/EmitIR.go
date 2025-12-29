package compiler

import (
	"fmt"
	"go/constant"

	"github.com/MoustaphaSaad/sabre-go/internal/compiler/spirv"
)

type IREmitter struct {
	unit           *Unit
	module         *spirv.Module
	objectBySymbol map[Symbol]spirv.Object
	blockStack     []*spirv.Block
}

func NewIREmitter(u *Unit) *IREmitter {
	return &IREmitter{
		unit: u,
		// we set the addressing and memory model to default values for now
		module:         spirv.NewModule(spirv.AddressingModelLogical, spirv.MemoryModelGLSL450),
		objectBySymbol: make(map[Symbol]spirv.Object),
		blockStack:     make([]*spirv.Block, 0),
	}
}

func (ir *IREmitter) enterBlock(block *spirv.Block) {
	ir.blockStack = append(ir.blockStack, block)
}

func (ir *IREmitter) leaveBlock() {
	if len(ir.blockStack) > 0 {
		ir.blockStack = ir.blockStack[:len(ir.blockStack)-1]
	}
}

func (ir *IREmitter) currentBlock() *spirv.Block {
	if len(ir.blockStack) > 0 {
		return ir.blockStack[len(ir.blockStack)-1]
	}
	return nil
}

func (ir *IREmitter) objectOfSymbol(sym Symbol) spirv.Object {
	if obj, ok := ir.objectBySymbol[sym]; ok {
		return obj
	}
	return nil
}

func (ir *IREmitter) setObjectOfSymbol(sym Symbol, obj spirv.Object) {
	ir.objectBySymbol[sym] = obj
}

func (ir *IREmitter) Emit() *spirv.Module {
	// we add this hardcoded capabilities for now
	ir.module.AddCapability(spirv.CapabilityShader)
	ir.module.AddCapability(spirv.CapabilityLinkage)

	for _, sym := range ir.unit.semanticInfo.ReachableSymbols {
		ir.emitSymbol(sym)
	}

	RewriteIR(ir.module)

	return ir.module
}

func (ir *IREmitter) emitSymbol(sym Symbol) {
	var obj spirv.Object
	switch s := sym.(type) {
	case *FuncSymbol:
		obj = ir.emitFunc(s)
	default:
		panic("unsupported symbol")
	}
	ir.setObjectOfSymbol(sym, obj)
}

func (ir *IREmitter) emitFunc(sym *FuncSymbol) spirv.Object {
	paramSymbols := func() (syms []Symbol) {
		funcDecl := sym.Decl().(*FuncDecl)
		for _, f := range funcDecl.Type.Parameters.Fields {
			if len(f.Names) == 0 {
				syms = append(syms, nil)
			} else {
				for _, idExpr := range f.Names {
					syms = append(syms, ir.unit.semanticInfo.SymbolOfIdentifier(idExpr))
				}
			}
		}
		return
	}()

	funcType := ir.unit.semanticInfo.TypeOf(sym).Type.(*FuncType)
	spirvFuncType := ir.emitType(funcType).(*spirv.FuncType)

	if len(paramSymbols) != len(spirvFuncType.ArgTypes) {
		panic(fmt.Sprintf(
			"Function parameters names count (%v) mismatches arguments types count (%v)",
			len(paramSymbols),
			len(spirvFuncType.ArgTypes)),
		)
	}
	params := make([]*spirv.FuncParam, len(spirvFuncType.ArgTypes))
	for i := range spirvFuncType.ArgTypes {
		if paramSymbols[i] == nil {
			params[i] = ir.module.NewFuncParam(fmt.Sprintf("UnnamedParam%v", i), spirvFuncType.ArgTypes[i])
		} else {
			params[i] = ir.module.NewFuncParam(paramSymbols[i].Name(), spirvFuncType.ArgTypes[i])
			ir.setObjectOfSymbol(paramSymbols[i], params[i])
		}
	}
	spirvFunction := ir.module.NewFunction(sym.Name(), spirvFuncType, params)

	funcDecl := sym.Decl().(*FuncDecl)
	if funcDecl.Body == nil {
		return spirvFunction
	}

	spirvBlock := spirvFunction.NewBlock(fmt.Sprintf("entry_%v", sym.Name()))
	ir.enterBlock(spirvBlock)
	defer ir.leaveBlock()

	ir.emitStatement(funcDecl.Body)

	return spirvFunction
}

func (ir *IREmitter) emitVar(symbol *VarSymbol, sc spirv.StorageClass, initExpr Expr) spirv.Object {
	tav := ir.unit.semanticInfo.TypeOf(symbol)
	spirvType := ir.emitType(tav.Type)
	ptrType := ir.module.InternPtr(spirvType, sc)
	variable := ir.module.NewVariable(symbol.Name(), ptrType, sc)

	var initValueID spirv.ID
	if initTAV := symbol.InitTypeAndValue; initTAV != nil && initTAV.Mode == AddressModeConstant {
		initValueID = ir.emitConstantValue(initTAV).ID()
	}

	block := ir.currentBlock()
	block.Push(&spirv.VariableInstruction{
		ResultType:   variable.Type.ID(),
		ResultID:     variable.ID(),
		StorageClass: variable.StorageClass,
		Initializer:  initValueID,
	})

	ir.setObjectOfSymbol(symbol, variable)

	if initExpr != nil {
		switch rhsExpr := initExpr.(type) {
		case *LiteralExpr:
		case *IdentifierExpr:
		default:
			block.Push(&spirv.StoreInstruction{
				Pointer: variable.ID(),
				Object:  ir.emitExpression(rhsExpr).ID(),
			})
		}
	}

	return variable
}

func (ir *IREmitter) emitExpression(expr Expr) spirv.Object {
	switch e := expr.(type) {
	case *LiteralExpr:
		return ir.emitLiteralExpr(e)
	case *IdentifierExpr:
		return ir.emitIdentifierExpr(e)
	case *UnaryExpr:
		return ir.emitUnaryExpr(e)
	case *BinaryExpr:
		return ir.emitBinaryExpr(e)
	case *CallExpr:
		return ir.emitCallExpr(e)
	case *ParenExpr:
		return ir.emitExpression(e.Base)
	default:
		panic("unsupported expression")
	}
}

func (ir *IREmitter) emitLiteralExpr(e *LiteralExpr) spirv.Object {
	tav := ir.unit.semanticInfo.TypeOf(e)
	return ir.emitConstantValue(tav)
}

func (ir *IREmitter) emitConstantValue(tav *TypeAndValue) spirv.Object {
	switch t := ir.emitType(tav.Type).(type) {
	case *spirv.BoolType:
		val := constant.BoolVal(tav.Value)
		return ir.module.InternBoolConstant(val, t)
	case *spirv.IntType:
		val, _ := constant.Int64Val(tav.Value)
		return ir.module.InternIntConstant(val, t)
	case *spirv.FloatType:
		val, _ := constant.Float64Val(tav.Value)
		return ir.module.InternFloatConstant(val, t)
	default:
		panic("unsupported literal type")
	}
}

func (ir *IREmitter) emitIdentifierExpr(e *IdentifierExpr) spirv.Object {
	symbol := ir.unit.semanticInfo.SymbolOfIdentifier(e)
	if symbol == nil {
		panic(fmt.Sprintf("unable to find symbol for identifier: %v", e.Token.Value()))
	}

	obj := ir.objectOfSymbol(symbol)

	if variable, ok := obj.(*spirv.Variable); ok {
		tav := ir.unit.semanticInfo.TypeOf(e)
		resultType := ir.emitType(tav.Type)
		loadedValue := ir.module.NewValue(resultType)
		block := ir.currentBlock()
		block.Push(&spirv.LoadInstruction{
			ResultType: resultType.ID(),
			ResultID:   loadedValue.ID(),
			Pointer:    variable.ID(),
		})
		return loadedValue
	}

	return obj
}

func (ir *IREmitter) emitUnaryExpr(e *UnaryExpr) spirv.Object {
	base := ir.emitExpression(e.Base)
	tav := ir.unit.semanticInfo.TypeOf(e)
	resultType := ir.emitType(tav.Type)
	result := ir.module.NewValue(resultType)
	block := ir.currentBlock()

	switch e.Operator.Kind() {
	case TokenAdd:
		// Unary + is a no-op, just return the base value
		return base
	case TokenSub:
		// Unary - requires negation
		props := tav.Type.Properties()
		if props.Floating {
			// Use FNegate for floating-point types
			block.Push(&spirv.FNegateInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand:    base.ID(),
			})
		} else if props.Integral {
			// Use SNegate for signed integer types
			block.Push(&spirv.SNegateInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand:    base.ID(),
			})
		} else {
			panic("unsupported type for unary minus")
		}
		return result
	case TokenNot:
		// Logical NOT for boolean types
		block.Push(&spirv.LogicalNotInstruction{
			ResultType: resultType.ID(),
			ResultID:   result.ID(),
			Operand:    base.ID(),
		})
		return result
	case TokenXor:
		// Bitwise NOT (complement) for integer types
		block.Push(&spirv.NotInstruction{
			ResultType: resultType.ID(),
			ResultID:   result.ID(),
			Operand:    base.ID(),
		})
		return result
	default:
		panic("unsupported unary operator")
	}
}

func (ir *IREmitter) emitBinaryExpr(e *BinaryExpr) spirv.Object {
	lhs := ir.emitExpression(e.LHS)
	rhs := ir.emitExpression(e.RHS)
	tav := ir.unit.semanticInfo.TypeOf(e)
	resultType := ir.emitType(tav.Type)
	result := ir.module.NewValue(resultType)
	block := ir.currentBlock()

	switch e.Operator.Kind() {
	case TokenLOr:
		// Logical OR - only for boolean types
		block.Push(&spirv.LogicalOrInstruction{
			ResultType: resultType.ID(),
			ResultID:   result.ID(),
			Operand1:   lhs.ID(),
			Operand2:   rhs.ID(),
		})
		return result
	case TokenLAnd:
		// Logical AND - only for boolean types
		block.Push(&spirv.LogicalAndInstruction{
			ResultType: resultType.ID(),
			ResultID:   result.ID(),
			Operand1:   lhs.ID(),
			Operand2:   rhs.ID(),
		})
		return result
	case TokenLT:
		// Less than comparison - need to check operand types
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Floating {
			// Floating-point ordered less than
			block.Push(&spirv.FOrdLessThanInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else if props.Integral {
			if props.Signed {
				// Signed integer less than
				block.Push(&spirv.SLessThanInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Operand1:   lhs.ID(),
					Operand2:   rhs.ID(),
				})
			} else {
				// Unsigned integer less than
				block.Push(&spirv.ULessThanInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Operand1:   lhs.ID(),
					Operand2:   rhs.ID(),
				})
			}
		} else {
			panic("unsupported type for less than comparison")
		}
		return result
	case TokenGT:
		// Greater than comparison - need to check operand types
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Floating {
			// Floating-point ordered greater than
			block.Push(&spirv.FOrdGreaterThanInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else if props.Integral {
			if props.Signed {
				// Signed integer greater than
				block.Push(&spirv.SGreaterThanInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Operand1:   lhs.ID(),
					Operand2:   rhs.ID(),
				})
			} else {
				// Unsigned integer greater than
				block.Push(&spirv.UGreaterThanInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Operand1:   lhs.ID(),
					Operand2:   rhs.ID(),
				})
			}
		} else {
			panic("unsupported type for greater than comparison")
		}
		return result
	case TokenLE:
		// Less than or equal comparison - need to check operand types
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Floating {
			// Floating-point ordered less than or equal
			block.Push(&spirv.FOrdLessThanEqualInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else if props.Integral {
			if props.Signed {
				// Signed integer less than or equal
				block.Push(&spirv.SLessThanEqualInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Operand1:   lhs.ID(),
					Operand2:   rhs.ID(),
				})
			} else {
				// Unsigned integer less than or equal
				block.Push(&spirv.ULessThanEqualInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Operand1:   lhs.ID(),
					Operand2:   rhs.ID(),
				})
			}
		} else {
			panic("unsupported type for less than or equal comparison")
		}
		return result

	case TokenGE:
		// Greater than or equal comparison - need to check operand types
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Floating {
			// Floating-point ordered greater than or equal
			block.Push(&spirv.FOrdGreaterThanEqualInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else if props.Integral {
			if props.Signed {
				// Signed integer greater than or equal
				block.Push(&spirv.SGreaterThanEqualInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Operand1:   lhs.ID(),
					Operand2:   rhs.ID(),
				})
			} else {
				// Unsigned integer greater than or equal
				block.Push(&spirv.UGreaterThanEqualInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Operand1:   lhs.ID(),
					Operand2:   rhs.ID(),
				})
			}
		} else {
			panic("unsupported type for greater than or equal comparison")
		}
		return result

	case TokenEQ:
		// Equality comparison - need to check operand types
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Floating {
			// Floating-point ordered equal
			block.Push(&spirv.FOrdEqualInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else if props.Integral {
			// Integer equal (same for signed and unsigned)
			block.Push(&spirv.IEqualInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else if props.HasEquality && props.HasLogicOps {
			// Bool equal (logical equal)
			block.Push(&spirv.LogicalEqualInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else {
			panic("unsupported type for equality comparison")
		}
		return result

	case TokenNE:
		// Not equal comparison - need to check operand types
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Floating {
			// Floating-point ordered not equal
			block.Push(&spirv.FOrdNotEqualInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else if props.Integral {
			// Integer not equal (same for signed and unsigned)
			block.Push(&spirv.INotEqualInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else if props.HasEquality && props.HasLogicOps {
			// Bool not equal (logical not equal)
			block.Push(&spirv.LogicalNotEqualInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else {
			panic("unsupported type for not equal comparison")
		}
		return result

	case TokenAdd:
		// Addition - need to check operand types
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Floating {
			// Floating-point addition
			block.Push(&spirv.FAddInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else if props.Integral {
			// Integer addition (same for signed and unsigned)
			block.Push(&spirv.IAddInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else {
			panic("unsupported type for addition")
		}
		return result

	case TokenSub:
		// Subtraction - need to check operand types
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Floating {
			// Floating-point subtraction
			block.Push(&spirv.FSubInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else if props.Integral {
			// Integer subtraction (same for signed and unsigned)
			block.Push(&spirv.ISubInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else {
			panic("unsupported type for subtraction")
		}
		return result

	case TokenXor:
		// Bitwise XOR - only for integers
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Integral {
			// Bitwise XOR (same for signed and unsigned)
			block.Push(&spirv.BitwiseXorInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else {
			panic("unsupported type for bitwise XOR")
		}
		return result

	case TokenOr:
		// Bitwise OR - only for integers
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Integral {
			// Bitwise OR (same for signed and unsigned)
			block.Push(&spirv.BitwiseOrInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else {
			panic("unsupported type for bitwise OR")
		}
		return result

	case TokenMul:
		// Multiplication - need to check operand types
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Floating {
			// Floating-point multiplication
			block.Push(&spirv.FMulInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else if props.Integral {
			// Integer multiplication (same for signed and unsigned)
			block.Push(&spirv.IMulInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else {
			panic("unsupported type for multiplication")
		}
		return result

	case TokenDiv:
		// Division - need to check operand types
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Floating {
			// Floating-point division
			block.Push(&spirv.FDivInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else if props.Integral {
			if props.Signed {
				// Signed integer division
				block.Push(&spirv.SDivInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Operand1:   lhs.ID(),
					Operand2:   rhs.ID(),
				})
			} else {
				// Unsigned integer division
				block.Push(&spirv.UDivInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Operand1:   lhs.ID(),
					Operand2:   rhs.ID(),
				})
			}
		} else {
			panic("unsupported type for division")
		}
		return result

	case TokenMod:
		// Modulo/Remainder - need to check operand types
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Floating {
			// Floating-point remainder
			block.Push(&spirv.FRemInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else if props.Integral {
			if props.Signed {
				// Signed integer remainder
				block.Push(&spirv.SRemInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Operand1:   lhs.ID(),
					Operand2:   rhs.ID(),
				})
			} else {
				// Unsigned integer modulo
				block.Push(&spirv.UModInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Operand1:   lhs.ID(),
					Operand2:   rhs.ID(),
				})
			}
		} else {
			panic("unsupported type for modulo")
		}
		return result

	case TokenAnd:
		// Bitwise AND - only for integers
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Integral {
			// Bitwise AND (same for signed and unsigned)
			block.Push(&spirv.BitwiseAndInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   rhs.ID(),
			})
		} else {
			panic("unsupported type for bitwise AND")
		}
		return result

	case TokenAndNot:
		// Bitwise AND NOT - only for integers
		// In Go, a &^ b is equivalent to a & (^b) - clear bits in a that are set in b
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Integral {
			// First, compute NOT of rhs: ^b
			notRhs := ir.module.NewValue(resultType)
			block.Push(&spirv.NotInstruction{
				ResultType: resultType.ID(),
				ResultID:   notRhs.ID(),
				Operand:    rhs.ID(),
			})

			// Then, compute AND with lhs: a & (^b)
			block.Push(&spirv.BitwiseAndInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Operand1:   lhs.ID(),
				Operand2:   notRhs.ID(),
			})
		} else {
			panic("unsupported type for bitwise AND NOT")
		}
		return result

	case TokenShl:
		// Shift left logical - only for integers
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Integral {
			// Shift left logical
			block.Push(&spirv.ShiftLeftLogicalInstruction{
				ResultType: resultType.ID(),
				ResultID:   result.ID(),
				Base:       lhs.ID(),
				Shift:      rhs.ID(),
			})
		} else {
			panic("unsupported type for shift left")
		}
		return result

	case TokenShr:
		// Shift right - only for integers
		// Use arithmetic shift for signed integers, logical shift for unsigned
		lhsType := ir.unit.semanticInfo.TypeOf(e.LHS).Type
		props := lhsType.Properties()

		if props.Integral {
			if props.Signed {
				// Arithmetic shift right (preserves sign bit)
				block.Push(&spirv.ShiftRightArithmeticInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Base:       lhs.ID(),
					Shift:      rhs.ID(),
				})
			} else {
				// Logical shift right (shifts in zeros)
				block.Push(&spirv.ShiftRightLogicalInstruction{
					ResultType: resultType.ID(),
					ResultID:   result.ID(),
					Base:       lhs.ID(),
					Shift:      rhs.ID(),
				})
			}
		} else {
			panic("unsupported type for shift right")
		}
		return result

	default:
		panic("unsupported binary operator")
	}
}

func (ir *IREmitter) emitCallExpr(e *CallExpr) spirv.Object {
	base := ir.emitExpression(e.Base)
	args := make([]spirv.ID, len(e.Args))
	for i, argExpr := range e.Args {
		emittedExpr := ir.emitExpression(argExpr)
		args[i] = emittedExpr.ID()
	}

	block := ir.currentBlock()

	tav := ir.unit.semanticInfo.TypeOf(e.Base)
	funcType := tav.Type.(*FuncType)

	// TODO: Handle multiple return types
	var resultType spirv.Type
	if len(funcType.ReturnTypes) > 0 {
		resultType = ir.emitType(funcType.ReturnTypes[0])
	} else {
		resultType = ir.module.InternVoid()
	}

	resultValue := ir.module.NewValue(resultType)
	block.Push(&spirv.FunctionCallInstruction{
		ResultType: resultType.ID(),
		ResultID:   resultValue.ID(),
		FunctionID: base.ID(),
		Args:       args,
	})
	return resultValue
}

func (ir *IREmitter) emitType(Type Type) spirv.Type {
	switch t := Type.(type) {
	case *VoidType:
		return ir.module.InternVoid()
	case *BoolType:
		return ir.module.InternBool()
	case *IntType:
		return ir.module.InternInt(32, t.Properties().Signed)
	case *Float32Type:
		return ir.module.InternFloat(32)
	case *Float64Type:
		return ir.module.InternFloat(64)
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

func (ir *IREmitter) emitStatement(stmt Stmt) {
	switch s := stmt.(type) {
	case *ReturnStmt:
		ir.emitReturnStmt(s)
	case *IncDecStmt:
		ir.emitIncDecStmt(s)
	case *ExprStmt:
		ir.emitExpression(s.Expr)
	case *DeclStmt:
		ir.emitDeclStmt(s)
	case *BlockStmt:
		ir.emitBlockStmt(s)
	case *AssignStmt:
		ir.emitAssignStmt(s)
	case *IfStmt:
		ir.emitIfStmt(s)
	default:
		panic("unsupported statement")
	}
}

func (ir *IREmitter) emitReturnStmt(s *ReturnStmt) {
	block := ir.currentBlock()
	if len(s.Exprs) > 0 {
		// TODO: Multiple return values
		block.Push(&spirv.ReturnValueInstruction{Value: ir.emitExpression(s.Exprs[0]).ID()})
	} else {
		block.Push(&spirv.ReturnInstruction{})
	}
	ir.leaveBlock()
	newBlock := block.Function.NewBlock(block.Function.Name())
	ir.enterBlock(newBlock)
}

func (ir *IREmitter) emitIncDecStmt(s *IncDecStmt) {
	tav := ir.unit.semanticInfo.TypeOf(s.Expr)
	resultType := ir.emitType(tav.Type)

	oneConst := func() spirv.Object {
		switch t := resultType.(type) {
		case *spirv.IntType:
			return ir.module.InternIntConstant(1, t)
		case *spirv.FloatType:
			return ir.module.InternFloatConstant(1.0, t)
		default:
			panic("unsupported type for inc/dec")
		}
	}()

	currentBlock := ir.currentBlock()

	symbol := ir.unit.semanticInfo.SymbolOfIdentifier(s.Expr.(*IdentifierExpr))
	obj := ir.objectOfSymbol(symbol)

	loadedValue := ir.module.NewValue(resultType)
	currentBlock.Push(&spirv.LoadInstruction{
		ResultType: resultType.ID(),
		ResultID:   loadedValue.ID(),
		Pointer:    obj.ID(),
	})
	resultValue := ir.module.NewValue(resultType)
	if s.Operator.Kind() == TokenInc {
		switch resultType.(type) {
		case *spirv.IntType:
			currentBlock.Push(&spirv.IAddInstruction{
				ResultType: resultType.ID(),
				ResultID:   resultValue.ID(),
				Operand1:   loadedValue.ID(),
				Operand2:   oneConst.ID(),
			})
		case *spirv.FloatType:
			currentBlock.Push(&spirv.FAddInstruction{
				ResultType: resultType.ID(),
				ResultID:   resultValue.ID(),
				Operand1:   loadedValue.ID(),
				Operand2:   oneConst.ID(),
			})
		default:
			panic("unsupported type for increment")
		}
	} else {
		switch resultType.(type) {
		case *spirv.IntType:
			currentBlock.Push(&spirv.ISubInstruction{
				ResultType: resultType.ID(),
				ResultID:   resultValue.ID(),
				Operand1:   loadedValue.ID(),
				Operand2:   oneConst.ID(),
			})
		case *spirv.FloatType:
			currentBlock.Push(&spirv.FSubInstruction{
				ResultType: resultType.ID(),
				ResultID:   resultValue.ID(),
				Operand1:   loadedValue.ID(),
				Operand2:   oneConst.ID(),
			})
		default:
			panic("unsupported type for decrement")
		}
	}

	currentBlock.Push(&spirv.StoreInstruction{
		Pointer: obj.ID(),
		Object:  resultValue.ID(),
	})
}

func (ir *IREmitter) emitDeclStmt(s *DeclStmt) {
	d := s.Decl.(*GenericDecl)
	switch d.DeclToken.Kind() {
	case TokenVar:
		ir.emitVarDecl(d, spirv.StorageClassFunction)
	default:
		panic("unsupported declaration in DeclStmt")
	}
}

func (ir *IREmitter) emitVarDecl(d *GenericDecl, sc spirv.StorageClass) {
	for _, spec := range d.Specs {
		v := spec.(*ValueSpec)
		for i, name := range v.LHS {
			symbol := ir.unit.semanticInfo.SymbolOfIdentifier(name).(*VarSymbol)
			var initExpr Expr = nil
			if v.RHS != nil {
				if i < len(v.RHS) {
					initExpr = v.RHS[i]
				} else {
					panic("variable initialization from tuple types is not supported yet")
				}
			}

			if i != symbol.ExprIndex && symbol.ExprIndex != -1 {
				panic(fmt.Sprintf("LHS index %v mismatches ExprIndex %v", i, symbol.ExprIndex))
			}

			ir.emitVar(symbol, sc, initExpr)
		}
	}
}

func (ir *IREmitter) emitBlockStmt(block *BlockStmt) {
	for _, s := range block.Stmts {
		ir.emitStatement(s)
	}
}

func (ir *IREmitter) emitAssignStmt(s *AssignStmt) {
	switch s.Operator.Kind() {
	case TokenColonAssign:
		for i, lhsExpr := range s.LHS {
			symbol := ir.unit.semanticInfo.SymbolOfIdentifier(lhsExpr.(*IdentifierExpr)).(*VarSymbol)

			var initExpr Expr = nil
			if s.RHS != nil {
				if i < len(s.RHS) {
					initExpr = s.RHS[i]
				} else {
					panic("variable initialization from tuple types is not supported yet")
				}
			}

			ir.emitVar(symbol, spirv.StorageClassFunction, initExpr)
		}
	case TokenAssign:
		for i, lhsExpr := range s.LHS {
			symbol := ir.unit.semanticInfo.SymbolOfIdentifier(lhsExpr.(*IdentifierExpr)).(*VarSymbol)
			obj := ir.objectOfSymbol(symbol)
			block := ir.currentBlock()
			block.Push(&spirv.StoreInstruction{
				Pointer: obj.ID(),
				Object:  ir.emitExpression(s.RHS[i]).ID(),
			})
		}
	case TokenAddAssign, TokenSubAssign, TokenMulAssign, TokenDivAssign, TokenAndAssign,
		TokenAndNotAssign, TokenOrAssign, TokenXorAssign, TokenShlAssign, TokenShrAssign:
		for i, lhsExpr := range s.LHS {
			symbol := ir.unit.semanticInfo.SymbolOfIdentifier(lhsExpr.(*IdentifierExpr)).(*VarSymbol)
			obj := ir.objectOfSymbol(symbol)
			t := obj.(*spirv.Variable).Type.To
			loadedValue := ir.module.NewValue(t)
			block := ir.currentBlock()
			block.Push(&spirv.LoadInstruction{
				ResultType: loadedValue.Type.ID(),
				ResultID:   loadedValue.ID(),
				Pointer:    obj.ID(),
			})
			rhsValue := ir.emitExpression(s.RHS[i])
			resultValue := ir.module.NewValue(loadedValue.Type)
			switch s.Operator.Kind() {
			case TokenAddAssign:
				switch t.(type) {
				case *spirv.IntType:
					block.Push(&spirv.IAddInstruction{
						ResultType: resultValue.Type.ID(),
						ResultID:   resultValue.ID(),
						Operand1:   loadedValue.ID(),
						Operand2:   rhsValue.ID(),
					})
				case *spirv.FloatType:
					block.Push(&spirv.FAddInstruction{
						ResultType: resultValue.Type.ID(),
						ResultID:   resultValue.ID(),
						Operand1:   loadedValue.ID(),
						Operand2:   rhsValue.ID(),
					})
				}
			case TokenSubAssign:
				switch t.(type) {
				case *spirv.IntType:
					block.Push(&spirv.ISubInstruction{
						ResultType: resultValue.Type.ID(),
						ResultID:   resultValue.ID(),
						Operand1:   loadedValue.ID(),
						Operand2:   rhsValue.ID(),
					})
				case *spirv.FloatType:
					block.Push(&spirv.FSubInstruction{
						ResultType: resultValue.Type.ID(),
						ResultID:   resultValue.ID(),
						Operand1:   loadedValue.ID(),
						Operand2:   rhsValue.ID(),
					})
				}
			case TokenMulAssign:
				switch t.(type) {
				case *spirv.IntType:
					block.Push(&spirv.IMulInstruction{
						ResultType: resultValue.Type.ID(),
						ResultID:   resultValue.ID(),
						Operand1:   loadedValue.ID(),
						Operand2:   rhsValue.ID(),
					})
				case *spirv.FloatType:
					block.Push(&spirv.FMulInstruction{
						ResultType: resultValue.Type.ID(),
						ResultID:   resultValue.ID(),
						Operand1:   loadedValue.ID(),
						Operand2:   rhsValue.ID(),
					})
				}
			case TokenDivAssign:
				switch t.(type) {
				case *spirv.IntType:
					block.Push(&spirv.SDivInstruction{
						ResultType: resultValue.Type.ID(),
						ResultID:   resultValue.ID(),
						Operand1:   loadedValue.ID(),
						Operand2:   rhsValue.ID(),
					})
				case *spirv.FloatType:
					block.Push(&spirv.FDivInstruction{
						ResultType: resultValue.Type.ID(),
						ResultID:   resultValue.ID(),
						Operand1:   loadedValue.ID(),
						Operand2:   rhsValue.ID(),
					})
				}
			case TokenAndAssign:
				block.Push(&spirv.BitwiseAndInstruction{
					ResultType: resultValue.Type.ID(),
					ResultID:   resultValue.ID(),
					Operand1:   loadedValue.ID(),
					Operand2:   rhsValue.ID(),
				})
			case TokenAndNotAssign:
				notRhs := ir.module.NewValue(loadedValue.Type)
				block.Push(&spirv.NotInstruction{
					ResultType: notRhs.Type.ID(),
					ResultID:   notRhs.ID(),
					Operand:    rhsValue.ID(),
				})
				block.Push(&spirv.BitwiseAndInstruction{
					ResultType: resultValue.Type.ID(),
					ResultID:   resultValue.ID(),
					Operand1:   loadedValue.ID(),
					Operand2:   notRhs.ID(),
				})
			case TokenOrAssign:
				block.Push(&spirv.BitwiseOrInstruction{
					ResultType: resultValue.Type.ID(),
					ResultID:   resultValue.ID(),
					Operand1:   loadedValue.ID(),
					Operand2:   rhsValue.ID(),
				})
			case TokenXorAssign:
				block.Push(&spirv.BitwiseXorInstruction{
					ResultType: resultValue.Type.ID(),
					ResultID:   resultValue.ID(),
					Operand1:   loadedValue.ID(),
					Operand2:   rhsValue.ID(),
				})
			case TokenShlAssign:
				block.Push(&spirv.ShiftLeftLogicalInstruction{
					ResultType: resultValue.Type.ID(),
					ResultID:   resultValue.ID(),
					Base:       loadedValue.ID(),
					Shift:      rhsValue.ID(),
				})
			case TokenShrAssign:
				if tt, ok := t.(*spirv.IntType); ok {
					if tt.IsSigned {
						block.Push(&spirv.ShiftRightArithmeticInstruction{
							ResultType: resultValue.Type.ID(),
							ResultID:   resultValue.ID(),
							Base:       loadedValue.ID(),
							Shift:      rhsValue.ID(),
						})
					} else {
						block.Push(&spirv.ShiftRightLogicalInstruction{
							ResultType: resultValue.Type.ID(),
							ResultID:   resultValue.ID(),
							Base:       loadedValue.ID(),
							Shift:      rhsValue.ID(),
						})
					}
				} else {
					panic("unsupported type for shift right assignment")
				}
			}
			block.Push(&spirv.StoreInstruction{
				Pointer: obj.ID(),
				Object:  resultValue.ID(),
			})
		}
	default:
		panic("unsupported assignment operator")
	}
}

func (ir *IREmitter) emitIfStmt(ifStmt *IfStmt) {
	if ifStmt.Init != nil {
		ir.emitStatement(ifStmt.Init)
	}

	cond := ir.emitExpression(ifStmt.Cond)

	block := ir.currentBlock()
	fn := block.Function

	trueBlock := fn.NewBlock("true_block")
	falseBlock := fn.NewBlock("false_block")
	mergeBlock := fn.NewBlock("if_merge")
	block.Push(&spirv.SelectionMergeInstruction{
		MergeBlock: mergeBlock.ID(),
		Control:    spirv.SelectionControlNone,
	})

	block.Push(&spirv.BranchConditional{
		Condition:  cond.ID(),
		TrueLabel:  trueBlock.ID(),
		FalseLabel: falseBlock.ID(),
	})
	ir.leaveBlock()

	ir.enterBlock(trueBlock)
	ir.emitStatement(ifStmt.Body)
	ir.branchToMergeBlockIfNeeded(trueBlock, mergeBlock)
	ir.leaveBlock()

	ir.enterBlock(falseBlock)
	if ifStmt.Else != nil {
		ir.emitStatement(ifStmt.Else)
	}
	ir.branchToMergeBlockIfNeeded(falseBlock, mergeBlock)
	ir.leaveBlock()

	// continue emitting code in merge block
	ir.enterBlock(mergeBlock)
}
func (ir *IREmitter) branchToMergeBlockIfNeeded(currentBlock, mergeBlock *spirv.Block) {
	if !currentBlock.IsTerminated() {
		currentBlock.Push(&spirv.Branch{
			TargetLabel: mergeBlock.ID(),
		})
	}
}

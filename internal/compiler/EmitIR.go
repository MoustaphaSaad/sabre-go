package compiler

import (
	"go/constant"

	"github.com/MoustaphaSaad/sabre-go/internal/compiler/spirv"
)

type IREmitter struct {
	unit       *Unit
	module     *spirv.Module
	blockStack []*spirv.Block
}

func NewIREmitter(u *Unit) *IREmitter {
	return &IREmitter{
		unit: u,
		// we set the addressing and memory model to default values for now
		module:     spirv.NewModule(spirv.AddressingModelLogical, spirv.MemoryModelGLSL450),
		blockStack: make([]*spirv.Block, 0),
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
	default:
		panic("unsupported symbol")
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
	ir.enterBlock(spirvBlock)
	defer ir.leaveBlock()

	if len(funcDecl.Body.Stmts) == 0 {
		spirvBlock.Push(&spirv.ReturnInstruction{})
		return
	}

	for _, stmt := range funcDecl.Body.Stmts {
		ir.emitStatement(stmt, spirvBlock)
	}
}

func (ir *IREmitter) emitExpression(expr Expr) spirv.Object {
	switch e := expr.(type) {
	case *LiteralExpr:
		return ir.emitLiteralExpr(e)
	case *UnaryExpr:
		return ir.emitUnaryExpr(e)
	case *BinaryExpr:
		return ir.emitBinaryExpr(e)
	default:
		panic("unsupported expression")
	}
}

func (ir *IREmitter) emitLiteralExpr(e *LiteralExpr) spirv.Object {
	tav := ir.unit.semanticInfo.TypeOf(e)
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

	default:
		panic("unsupported binary operator")
	}
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

func (ir *IREmitter) emitStatement(stmt Stmt, block *spirv.Block) {
	switch s := stmt.(type) {
	case *ReturnStmt:
		ir.emitReturnStmt(s, block)
	default:
		panic("unsupported statement")
	}
}

func (ir *IREmitter) emitReturnStmt(s *ReturnStmt, block *spirv.Block) {
	if len(s.Exprs) > 0 {
		// TODO: Multiple return values
		block.Push(&spirv.ReturnValueInstruction{Value: ir.emitExpression(s.Exprs[0]).ID()})
	} else {
		block.Push(&spirv.ReturnInstruction{})
	}
}

package compiler

import (
	"slices"

	"github.com/MoustaphaSaad/sabre-go/internal/compiler/spirv"
)

func RewriteIR(mod *spirv.Module) {
	PassTerminateBlocks(mod)
	PassPullLocalVarsToFuncEntry(mod)
}

func PassPullLocalVarsToFuncEntry(mod *spirv.Module) {
	for _, obj := range mod.Objects {
		if fn, ok := obj.(*spirv.Function); ok {
			pullLocalVarsToFuncEntry(fn)
		}
	}
}
func pullLocalVarsToFuncEntry(fn *spirv.Function) {
	newEntry := make([]spirv.Instruction, 0)
	for _, bb := range fn.Blocks {
		bb.Instructions = slices.DeleteFunc(bb.Instructions, func(ins spirv.Instruction) bool {
			if ins.Opcode() == spirv.OpVariable {
				newEntry = append(newEntry, ins)
				return true
			}
			return false
		})
	}
	newEntry = append(newEntry, fn.Blocks[0].Instructions...)
	fn.Blocks[0].Instructions = newEntry
}

func PassTerminateBlocks(mod *spirv.Module) {
	for _, obj := range mod.Objects {
		if fn, ok := obj.(*spirv.Function); ok {
			terminateBlocksForVoidFuncs(fn)
		}
	}
}
func terminateBlocksForVoidFuncs(fn *spirv.Function) {
	_, isVoid := fn.Type.ReturnType.(*spirv.VoidType)

	for _, bb := range fn.Blocks {
		if !isBlockTerminated(bb) {
			if isVoid {
				bb.Push(&spirv.ReturnInstruction{})
			} else {
				bb.Push(&spirv.UnreachableInstruction{})
			}
		}
	}
}
func isBlockTerminated(bb *spirv.Block) bool {
	if len(bb.Instructions) == 0 {
		return false
	}

	switch bb.Instructions[len(bb.Instructions)-1].(type) {
	case *spirv.ReturnInstruction, *spirv.ReturnValueInstruction:
		return true
	default:
		return false
	}
}

package compiler

import (
	"slices"

	"github.com/MoustaphaSaad/sabre-go/internal/compiler/spirv"
)

func RewriteIR(mod *spirv.Module) {
	PassPullLocalVarsToFuncEntry(mod)
	PassRemoveUnreachableBlocks(mod)
	PassTerminateBlocks(mod)
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
		if !bb.IsTerminated() {
			if isVoid {
				bb.Push(&spirv.ReturnInstruction{})
			} else {
				bb.Push(&spirv.UnreachableInstruction{})
			}
		}
	}
}

func PassRemoveUnreachableBlocks(mod *spirv.Module) {
	var removedBBs []spirv.ID
	for _, obj := range mod.Objects {
		if fn, ok := obj.(*spirv.Function); ok {
			removedBBs = removeUnreachableBlocks(fn)
		}
	}

	mod.RemoveObjects(removedBBs)
}
func removeUnreachableBlocks(fn *spirv.Function) (res []spirv.ID) {
	// skip functions with only the entry block
	if len(fn.Blocks) <= 1 {
		return
	}

	cfg := spirv.BuildCFG(fn)
	reachable := cfg.ReachableBlocks()
	fn.Blocks = slices.DeleteFunc(fn.Blocks, func(bb *spirv.Block) bool {
		if reachable[bb] {
			return false
		}
		res = append(res, bb.ID())
		return true
	})
	return
}

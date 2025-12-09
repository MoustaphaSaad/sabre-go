package compiler

import (
	"slices"

	"github.com/MoustaphaSaad/sabre-go/internal/compiler/spirv"
)

func RewriteIR(mod *spirv.Module) {
	PassPullLocalVarsToFuncEntry(mod)
}

func PassPullLocalVarsToFuncEntry(mod *spirv.Module) {
	for _, obj := range mod.Objects {
		if fn, ok := obj.(*spirv.Function); ok {
			PullLocalVarsToFuncEntry(fn)
		}
	}
}
func PullLocalVarsToFuncEntry(fn *spirv.Function) {
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

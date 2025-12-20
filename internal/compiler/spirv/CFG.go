package spirv

import "fmt"

type CFG struct {
	Function  *Function
	succEdges map[*Block][]*Block
	predEdges map[*Block][]*Block
}

func BuildCFG(fn *Function) (cfg *CFG) {
	cfg = &CFG{
		Function:  fn,
		succEdges: make(map[*Block][]*Block),
		predEdges: make(map[*Block][]*Block),
	}

	for _, block := range fn.Blocks {
		cfg.addEdges(block, block.SuccessorIDs())
	}

	return cfg
}

func (cfg *CFG) addEdges(block *Block, outBlockIDs []ID) {
	seen := make(map[ID]struct{}, len(outBlockIDs))
	for _, outID := range outBlockIDs {
		if _, ok := seen[outID]; ok {
			continue
		}
		seen[outID] = struct{}{}

		outObj := cfg.Function.Module.GetObject(outID)
		if outObj == nil {
			panic(fmt.Sprintf("Missing block %v", outID))
		}
		outBlock := outObj.(*Block)
		cfg.succEdges[block] = append(cfg.succEdges[block], outBlock)
		cfg.predEdges[outBlock] = append(cfg.predEdges[outBlock], block)
	}
}

func (cfg *CFG) Successors(block *Block) []*Block {
	return cfg.succEdges[block]
}

func (cfg *CFG) Predecessors(block *Block) []*Block {
	return cfg.predEdges[block]
}

func (cfg *CFG) ReachableBlocks() map[*Block]bool {
	reachable := make(map[*Block]bool, len(cfg.Function.Blocks))
	if len(cfg.Function.Blocks) == 0 {
		return reachable
	}

	queue := []*Block{cfg.Function.Blocks[0]}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if reachable[current] {
			continue
		}
		reachable[current] = true

		for _, succ := range cfg.Successors(current) {
			if !reachable[succ] {
				queue = append(queue, succ)
			}
		}
	}

	return reachable
}

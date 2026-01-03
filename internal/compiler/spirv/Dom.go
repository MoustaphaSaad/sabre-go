package spirv

import (
	"slices"
	"sort"
)

type domTimestamp struct {
	discoveryTime, finishingTime int
}

func (a domTimestamp) encapsulates(b domTimestamp) bool {
	return a.discoveryTime <= b.discoveryTime && a.finishingTime >= b.finishingTime
}

type DomInfo struct {
	CFG              *CFG
	ReversePostOrder []*Block
	IDoms            map[*Block]*Block
	domTimestamps    map[*Block]domTimestamp
}

func BuildDomInfo(cfg *CFG) (dom *DomInfo) {
	dom = &DomInfo{
		CFG:              cfg,
		ReversePostOrder: make([]*Block, 0, len(cfg.Function.Blocks)),
		IDoms:            make(map[*Block]*Block, len(cfg.Function.Blocks)),
		domTimestamps:    make(map[*Block]domTimestamp, len(cfg.Function.Blocks)),
	}

	dom.buildReversePostOrder()
	dom.buildImmediateDominators()
	dom.buildDomTimestamps()
	return
}

func (dom *DomInfo) buildReversePostOrder() {
	visited := make(map[*Block]bool)

	type frame struct {
		block     *Block
		succIndex int
	}

	entry := dom.CFG.Function.Blocks[0]
	stack := []frame{
		{block: entry, succIndex: 0},
	}
	visited[entry] = true

	for len(stack) > 0 {
		currentStackIndex := len(stack) - 1
		currentFrame := stack[currentStackIndex]
		currentSuccs := dom.CFG.Successors(currentFrame.block)

		if currentFrame.succIndex < len(currentSuccs) {
			nextSucc := currentSuccs[currentFrame.succIndex]
			stack[currentStackIndex].succIndex++

			if !visited[nextSucc] {
				visited[nextSucc] = true
				stack = append(stack, frame{block: nextSucc, succIndex: 0})
			}
		} else {
			dom.ReversePostOrder = append(dom.ReversePostOrder, currentFrame.block)
			stack = stack[:currentStackIndex]
		}
	}

	slices.Reverse(dom.ReversePostOrder)
}

func (dom *DomInfo) buildImmediateDominators() {
	rpo := dom.ReversePostOrder

	rpoPos := make(map[*Block]int)
	for i, b := range rpo {
		rpoPos[b] = i
	}

	entry := dom.CFG.Function.Blocks[0]
	if entry != rpo[0] {
		panic("entry block should be the first element of reverse post order")
	}

	doms := make([]int, len(rpo))
	for i := range doms {
		doms[i] = -1
	}
	// entry block dominates itself
	doms[0] = 0

	changed := true
	for changed {
		changed = false
		for i := 1; i < len(rpo); i++ {
			block := rpo[i]
			newIdom := -1

			for _, p := range dom.CFG.Predecessors(block) {
				if pos, ok := rpoPos[p]; ok && doms[pos] != -1 {
					newIdom = pos
					break
				}
			}

			if newIdom != -1 {
				for _, p := range dom.CFG.Predecessors(block) {
					pPos, ok := rpoPos[p]
					if !ok || doms[pPos] == -1 {
						continue
					}
					if pPos != newIdom {
						newIdom = intersect(doms, pPos, newIdom)
					}
				}

				if doms[i] != newIdom {
					doms[i] = newIdom
					changed = true
				}
			}
		}
	}

	for i, domIdx := range doms {
		// exclude entry node because it should have no idom
		if i != 0 && domIdx != -1 {
			dom.IDoms[rpo[i]] = rpo[domIdx]
		}
	}
}

func intersect(doms []int, b1, b2 int) int {
	finger1, finger2 := b1, b2
	for finger1 != finger2 {
		for finger1 > finger2 {
			finger1 = doms[finger1]
		}
		for finger2 > finger1 {
			finger2 = doms[finger2]
		}
	}
	return finger1
}

func (dom *DomInfo) buildDomTimestamps() {
	adj := make(map[*Block][]*Block)
	for child, parent := range dom.IDoms {
		adj[parent] = append(adj[parent], child)
	}

	time := 0
	var walk func(*Block)
	walk = func(currentBlock *Block) {
		time++
		timestamp := domTimestamp{discoveryTime: time, finishingTime: 0}
		for _, child := range adj[currentBlock] {
			walk(child)
		}
		timestamp.finishingTime = time
		dom.domTimestamps[currentBlock] = timestamp
	}

	walk(dom.CFG.Function.Blocks[0])
}

func (dom *DomInfo) Dominates(parent, child *Block) bool {
	if parent == child {
		return true
	}
	return dom.domTimestamps[parent].encapsulates(dom.domTimestamps[child])
}

func (dom *DomInfo) SortBlocks() []*Block {
	sorted := make([]*Block, len(dom.ReversePostOrder))
	copy(sorted, dom.ReversePostOrder)

	sort.SliceStable(sorted, func(i, j int) bool {
		blockA, blockB := sorted[i], sorted[j]

		if dom.Dominates(blockA, blockB) {
			return true
		}
		if dom.Dominates(blockB, blockA) {
			return false
		}

		// keep relative reverse post order
		return false
	})
	return sorted
}

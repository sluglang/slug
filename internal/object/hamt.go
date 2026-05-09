package object

import "math/bits"

const (
	hamtBitsPerLevel = 5
	hamtMask         = (1 << hamtBitsPerLevel) - 1
	hamtMaxDepth     = 13 // covers 64-bit hash in 5-bit chunks
)

type hamtLeaf struct {
	key  MapKey
	pair MapPair
}

type hamtNode struct {
	bitmap   uint32
	children [32]*hamtNode
	leaves   []hamtLeaf
	isBucket bool
}

func mapKeyHash(k MapKey) uint64 {
	return (uint64(k.Type[0]) << 56) ^ k.Value
}

func hamtInsert(node *hamtNode, hash uint64, depth int, key MapKey, pair MapPair, replaced *bool) *hamtNode {
	if node == nil {
		return &hamtNode{
			isBucket: true,
			leaves:   []hamtLeaf{{key: key, pair: pair}},
		}
	}
	if node.isBucket {
		out := &hamtNode{
			isBucket: true,
			leaves:   append([]hamtLeaf(nil), node.leaves...),
		}
		for i := range out.leaves {
			if out.leaves[i].key == key {
				out.leaves[i].pair = pair
				*replaced = true
				return out
			}
		}
		out.leaves = append(out.leaves, hamtLeaf{key: key, pair: pair})
		if len(out.leaves) <= 4 || depth >= hamtMaxDepth {
			return out
		}
		// promote bucket to branch node
		branch := &hamtNode{}
		for i := range out.leaves {
			leaf := out.leaves[i]
			h := mapKeyHash(leaf.key)
			replacedInner := false
			branch = hamtInsert(branch, h, depth, leaf.key, leaf.pair, &replacedInner)
		}
		return branch
	}

	seg := int((hash >> (depth * hamtBitsPerLevel)) & hamtMask)
	bit := uint32(1 << seg)
	out := *node
	if node.bitmap&bit == 0 {
		out.bitmap |= bit
		out.children[seg] = &hamtNode{
			isBucket: true,
			leaves:   []hamtLeaf{{key: key, pair: pair}},
		}
		return &out
	}
	out.children[seg] = hamtInsert(node.children[seg], hash, depth+1, key, pair, replaced)
	return &out
}

func hamtGet(node *hamtNode, hash uint64, depth int, key MapKey) (MapPair, bool) {
	if node == nil {
		return MapPair{}, false
	}
	if node.isBucket {
		for i := range node.leaves {
			if node.leaves[i].key == key {
				return node.leaves[i].pair, true
			}
		}
		return MapPair{}, false
	}
	seg := int((hash >> (depth * hamtBitsPerLevel)) & hamtMask)
	bit := uint32(1 << seg)
	if node.bitmap&bit == 0 {
		return MapPair{}, false
	}
	return hamtGet(node.children[seg], hash, depth+1, key)
}

func hamtDelete(node *hamtNode, hash uint64, depth int, key MapKey, removed *bool) *hamtNode {
	if node == nil {
		return nil
	}
	if node.isBucket {
		idx := -1
		for i := range node.leaves {
			if node.leaves[i].key == key {
				idx = i
				break
			}
		}
		if idx < 0 {
			return node
		}
		*removed = true
		if len(node.leaves) == 1 {
			return nil
		}
		out := &hamtNode{
			isBucket: true,
			leaves:   make([]hamtLeaf, 0, len(node.leaves)-1),
		}
		out.leaves = append(out.leaves, node.leaves[:idx]...)
		out.leaves = append(out.leaves, node.leaves[idx+1:]...)
		return out
	}

	seg := int((hash >> (depth * hamtBitsPerLevel)) & hamtMask)
	bit := uint32(1 << seg)
	if node.bitmap&bit == 0 {
		return node
	}
	child := node.children[seg]
	nextChild := hamtDelete(child, hash, depth+1, key, removed)
	if !*removed {
		return node
	}

	out := *node
	if nextChild == nil {
		out.bitmap &^= bit
		out.children[seg] = nil
		if out.bitmap == 0 {
			return nil
		}
		if out.bitmap&(out.bitmap-1) == 0 {
			onlySeg := trailingSegment(out.bitmap)
			only := out.children[onlySeg]
			if only != nil && only.isBucket {
				return only
			}
		}
		return &out
	}
	out.children[seg] = nextChild
	return &out
}

func hamtForEach(node *hamtNode, fn func(MapKey, MapPair) bool) bool {
	if node == nil {
		return true
	}
	if node.isBucket {
		for i := range node.leaves {
			leaf := node.leaves[i]
			if !fn(leaf.key, leaf.pair) {
				return false
			}
		}
		return true
	}
	bm := node.bitmap
	for bm != 0 {
		seg := trailingSegment(bm)
		if !hamtForEach(node.children[seg], fn) {
			return false
		}
		bm &= bm - 1
	}
	return true
}

func trailingSegment(bitmap uint32) int {
	return bits.TrailingZeros32(bitmap)
}

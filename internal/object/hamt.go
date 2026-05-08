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
	children []*hamtNode
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
	idx := hamtIndex(node.bitmap, bit)
	out := &hamtNode{
		bitmap:   node.bitmap,
		children: append([]*hamtNode(nil), node.children...),
	}
	if node.bitmap&bit == 0 {
		out.bitmap |= bit
		child := &hamtNode{
			isBucket: true,
			leaves:   []hamtLeaf{{key: key, pair: pair}},
		}
		out.children = append(out.children, nil)
		copy(out.children[idx+1:], out.children[idx:])
		out.children[idx] = child
		return out
	}
	out.children[idx] = hamtInsert(node.children[idx], hash, depth+1, key, pair, replaced)
	return out
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
	idx := hamtIndex(node.bitmap, bit)
	return hamtGet(node.children[idx], hash, depth+1, key)
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
	for i := range node.children {
		if !hamtForEach(node.children[i], fn) {
			return false
		}
	}
	return true
}

func hamtIndex(bitmap uint32, bit uint32) int {
	return bits.OnesCount32(bitmap & (bit - 1))
}

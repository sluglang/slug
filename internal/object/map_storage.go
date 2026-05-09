package object

import "sort"

type mapStorage interface {
	put(k MapKey, v MapPair) mapStorage
	del(k MapKey, removed *bool) mapStorage
	get(k MapKey) (MapPair, bool)
	len() int
	forEach(fn func(MapKey, MapPair) bool)
}

type hamtMapStorage struct {
	root *hamtNode
	size int
}

func (s hamtMapStorage) put(k MapKey, v MapPair) mapStorage {
	var replaced bool
	root := hamtInsert(s.root, mapKeyHash(k), 0, k, v, &replaced)
	size := s.size
	if !replaced {
		size++
	}
	return hamtMapStorage{root: root, size: size}
}

func (s hamtMapStorage) get(k MapKey) (MapPair, bool) {
	return hamtGet(s.root, mapKeyHash(k), 0, k)
}

func (s hamtMapStorage) del(k MapKey, removed *bool) mapStorage {
	root := hamtDelete(s.root, mapKeyHash(k), 0, k, removed)
	size := s.size
	if *removed {
		size--
	}
	return hamtMapStorage{root: root, size: size}
}

func (s hamtMapStorage) len() int {
	return s.size
}

func (s hamtMapStorage) forEach(fn func(MapKey, MapPair) bool) {
	hamtForEach(s.root, fn)
}

// stableMapPairs returns entries sorted by type/value for deterministic output.
func stableMapPairs(s mapStorage) []MapPair {
	if s == nil {
		return nil
	}
	entries := make([]struct {
		k MapKey
		p MapPair
	}, 0, s.len())
	s.forEach(func(k MapKey, p MapPair) bool {
		entries = append(entries, struct {
			k MapKey
			p MapPair
		}{k: k, p: p})
		return true
	})
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].k.Type != entries[j].k.Type {
			return entries[i].k.Type < entries[j].k.Type
		}
		return entries[i].k.Value < entries[j].k.Value
	})
	out := make([]MapPair, len(entries))
	for i, entry := range entries {
		out[i] = entry.p
	}
	return out
}

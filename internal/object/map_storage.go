package object

import "sort"

type mapBackend int

const (
	mapBackendNative mapBackend = iota
	mapBackendHAMT
)

var defaultMapBackend = mapBackendNative

func SetDefaultMapBackend(name string) {
	switch name {
	case "hamt":
		defaultMapBackend = mapBackendHAMT
	default:
		defaultMapBackend = mapBackendNative
	}
}

type mapStorage interface {
	put(k MapKey, v MapPair) mapStorage
	get(k MapKey) (MapPair, bool)
	len() int
	forEach(fn func(MapKey, MapPair) bool)
}

type nativeMapStorage struct {
	pairs map[MapKey]MapPair
}

func (s nativeMapStorage) put(k MapKey, v MapPair) mapStorage {
	if s.pairs == nil {
		s.pairs = make(map[MapKey]MapPair)
	}
	s.pairs[k] = v
	return s
}

func (s nativeMapStorage) get(k MapKey) (MapPair, bool) {
	if s.pairs == nil {
		return MapPair{}, false
	}
	p, ok := s.pairs[k]
	return p, ok
}

func (s nativeMapStorage) len() int {
	return len(s.pairs)
}

func (s nativeMapStorage) forEach(fn func(MapKey, MapPair) bool) {
	for k, v := range s.pairs {
		if !fn(k, v) {
			return
		}
	}
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

package common

type IntSet map[int]struct{}

func (s IntSet) Contains(elem int) bool {
	_, ok := s[elem]
	return ok
}

func (s IntSet) Add(elem int) {
	s[elem] = struct{}{}
}

func (s IntSet) AddList(elemList []int) {
	for _, e := range elemList {
		s[e] = struct{}{}
	}
}

func (s IntSet) Remove(elem int) {
	delete(s, elem)
}

func (s IntSet) ForEach(callback func(int)) {
	for elem, _ := range s {
		callback(elem)
	}
}

func (s IntSet) ToList() []int {
	var l []int
	for elem, _ := range s {
		l = append(l, elem)
	}
	return l
}

func (s IntSet) Size() int {
	return len(s)
}

func (s IntSet) Copy() IntSet {
	if s == nil {
		return nil
	}

	cpy := IntSet{}
	for k, v := range s {
		cpy[k] = v
	}
	return cpy
}

type StringSet map[string]struct{}

func (s StringSet) Contains(elem string) bool {
	_, ok := s[elem]
	return ok
}

func (s StringSet) Add(elem string) {
	s[elem] = struct{}{}
}

func (s StringSet) Remove(elem string) {
	delete(s, elem)
}

func (s StringSet) ForEach(callback func(string) bool) {
	for elem, _ := range s {
		if !callback(elem) {
			return
		}
	}
}

type UInt32Set map[uint32]struct{}

func (s UInt32Set) Contains(elem uint32) bool {
	_, ok := s[elem]
	return ok
}

func (s UInt32Set) Add(elem uint32) {
	s[elem] = struct{}{}
}

func (s UInt32Set) Remove(elem uint32) {
	delete(s, elem)
}

func (s UInt32Set) ForEach(callback func(uint32) bool) {
	for elem, _ := range s {
		if !callback(elem) {
			return
		}
	}
}

func (s UInt32Set) AddList(elemList []uint32) {
	for _, e := range elemList {
		s[e] = struct{}{}
	}
}

func (s UInt32Set) ToList() []uint32 {
	var l []uint32
	for elem, _ := range s {
		l = append(l, elem)
	}
	return l
}

func (s UInt32Set) Size() int {
	return len(s)
}

type UInt64Set map[uint64]struct{}

func (s UInt64Set) Contains(elem uint64) bool {
	_, ok := s[elem]
	return ok
}

func (s UInt64Set) Add(elem uint64) {
	s[elem] = struct{}{}
}

func (s UInt64Set) Remove(elem uint64) {
	delete(s, elem)
}

func (s UInt64Set) ForEach(callback func(uint64) bool) {
	for elem, _ := range s {
		if !callback(elem) {
			return
		}
	}
}

func (s UInt64Set) Size() int {
	return len(s)
}

func (s UInt64Set) ToInterfaceList() []interface{} {
	var l []interface{}
	for elem, _ := range s {
		l = append(l, elem)
	}
	return l
}

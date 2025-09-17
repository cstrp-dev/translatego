package utils

type Set struct {
	data map[string]bool
}

func NewSet(items []string) *Set {
	set := &Set{
		data: make(map[string]bool),
	}

	for _, item := range items {
		set.Add(item)
	}

	return set
}

func (s *Set) Add(item string) {
	s.data[item] = true
}

func (s *Set) Contains(item string) bool {
	_, exist := s.data[item]

	return exist
}

func (s *Set) Remove(item string) {
	delete(s.data, item)
}

func (s *Set) Size() int {

	return len(s.data)
}

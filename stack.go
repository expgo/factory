package factory

type setStack[T comparable] struct {
	items []T
}

func (s *setStack[T]) contains(item T) bool {
	for _, v := range s.items {
		if v == item {
			return true
		}
	}
	return false
}

func (s *setStack[T]) Push(item T) bool {
	if !s.contains(item) {
		s.items = append(s.items, item)
		return true
	}

	return false
}

func (s *setStack[T]) Pop() (t T, b bool) {
	if len(s.items) == 0 {
		return
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, true
}

func (s *setStack[T]) Last() (t T, b bool) {
	if len(s.items) == 0 {
		return t, false
	}

	return s.items[len(s.items)-1], true
}

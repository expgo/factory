package factory

type setStack struct {
	items []any
}

func (s *setStack) contains(item any) bool {
	for _, v := range s.items {
		if v == item {
			return true
		}
	}
	return false
}

func (s *setStack) Push(item any) bool {
	if !s.contains(item) {
		s.items = append(s.items, item)
		return true
	}

	return false
}

func (s *setStack) Pop() (t any, b bool) {
	if len(s.items) == 0 {
		return
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, true
}

func (s *setStack) Last() (t any, b bool) {
	if len(s.items) == 0 {
		return t, false
	}

	return s.items[len(s.items)-1], true
}

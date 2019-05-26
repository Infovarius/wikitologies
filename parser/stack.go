package parser

type stack []*Section

func (s *stack) empty() bool {
	return len(*s) == 0
}

func (s *stack) push(sections Sections) {
	for i := len(sections) - 1; i >= 0; i-- {
		*s = append(*s, sections[i])
	}
}

func (s *stack) pop() *Section {
	l := len(*s)
	last := (*s)[l-1]
	*s = (*s)[:l-1]

	return last
}

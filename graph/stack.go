package graph

type pair struct {
	title string
	hyper string
}

type stack []pair

func (s *stack) empty() bool {
	return len(*s) == 0
}

func (s *stack) push(title string, hypers ...string) {
	for i := len(hypers) - 1; i >= 0; i-- {
		*s = append(*s, pair{title: title, hyper: hypers[i]})
	}
}

func (s *stack) pop() (string, string) {
	l := len(*s)
	last := (*s)[l-1]
	*s = (*s)[:l-1]

	return last.hyper, last.title
}

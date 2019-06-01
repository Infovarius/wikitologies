package graph

type pair struct {
	t, h, ru string
}

type stack []pair

func (s *stack) empty() bool {
	return len(*s) == 0
}

func (s *stack) push(t string, hs []string) {
	for i := len(hs) - 1; i >= 0; i-- {
		*s = append(*s, pair{t: t, h: hs[i]})
	}
}

func (s *stack) push2(t string, hs, rus []string) {
	for i := len(hs) - 1; i >= 0; i-- {
		*s = append(*s, pair{t: t, h: hs[i], ru: rus[i]})
	}
}

func (s *stack) pop() (string, string, string) {
	l := len(*s)
	last := (*s)[l-1]
	*s = (*s)[:l-1]

	return last.t, last.h, last.ru
}

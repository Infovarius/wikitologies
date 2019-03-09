package hyperonyms

import (
	"sync"

	"github.com/stillpiercer/wikitologies/parser"
	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

func Find(titles []string, lang string) map[string][]string {
	var mu sync.Mutex
	var wg sync.WaitGroup
	hs := make(map[string][]string)

	for i := range titles {
		wg.Add(1)
		title := titles[i]

		go func() {
			defer wg.Done()

			text, err := wikt.GetText(title)
			if err != nil {
				return
			}

			s1 := parser.ParseText(text).ByHeader(lang)
			if s1 == nil || s1.SubSections == nil {
				return
			}

			if s1.SubSections[0].Level == wikt.L2 {
				for _, s2 := range s1.SubSections {
					if s2.SubSections == nil {
						continue
					}

					semProps := s2.SubSections.ByHeader(wikt.SemProps)
					if semProps == nil {
						continue
					}

					for _, m := range parser.ParseMeanings(semProps) {
						if len(m.Hyperonyms) > 0 {
							mu.Lock()
							hs[s2.Header] = append(hs[s2.Header], m.Hyperonyms...)
							mu.Unlock()
						}
					}
				}
			} else {
				semProps := s1.SubSections.ByHeader(wikt.SemProps)
				if semProps == nil {
					return
				}

				for _, m := range parser.ParseMeanings(semProps) {
					if len(m.Hyperonyms) > 0 {
						mu.Lock()
						hs[title] = append(hs[title], m.Hyperonyms...)
						mu.Unlock()
					}
				}
			}
		}()
	}

	wg.Wait()

	return hs
}

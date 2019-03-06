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
			text, err := wikt.GetText(title)
			if err != nil {
				return
			}

			sections := parser.ParseText(text)
			section := sections.ByHeader(lang)
			if section == nil || section.SubSections == nil {
				return
			}

			if section.SubSections[0].Level == wikt.L2 {
				section = section.SubSections[0]
				if section.SubSections == nil {
					return
				}
			}

			semProps := section.SubSections.ByHeader(wikt.SemProps)
			if semProps == nil {
				return
			}

			for _, meaning := range parser.ParseMeanings(semProps) {
				if len(meaning.Hyperonyms) > 0 {
					mu.Lock()
					hs[title] = append(hs[title], meaning.Hyperonyms...)
					mu.Unlock()
				}
			}

			wg.Done()
		}()
	}

	wg.Wait()

	return hs
}

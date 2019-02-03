package hypers

import (
	"github.com/stillpiercer/wikitologies/parser"
	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

func FindAll(titles []string, lang string) [][]string {
	hypers := make([][]string, len(titles))
	for i, title := range titles {
		text, err := wikt.GetText(title)
		if err != nil {
			continue
		}

		sections := parser.ParseText(text)
		section := sections.ByHeader(lang)
		if section == nil || section.SubSections == nil {
			continue
		}

		if section.SubSections.Level == wikt.L2 {
			section = section.SubSections.Sections[0]
			if section.SubSections == nil {
				continue
			}
		}

		semProps := section.SubSections.Sections.ByHeader(wikt.SemProps)
		if semProps == nil {
			continue
		}

		for _, meaning := range parser.ParseMeanings(semProps) {
			hypers[i] = append(hypers[i], meaning.Hyperonyms...)
		}
	}

	return hypers
}

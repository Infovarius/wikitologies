package main

import (
	"fmt"

	"github.com/stillpiercer/wikitologies/hypers"
	"github.com/stillpiercer/wikitologies/parser"
	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

const title = "prison"

func main() {
	// Shorter version of example_parser.go
	// Don't do that, you should always check errors!
	text, _ := wikt.GetText(title)
	s := parser.ParseText(text)[0]
	lang := s.Header
	foreign := lang != wikt.Russian
	if s.SubSections.Level == wikt.L2 {
		s = s.SubSections.Sections[0]
	}

	semProps := s.SubSections.Sections.ByHeader(wikt.SemProps)
	meanings := parser.ParseMeanings(semProps)
	var html string
	if foreign {
		html, _ = wikt.GetSectionHTML(title, semProps.SubSections.Sections.ByHeader(wikt.Meanings).Number)
	} else {
		html, _ = wikt.GetSectionHTML(title, s.SubSections.Sections.ByHeader(wikt.Translations).Number)
	}
	parser.ParseTranslations(meanings, foreign, html)

	meaning := meanings[0]
	fmt.Println("hyperonyms:")
	fmt.Println(meaning.Hyperonyms)
	fmt.Println("+ by synonyms:")
	synonyms := meaning.Synonyms
	hs := hypers.FindAll(synonyms, lang)
	for i, s := range synonyms {
		if len(hs[i]) > 0 {
			fmt.Printf("%s: %s\n", s, hs[i])
		}
	}

	// Better use search by translations only with foreign words.
	// Russian words usually have lots of translations,
	// so it will take a long time to parse everything.
	if foreign {
		fmt.Println("+ by translations:")
		for _, t := range meaning.Translations {
			hs := hypers.FindAll(t.Values, t.Language)
			for i, v := range t.Values {
				if len(hs[i]) > 0 {
					fmt.Printf("%s (%s): %s\n", v, t.Language, hs[i])
				}
			}
		}
	}
}

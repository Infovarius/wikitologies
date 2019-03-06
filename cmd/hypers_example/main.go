package main

import (
	"fmt"

	"github.com/stillpiercer/wikitologies/hyperonyms"
	"github.com/stillpiercer/wikitologies/parser"
	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

const title = "prison"

func main() {
	// Shorter version of parser_example.go
	// Don't do that, you should always check errors!
	text, _ := wikt.GetText(title)
	section := parser.ParseText(text)[0]
	lang := section.Header
	foreign := lang != wikt.Russian

	if section.SubSections[0].Level == wikt.L2 {
		section = section.SubSections[0]
	}

	semProps := section.SubSections.ByHeader(wikt.SemProps)
	meanings := parser.ParseMeanings(semProps)

	var html string
	if foreign {
		html, _ = wikt.GetSectionHTML(title, semProps.SubSections.ByHeader(wikt.Meanings).Number)
	} else {
		html, _ = wikt.GetSectionHTML(title, section.SubSections.ByHeader(wikt.Translations).Number)
	}
	parser.ParseTranslations(meanings, foreign, html)

	fmt.Println("hyperonyms:")
	fmt.Println(meanings[0].Hyperonyms)

	fmt.Println("+ by synonyms:")
	hs := hyperonyms.Find(meanings[0].Synonyms, lang)
	for _, s := range meanings[0].Synonyms {
		if len(hs[s]) > 0 {
			fmt.Printf("%s: %s\n", s, hs[s])
		}
	}

	// Better use search by translations only with foreign words.
	// Russian words usually have lots of translations,
	// so it will take a long time to parse everything.
	if foreign {
		fmt.Println("+ by translation to russian:")
		t := meanings[0].Translations[0]
		hs := hyperonyms.Find(t.Values, t.Language)
		for _, v := range t.Values {
			if len(hs[v]) > 0 {
				fmt.Printf("%s: %s\n", v, hs[v])
			}
		}
	}
}

package main

import (
	"fmt"
	"log"

	"github.com/stillpiercer/wikitologies/parser"
	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

const title = "лосось"

func main() {
	// First of all get full page text.
	text, err := wikt.GetText(title)
	if err != nil {
		if err == wikt.ErrMissing {
			fmt.Printf(title + " ")
		}
		fmt.Println(err)
		return
	}

	// Then use parser to split text by sections.
	sections := parser.ParseText(text)
	if len(sections) == 0 {
		fmt.Println("No sections found")
		return
	}
	fmt.Println(sections)

	// Let's choose the first language section and check its subsections.
	s := sections[0]
	if s.SubSections == nil {
		fmt.Printf("%s section does not contain any subsections\n", s)
		return
	}
	fmt.Println(s.SubSections)

	// Save a flag storing if selected one is foreign or not, it'll be used to parse translations later.
	foreign := s.Header != wikt.Russian

	// If there is any level 2 sections, let's choose the first one, again.
	if s.SubSections[0].Level == wikt.L2 {
		// If you no longer need selected level 1 section (like now), you're free to reassign variable.
		s = s.SubSections[0]
		// Don't forget to check subsections!
		if s.SubSections == nil {
			fmt.Printf("%s section does not contain any subsections\n", s)
			return
		}
	}

	// Find a semantic properties section, make sure it's presented.
	semProps := s.SubSections.ByHeader(wikt.SemProps)
	if semProps == nil {
		fmt.Println("No semantic properties section found")
		return
	}

	// Ready to parse meanings!
	meanings := parser.ParseMeanings(semProps)
	if len(meanings) == 0 {
		fmt.Println("No meanings found")
		return
	}
	fmt.Println(meanings)

	// To parse translations we need to get html code of their section.
	var html string
	if foreign { // For foreign words translations are just parsed meanings (which are written in russian).
		html, err = wikt.GetSectionHTML(title, semProps.SubSections.ByHeader(wikt.Meanings).Number)
		if err != nil {
			log.Fatal(err)
		}
	} else { // For russian words there is special section (may be missed).
		translations := s.SubSections.ByHeader(wikt.Translations)
		if translations != nil {
			html, err = wikt.GetSectionHTML(title, translations.Number)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Add translations to already found meanings (optional).
	parser.ParseTranslations(meanings, foreign, html)

	// Let's print first meaning's translations.
	fmt.Print(meanings[0].Translations)
}

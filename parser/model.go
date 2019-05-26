package parser

import (
	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

type Section struct {
	Header      string
	Text        string
	Number      int
	Level       wikt.Level
	SubSections Sections
}

type Sections []*Section

func (ss Sections) ByHeader(header string) *Section {
	for _, s := range ss {
		if s.Header == header {
			return s
		}
	}

	return nil
}

func (ss Sections) Headers() []string {
	var headers []string
	for _, s := range ss {
		headers = append(headers, s.Header)
	}

	return headers
}

type Word []struct {
	Language string
	Meanings Meanings
}

func (w Word) ByLanguage(lang string) Meanings {
	for _, v := range w {
		if v.Language == lang {
			return v.Meanings
		}
	}

	return nil
}

type Meaning struct {
	Value        string
	Examples     []string
	Synonyms     []string
	Antonyms     []string
	Hyperonyms   []string
	Hyponyms     []string
	Translations Translations
}

type Meanings []*Meaning

type Translation struct {
	Language string
	Values   []string
}

type Translations []*Translation

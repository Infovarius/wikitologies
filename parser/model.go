package parser

import (
	"fmt"
	"strings"

	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

type Section struct {
	Header      string
	Text        string
	Number      int
	Level       wikt.Level
	SubSections Sections
}

func (s Section) String() string {
	return s.Header
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

type Meaning struct {
	Value        string
	Examples     []string
	Synonyms     []string
	Antonyms     []string
	Hyperonyms   []string
	Hyponyms     []string
	Translations Translations
}

func (m Meaning) String() string {
	return m.Value
}

type Meanings []*Meaning

type Translation struct {
	Language string
	Values   []string
}

func (t Translation) String() string {
	return fmt.Sprintf("%s: %s", t.Language, strings.Join(t.Values, ", "))
}

type Translations []*Translation

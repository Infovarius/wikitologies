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

func (ss Sections) String() string {
	var str strings.Builder
	for i, s := range ss {
		str.WriteString(fmt.Sprintf("%d. %s\n", i, s))
	}

	return str.String()
}

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
	parts := []string{m.Value}

	if len(m.Synonyms) > 0 {
		parts = append(parts, fmt.Sprintf("Synonyms: %s", strings.Join(m.Synonyms, ", ")))
	}
	if len(m.Antonyms) > 0 {
		parts = append(parts, fmt.Sprintf("Antonyms: %s", strings.Join(m.Antonyms, ", ")))
	}
	if len(m.Hyperonyms) > 0 {
		parts = append(parts, fmt.Sprintf("Hyperonyms: %s", strings.Join(m.Hyperonyms, ", ")))
	}
	if len(m.Hyponyms) > 0 {
		parts = append(parts, fmt.Sprintf("Hyponyms: %s", strings.Join(m.Hyponyms, ", ")))
	}

	return strings.Join(parts, "\n")
}

type Meanings []*Meaning

func (ms Meanings) String() string {
	var str strings.Builder
	for i, m := range ms {
		str.WriteString(fmt.Sprintf("%d. %s\n", i, m))
	}

	return str.String()
}

type Translation struct {
	Language string
	Values   []string
}

func (t Translation) String() string {
	return fmt.Sprintf("%s: %s", t.Language, strings.Join(t.Values, ", "))
}

type Translations []*Translation

func (ts Translations) String() string {
	var str strings.Builder
	for _, t := range ts {
		str.WriteString(fmt.Sprintln(t))
	}

	return str.String()
}

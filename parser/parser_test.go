package parser_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stillpiercer/wikitologies/parser"
)

func TestParseText(t *testing.T) {
	cases := []struct {
		title    string
		headers  []string // level 1 headers
		lastNumb int      // ...last header's number
		subs     []string // ...first header's subheaders
	}{
		{
			title:    "лосось",
			headers:  []string{"Русский"},
			lastNumb: 1,
			subs: []string{
				"Морфологические и синтаксические свойства",
				"Произношение",
				"Семантические свойства",
				"Родственные слова",
				"Этимология",
				"Фразеологизмы и устойчивые сочетания",
				"Перевод",
				"Библиография",
			},
		},
		{
			title: "test",
			headers: []string{
				"Английский",
				"Бретонский",
				"Венгерский",
				"Интерлингва",
				"Французский",
				"Фриульский",
			},
			lastNumb: 83,
			subs:     []string{"test I", "test II", "test III"},
		},
		{
			title:    "light",
			headers:  []string{"Английский"},
			lastNumb: 1,
			subs: []string{
				"light I (существительное)",
				"light I (прилагательное)",
				"light I (глагол)",
				"light II (прилагательное)",
				"light II (глагол)",
				"light II (наречие)",
			},
		},
	}

	for _, c := range cases {
		text, _ := ioutil.ReadFile("mock/" + c.title + "_full.txt")
		sections := parser.ParseText(string(text))

		assert.Equal(t, c.headers, sections.Headers(), "headers mismatch")

		last := sections[len(sections)-1]
		assert.Equal(t, c.lastNumb, last.Number, "last header's number mismatch")

		subs := sections[0].SubSections
		assert.Equal(t, c.subs, subs.Headers(), "subheaders mismatch")
	}
}

func TestParseMeanings(t *testing.T) {
	cases := []struct {
		title      string
		count      int      // meanings count
		synonyms   []string // first meaning's synonyms
		antonyms   []string // ...antonyms
		hyperonyms []string // ...hyperonyms
		hyponyms   []string // ...hyponyms
	}{
		{
			title:      "лосось",
			count:      2,
			synonyms:   nil,
			antonyms:   nil,
			hyperonyms: []string{"рыба", "животное"},
			hyponyms:   []string{"кета", "сёмга", "форель", "благородный лосось", "озёрный лосось"},
		},
		{
			title:      "test",
			count:      7,
			synonyms:   nil,
			antonyms:   nil,
			hyperonyms: nil,
			hyponyms:   nil,
		},
		{
			title:      "light",
			count:      4,
			synonyms:   []string{"electromagnetic radiation"},
			antonyms:   []string{"darkness"},
			hyperonyms: []string{"radiation"},
			hyponyms:   []string{"infrared light"},
		},
	}

	for _, c := range cases {
		text, _ := ioutil.ReadFile("mock/" + c.title + "_semantics.txt")
		semProps := parser.ParseText(string(text))[0]
		meanings := parser.ParseMeanings(semProps)

		assert.Equal(t, c.count, len(meanings), "meanings count mismatch")

		assert.Equal(t, c.synonyms, meanings[0].Synonyms, "synonyms mismatch")
		assert.Equal(t, c.antonyms, meanings[0].Antonyms, "antonyms mismatch")
		assert.Equal(t, c.hyperonyms, meanings[0].Hyperonyms, "hyperonyms mismatch")
		assert.Equal(t, c.hyponyms, meanings[0].Hyponyms, "hyponyms mismatch")
	}
}

func TestParseTranslations(t *testing.T) {
	cases := []struct {
		title        string
		foreign      bool
		translations parser.Translations // first meaning's last translation
	}{
		{
			title:   "лосось",
			foreign: false,
			translations: []*parser.Translation{
				{Language: "Японский", Values: []string{"サーモン"}},
			},
		},
		{
			title:   "test",
			foreign: true,
			translations: []*parser.Translation{
				{Language: "Русский", Values: []string{"проверка", "испытание", "тест"}},
			},
		},
		{
			title:   "light",
			foreign: true,
			translations: []*parser.Translation{
				{Language: "Русский", Values: []string{"свет"}},
			},
		},
	}

	for _, c := range cases {
		text, _ := ioutil.ReadFile("mock/" + c.title + "_semantics.txt")
		semProps := parser.ParseText(string(text))[0]
		meanings := parser.ParseMeanings(semProps)

		html, _ := ioutil.ReadFile("mock/" + c.title + "_translations.txt")
		parser.ParseTranslations(meanings, c.foreign, string(html))

		l := len(meanings[0].Translations)
		assert.Equal(t, c.translations, meanings[0].Translations[l-1:], "translations mismatch")
	}
}

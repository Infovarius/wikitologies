package wiktionary

import (
	"regexp"
)

type Level int

const (
	L1 Level = iota + 1
	L2
	L3
	L4
)

var HeadersRE = map[Level]*regexp.Regexp{
	L1: regexp.MustCompile("\n= (.*?) =(\n|$)"),
	L2: regexp.MustCompile("\n== (.*?) ==(\n|$)"),
	L3: regexp.MustCompile("\n=== (.*?) ===(\n|$)"),
	L4: regexp.MustCompile("\n==== (.*?) ====(\n|$)"),
}

const (
	T2Content      = "Type 2 synonyms, antonyms, etc."
	Brackets       = "Round brackets"
	Link           = "Wikitext link"
	Template       = "Wikitext any template"
	Semantics      = "Wikitext semantics template"
	Meaning        = "Wikitext meaning template"
	HTMLcomment    = "HTML comment"
	TranslationsRU = "Wikitext translations template"
)

var TemplatesRE = map[string]*regexp.Regexp{
	T2Content:      regexp.MustCompile("синонимы:|конверсивы:|антонимы:|гиперонимы:|гипонимы:|согипонимы:|холонимы:|меронимы:|управление:|время:|категории:|якорь:|язык"),
	Brackets:       regexp.MustCompile(`\([^(]*?\)`),
	Link:           regexp.MustCompile(`\[\[([^|[]*?)]]`),
	Template:       regexp.MustCompile("{{[^{]*?}}"),
	Semantics:      regexp.MustCompile(`{{семантика\|([^{]*?)}}`),
	Meaning:        regexp.MustCompile(`#(.*?)(\n|$)`),
	HTMLcomment:    regexp.MustCompile(`<!--.*?-->`),
	TranslationsRU: regexp.MustCompile(`{{перев-блок.*?\n`),
}

const (
	Russian = "Русский"

	SemProps     = "Семантические свойства"
	Meanings     = "Значение"
	Synonyms     = "Синонимы"
	Antonyms     = "Антонимы"
	Hyperonyms   = "Гиперонимы"
	Hyponyms     = "Гипонимы"
	Translations = "Перевод"

	ExampleSep     = "◆"
	Proto          = "Общее прототипическое значение"
	MissingExample = "Отсутствует пример употребления (см. рекомендации)."
)

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
	L1: regexp.MustCompile("\n= (.*) =\n"),
	L2: regexp.MustCompile("\n== (.*) ==\n"),
	L3: regexp.MustCompile("\n=== (.*) ===\n"),
	L4: regexp.MustCompile("\n==== (.*) ====\n"),
}

type Template string

const (
	LinkedWord   Template = "Linked word"
	LinkedWordRu Template = "Linked word (russian)"
	T2Content    Template = "Type 2 synonyms, antonyms, etc."
)

var TemplatesRE = map[Template]*regexp.Regexp{
	LinkedWord:   regexp.MustCompile(`(\p{L}+) \( /w`),
	LinkedWordRu: regexp.MustCompile(`([а-яё]+) \( /w`),
	T2Content:    regexp.MustCompile("синонимы:|конверсивы:|антонимы:|гиперонимы:|гипонимы:|согипонимы:|холонимы:|меронимы:|управление:|категории:"), // якорь? язык?
}

const (
	Russian = "Русский"

	SemProps     = "Семантические свойства"
	Translations = "Перевод"

	Meanings   = "Значение"
	Synonyms   = "Синонимы"
	Antonyms   = "Антонимы"
	Hyperonyms = "Гиперонимы"
	Hyponyms   = "Гипонимы"

	ExampleSep     = "◆"
	Proto          = "Общее прототипическое значение"
	MissingExample = "Отсутствует пример употребления (см. рекомендации)."
)

package wiktionary

import (
	"regexp"
)

type Level string

const (
	Level1 Level = "Level 1"
	Level2 Level = "Level 2"
	Level3 Level = "Level 3"
	Level4 Level = "Level 4"
)

var Levels = map[Level]*regexp.Regexp{
	Level1: regexp.MustCompile("\n= (.*) =\n"),
	Level2: regexp.MustCompile("\n== (.*) ==\n"),
	Level3: regexp.MustCompile("\n=== (.*) ===\n"),
	Level4: regexp.MustCompile("\n==== (.*) ====\n"),
}

type Template string

const (
	Proto       Template = "Proto meaning"
	T2Examples  Template = "Type 2 examples"
	T2Content   Template = "Type 2 synonyms, antonyms, etc."
	TranslLinks Template = "Translations links"
)

var Templates = map[Template]*regexp.Regexp{
	Proto:       regexp.MustCompile("Общее прототипическое значение — .*\n"),
	T2Examples:  regexp.MustCompile("◆ (.*) {4}"),
	T2Content:   regexp.MustCompile("синонимы:|конверсивы:|антонимы:|гиперонимы:|гипонимы:|согипонимы:|холонимы:|меронимы:|управление:|категории:"), // якорь? язык?
	TranslLinks: regexp.MustCompile("\\([^()]*\\)"),
}

const (
	// Level 3
	SemProps     = "Семантические свойства"
	Translations = "Перевод"
	// Level 4
	Meanings   = "Значение"
	Synonyms   = "Синонимы"
	Antonyms   = "Антонимы"
	Hyperonyms = "Гиперонимы"
	Hyponyms   = "Гипонимы"
)

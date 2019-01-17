package parser

import (
	"github.com/stillpiercer/wikitologies/utils"
	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

type Section struct {
	Number  int
	Content interface{}
}

type SubSections struct {
	Level    wikt.Level
	Sections *utils.OrderedMap
}

type Content struct {
	Examples   []string
	Synonyms   []string
	Antonyms   []string
	Hyperonyms []string
	Hyponyms   []string
	// TODO: Холонимы
	// TODO: Меронимы
	Translations *utils.OrderedMap
}

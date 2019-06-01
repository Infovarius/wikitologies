package graph

import (
	"errors"
	"fmt"
	"log"
	"strings"

	dot "github.com/awalterschulze/gographviz"

	"github.com/stillpiercer/wikitologies/parser"
	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

var presets = map[string]int{
	"реальность": 1,
	"организм":   1,
	"мир":        3,
}

func Build(title, lang string, strict bool, params map[string]int) (*dot.Graph, error) {
	word, err := parser.Parse(title)
	if err != nil {
		return nil, err
	}

	var meanings parser.Meanings
	if lang == "" {
		lang = word[0].Language
		meanings = word[0].Meanings
	} else {
		if meanings = word.ByLanguage(lang); meanings == nil {
			return nil, errors.New(fmt.Sprintf("ошибка: %s язык для слова %s не найден", lang, title))
		}
	}

	idx := index(title, lang, meanings, params)
	l := len(meanings)
	if idx >= l {
		return nil, errors.New(
			fmt.Sprintf("ошибка: некорректные предустановки/параметры запроса для слова %s: запрошено значение %d (всего доступно %d)", title, idx, l))
	}

	g := dot.NewGraph()
	g.Directed = true
	g.Name = glue(fmt.Sprintf("%s (%s)", title, lang))
	attrs := map[string]string{"tooltip": glue(meanings[idx].Value)}
	if l > 1 {
		attrs["color"] = "red"
	}
	_ = g.AddNode(g.Name, glue(title), attrs)

	stack := stack{}
	stack.push(title, meanings[idx].Hyperonyms)
	if lang != wikt.Russian {
		hs, rus, err := hypersRU(title, lang, meanings[idx], strict)
		if err != nil {
			return nil, err
		}
		stack.push2(title, hs, rus)
	}

	for !stack.empty() {
		t, h, ru := stack.pop()
		if _, ok := g.Nodes.Lookup[glue(h)]; ok && !strict {
			if _, ok := g.Edges.SrcToDsts[glue(t)][glue(h)]; !ok {
				_ = g.AddEdge(glue(t), glue(h), true, nil)
			}
			continue
		}

		word, err := parser.Parse(h)
		if err != nil {
			if err == wikt.ErrMissing {
				log.Println(h, err)
				continue
			}
			return nil, err
		}

		idx := -1
		meanings := word.ByLanguage(lang)
		if strict {
		loop:
			for i, m := range meanings {
				if ru != "" {
					if contains(m.Translations.ByLanguage(wikt.Russian), ru) {
						idx = i
						break loop
					}
				} else {
					if contains(m.Hyponyms, t) {
						idx = i
						break loop
					}
				}
			}
			if idx == -1 {
				continue
			}
			if node, ok := g.Nodes.Lookup[glue(h)]; ok && node.Attrs["tooltip"] == glue(meanings[idx].Value) {
				_ = g.AddEdge(glue(t), glue(h), true, nil)
				continue
			}
		} else {
			idx = index(h, lang, meanings, params)
			if idx >= len(meanings) {
				return nil, errors.New(fmt.Sprintf("ошибка: некорректные предустановки/параметры запроса для слова %s: запрошено значение %d (всего доступно %d)", h, idx, l))
			}
		}

		attrs := map[string]string{"tooltip": glue(meanings[idx].Value)}
		if !strict && len(meanings) > 1 {
			attrs["color"] = "red"
		}
		if ru != "" {
			attrs["xlabel"] = glue("(" + ru + ")")
			hs, rus, err := hypersRU(title, lang, meanings[idx], strict)
			if err != nil {
				return nil, err
			}
			stack.push2(title, hs, rus)
		}
		stack.push(h, meanings[idx].Hyperonyms)
		_ = g.AddNode(g.Name, glue(h), attrs)
		_ = g.AddEdge(glue(t), glue(h), true, nil)
	}

	return g, nil
}

func index(title, lang string, meanings parser.Meanings, params map[string]int) int {
	if i, ok := params[title]; ok {
		return i
	}

	if i, ok := presets[title]; ok && lang == wikt.Russian {
		return i
	}

	if len(meanings) > 1 && strings.HasPrefix(meanings[0].Value, "действие по значению гл.") {
		return 1
	}

	return 0
}

func hypersRU(title, lang string, meaning *parser.Meaning, strict bool) ([]string, []string, error) {
	var hs, rus []string
	for _, v := range meaning.Translations.ByLanguage(wikt.Russian) {
		w, err := parser.Parse(v)
		if err != nil {
			if err == wikt.ErrMissing {
				log.Println(v, err)
				continue
			}
			return nil, nil, err
		}

		for _, mru := range w.ByLanguage(wikt.Russian) {
			if contains(mru.Translations.ByLanguage(lang), title) {
				for _, hru := range mru.Hyperonyms {
					wh, err := parser.Parse(hru)
					if err != nil {
						if err == wikt.ErrMissing {
							log.Println(hru, err)
							continue
						}
						return nil, nil, err
					}

					if strict {
						for _, mhru := range wh.ByLanguage(wikt.Russian) {
							if contains(mhru.Hyponyms, v) {
								for _, v := range mhru.Translations.ByLanguage(lang) {
									if !contains(hs, v) {
										hs = append(hs, v)
										rus = append(rus, hru)
									}
								}
							}
						}
					} else {
						mhrus := wh.ByLanguage(wikt.Russian)
						if len(mhrus) > 0 {
							for _, v := range mhrus[0].Translations.ByLanguage(lang) {
								if !contains(hs, v) {
									hs = append(hs, v)
									rus = append(rus, hru)
								}
							}
						}
					}
				}
			}
		}
	}

	return hs, rus, nil
}

func contains(strings []string, value string) bool {
	for _, s := range strings {
		if value == s {
			return true
		}
	}

	return false
}

func glue(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

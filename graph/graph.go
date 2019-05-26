package graph

import (
	"errors"
	"fmt"
	"log"

	dot "github.com/awalterschulze/gographviz"

	"github.com/stillpiercer/wikitologies/parser"
	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

func Build(title, lang string, strict bool, params map[string]int) (*dot.Graph, error) {
	text, err := wikt.GetText(title)
	if err != nil {
		return nil, err
	}

	sections := parser.ParseText(text)
	word := parser.NewWord(title, sections)
	var meanings parser.Meanings
	if lang == "" {
		lang = word[0].Language
		meanings = word[0].Meanings
	} else {
		if meanings = word.ByLanguage(lang); meanings == nil {
			return nil, errors.New(fmt.Sprintf("ошибка: %s язык для слова %s не найден", lang, title))
		}
	}

	i := number(title, lang, params)
	l := len(meanings)
	if i >= l {
		return nil, errors.New(
			fmt.Sprintf("ошибка: некорректные предустановки/параметры запроса для слова %s: запрошено значение %d (всего доступно %d)", title, i, l))
	}

	g := dot.NewGraph()
	g.Directed = true
	g.Name = glue(fmt.Sprintf("%s (%s)", title, lang))
	_ = g.AddNode(g.Name, glue(title), map[string]string{"tooltip": glue(meanings[i].Value)})

	stack := stack{}
	stack.push(title, meanings[i].Hyperonyms...)
	for !stack.empty() {
		h, t := stack.pop()
		if _, ok := g.Nodes.Lookup[glue(h)]; ok && !strict {
			if _, ok := g.Edges.SrcToDsts[glue(t)][glue(h)]; !ok {
				_ = g.AddEdge(glue(t), glue(h), true, nil)
			}
			continue
		}

		text, err := wikt.GetText(h)
		if err != nil {
			if err == wikt.ErrMissing {
				log.Printf("%s: %s", h, err)
				continue
			}
			return nil, err
		}

		sections := parser.ParseText(text)
		w := parser.NewWord(h, sections)
		meanings := w.ByLanguage(lang)

		if strict {
		loop:
			for i, m := range meanings {
				for _, hyponym := range m.Hyponyms {
					if t == hyponym {
						if node, ok := g.Nodes.Lookup[glue(h)]; ok {
							if node.Attrs["tooltip"] == glue(meanings[i].Value) {
								_ = g.AddEdge(glue(t), glue(h), true, nil)
							}
						} else {
							_ = g.AddNode(g.Name, glue(h), map[string]string{"tooltip": glue(meanings[i].Value)})
							_ = g.AddEdge(glue(t), glue(h), true, nil)
							stack.push(h, meanings[i].Hyperonyms...)
						}
						break loop
					}
				}
			}
		} else {
			i := number(h, lang, params)
			l := len(meanings)
			if i >= l {
				return nil, errors.New(fmt.Sprintf("ошибка: некорректные предустановки/параметры запроса для слова %s: запрошено значение %d (всего доступно %d)", h, i, l))
			}

			_ = g.AddNode(g.Name, glue(h), map[string]string{"tooltip": glue(meanings[i].Value)})
			_ = g.AddEdge(glue(t), glue(h), true, nil)
			stack.push(h, meanings[i].Hyperonyms...)
		}
	}

	return g, nil
}

var presets = map[string]int{
	"реальность":     1,
	"создание":       2,
	"организм":       3,
	"мир":            3,
	"приспособление": 1,
	"организация":    2,
	"объединение":    1,
	"учреждение":     1,
	"сооружение":     1,
	"питьё":          1,
	"изделие":        1,
}

func number(title, lang string, params map[string]int) int {
	if i, ok := params[title]; ok {
		return i
	}

	if i, ok := presets[title]; ok && lang == wikt.Russian {
		return i
	}

	return 0
}

func glue(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

package graph

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	dot "github.com/awalterschulze/gographviz"
	"github.com/gomodule/redigo/redis"

	"github.com/stillpiercer/wikitologies/parser"
	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

type kind string

const (
	own       kind = "own"
	predicted kind = "predicted"
)

var presets = map[string]int{
	"реальность": 1,
	"организм":   1,
	"мир":        3,
}

func Build(title, lang string, strict bool, params map[string]int, pool *redis.Pool) (*dot.Graph, error) {
	log.Printf("=== building %s ===", title)
	word, err := GetWord(title, pool)
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
	attrs := map[string]string{
		"tooltip": glue(meanings[idx].Value),
	}
	name := title
	if l > 1 {
		attrs["color"] = "green"
		name += fmt.Sprintf(":%d", idx)
	}
	_ = g.AddNode(g.Name, glue(name), attrs)

	stack := stack{}
	if len(meanings[idx].Hyperonyms) > 0 {
		stack.push(name, meanings[idx].Hyperonyms)
		log.Printf("%s own: %s", name, meanings[idx].Hyperonyms)
	}
	if lang != wikt.Russian {
		hs, rus, err := predict(title, lang, meanings[idx], strict, meanings[idx].Hyperonyms, pool)
		if err != nil {
			return nil, err
		}
		if len(hs) > 0 {
			stack.push2(name, hs, rus)
			log.Printf("%s predicted: %s %s", name, hs, rus)
		}
	}

	for !stack.empty() {
		t, h, ru := stack.pop()
		var kind kind
		if ru == "" {
			kind = own
		} else {
			kind = predicted
		}
		log.Printf("%s -> %s [%s]: checking...", t, h, kind)

		word, err := GetWord(h, pool)
		if err != nil {
			if err == wikt.ErrMissing {
				log.Println(h, err)
				continue
			}
			return nil, err
		}

		meanings := word.ByLanguage(lang)
		l := len(meanings)
		if l == 0 {
			continue
		}

		idx := -1
		switch kind {
		case own:
			if strict {
				for i, m := range meanings {
					if contains(m.Hyponyms, t) {
						idx = i
						break
					}
				}
				if idx == -1 {
					log.Printf("%s -> %s [%s]: denied", t, h, kind)
					continue
				}
			} else {
				idx = index(h, lang, meanings, params)
				if idx >= l {
					return nil, errors.New(fmt.Sprintf("ошибка: некорректные предустановки/параметры запроса для слова %s: запрошено значение %d (всего доступно %d)", h, idx, l))
				}
			}
		case predicted:
			for i, m := range meanings {
				if contains(m.Translations.ByLanguage(wikt.Russian), ru) {
					idx = i
					break
				}
			}
			if idx == -1 {
				log.Printf("%s -> %s [%s]: denied", t, h, kind)
				continue
			}
		}

		name := h
		if l > 1 {
			name += fmt.Sprintf(":%d", idx)
			log.Printf("%s -> %s [%s]: %d/%d selected", t, h, kind, idx, l)
		}

		if node, ok := g.Nodes.Lookup[glue(name)]; ok {
			log.Printf("%s -> %s [%s]: %s node exists", t, h, kind, name)
			if node.Attrs["color"] != "green" && kind == own && !strict && l > 1 {
				node.Attrs["color"] = "green"
				log.Printf("%s color changed to green", name)
			}
			if _, ok := g.Edges.SrcToDsts[glue(t)][glue(name)]; !ok && glue(t) != glue(name) {
				_ = g.AddEdge(glue(t), glue(name), true, nil)
				log.Printf("%s -> %s [%s]: edge added", t, h, kind)
			}
			continue
		}

		attrs := map[string]string{"tooltip": glue(meanings[idx].Value)}
		if kind == predicted {
			attrs["color"] = "blue"
		}
		if l > 1 && kind == own && !strict {
			attrs["color"] = "green"
		}
		_ = g.AddNode(g.Name, glue(name), attrs)
		_ = g.AddEdge(glue(t), glue(name), true, nil)
		log.Printf("%s -> %s [%s]: added", t, h, kind)

		if len(meanings[idx].Hyperonyms) > 0 {
			stack.push(name, meanings[idx].Hyperonyms)
			log.Printf("%s own: %s", name, meanings[idx].Hyperonyms)
		}
		if lang != wikt.Russian {
			hs, rus, err := predict(h, lang, meanings[idx], strict, meanings[idx].Hyperonyms, pool)
			if err != nil {
				return nil, err
			}
			if len(hs) > 0 {
				stack.push2(name, hs, rus)
				log.Printf("%s predicted: %s %s", name, hs, rus)
			}
		}
	}

	log.Printf("=== done %s ===", title)
	return g, nil
}

func GetWord(title string, pool *redis.Pool) (parser.Word, error) {
	const datePrefix = "date:"
	const wordPrefix = "word:"

	c := pool.Get()
	defer c.Close()

	update := func(date string) (parser.Word, error) {
		word, err := parser.Parse(title)
		if err != nil {
			return nil, err
		}

		data, err := json.Marshal(word)
		if err != nil {
			return nil, err
		}

		_, err = c.Do("SET", datePrefix+title, date)
		if err != nil {
			return nil, err
		}

		_, err = c.Do("SET", wordPrefix+title, data)
		if err != nil {
			return nil, err
		}

		return word, nil
	}

	dateWikiStr, err := wikt.GetLastRevision(title)
	if err != nil {
		return nil, err
	}

	dateRedisStr, err := redis.String(c.Do("GET", datePrefix+title))
	if err != nil {
		if err == redis.ErrNil {
			return update(dateWikiStr)
		}
	}

	dateWiki, err := time.Parse(time.RFC3339, dateWikiStr)
	if err != nil {
		return nil, err
	}

	dateRedis, err := time.Parse(time.RFC3339, dateRedisStr)
	if err != nil {
		return nil, err
	}

	if dateWiki.Sub(dateRedis) > 0 {
		return update(dateWikiStr)
	}

	s, err := redis.String(c.Do("GET", wordPrefix+title))
	if err != nil {
		if err == redis.ErrNil {
			return update(dateWikiStr)
		}
		return nil, err
	}

	word := parser.Word{}
	err = json.Unmarshal([]byte(s), &word)

	return word, nil
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

func predict(title, lang string, meaning *parser.Meaning, strict bool, existing []string, pool *redis.Pool) ([]string, []string, error) {
	var hs, rus []string
	for _, tru := range meaning.Translations.ByLanguage(wikt.Russian) {
		w, err := GetWord(tru, pool)
		if err != nil {
			if err == wikt.ErrMissing {
				log.Println(tru, err)
				continue
			}
			return nil, nil, err
		}

		idx := -1
		for i, mru := range w.ByLanguage(wikt.Russian) {
			if contains(mru.Translations.ByLanguage(lang), title) {
				idx = i
				break
			}
		}
		if idx == -1 {
			continue
		}

		for _, hru := range w.ByLanguage(wikt.Russian)[idx].Hyperonyms {
			wh, err := GetWord(hru, pool)
			if err != nil {
				if err == wikt.ErrMissing {
					log.Println(hru, err)
					continue
				}
				return nil, nil, err
			}

			idx2 := -1
			if strict {
				for i, mhru := range wh.ByLanguage(wikt.Russian) {
					if contains(mhru.Hyponyms, tru) {
						idx2 = i
						break
					}
				}
				if idx2 == -1 {
					continue
				}
			} else {
				idx2 = index(hru, wikt.Russian, wh.ByLanguage(wikt.Russian), nil)
			}

			for _, t := range wh.ByLanguage(wikt.Russian)[idx2].Translations.ByLanguage(lang) {
				if !contains(existing, t) && !contains(hs, t) {
					hs = append(hs, t)
					rus = append(rus, hru)
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

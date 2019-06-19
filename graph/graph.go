package graph

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
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

type predictedParams struct {
	ru         string
	tooltip    string
	polysemous bool
}

var global = map[string]int{
	"реальность": 1,
	"организм":   1,
	"мир":        3,
}

func Build(titles []string, lang string, strict bool, presets map[string]int, pool *redis.Pool) (*dot.Graph, error) {
	log.Printf("=== building %s ===", titles)
	g := dot.NewGraph()
	g.Directed = true
	g.Name = glue(fmt.Sprintf("%s (%s)", titles, lang))

	stack := stack{}
	for _, title := range titles {
		word, err := GetWord(title, pool)
		if err != nil {
			if err == wikt.ErrMissing {
				log.Println(title, err)
				continue
			}
			return nil, err
		}

		var meanings parser.Meanings
		if meanings = word.ByLanguage(lang); meanings == nil {
			log.Printf("[WARNING] %s язык для слова %s не найден", lang, title)
			continue
		}

		idx := index(title, lang, meanings, presets)
		l := len(meanings)
		if idx >= l {
			return nil, fmt.Errorf("[ERROR] некорректные параметры запроса для слова %s: запрошено значение %d (всего доступно %d)", title, idx, l)
		}

		attrs := map[string]string{
			"tooltip":  glue(meanings[idx].Value),
			"penwidth": "3",
		}
		name := title
		if l > 1 {
			attrs["color"] = "green"
			name += fmt.Sprintf(":%d", idx)
		}
		_ = g.AddNode(g.Name, glue(name), attrs)

		if len(meanings[idx].Hyperonyms) > 0 {
			stack.push(name, meanings[idx].Hyperonyms)
			log.Printf("%s own: %s", name, meanings[idx].Hyperonyms)
		}
		if lang != wikt.Russian {
			hs, rus, err := predict(title, lang, meanings[idx], strict, presets, meanings[idx].Hyperonyms, pool)
			if err != nil {
				return nil, err
			}
			if len(hs) > 0 {
				stack.push2(name, hs, rus)
				log.Printf("%s predicted: %s %s", name, hs, rus)
			}
		}
	}

	for !stack.empty() {
		t, h, pp := stack.pop()
		var kind kind
		if pp == nil {
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
		var polysemous bool
		tooltip := fmt.Sprintf("%s->%s", t, h)
		switch kind {
		case own:
			if strict {
				for i, m := range meanings {
					if contains(m.Hyponyms, strings.Split(t, ":")[0]) {
						idx = i
						break
					}
				}
			} else {
				idx = index(tooltip, lang, meanings, presets)
				if idx >= l {
					return nil, fmt.Errorf("[ERROR] некорректные параметры запроса для слова %s: запрошено значение %d (всего доступно %d)", h, idx, l)
				}
			}
			polysemous = l > 1
			if polysemous {
				tooltip += ":" + strconv.Itoa(idx)
			}
		case predicted:
			for i, m := range meanings {
				if contains(m.Translations.ByLanguage(wikt.Russian), pp.ru) {
					idx = i
					break
				}
			}
			polysemous = pp.polysemous
			tooltip = pp.tooltip
		}
		if idx == -1 {
			log.Printf("%s -> %s [%s]: denied", t, h, kind)
			continue
		}

		name := h
		if l > 1 {
			name += fmt.Sprintf(":%d", idx)
			log.Printf("%s -> %s [%s]: %d/%d selected", t, h, kind, idx, l)
		}

		if _, ok := g.Nodes.Lookup[glue(name)]; ok {
			log.Printf("%s node exists", name)
			if _, ok := g.Edges.SrcToDsts[glue(t)][glue(name)]; !ok && t != name {
				addEdge(g, t, name, tooltip, kind, strict, polysemous)
				log.Printf("%s -> %s [%s]: edge added", t, name, kind)
			}
			continue
		}

		_ = g.AddNode(g.Name, glue(name), map[string]string{
			"tooltip":  glue(meanings[idx].Value),
			"penwidth": "3",
		})
		log.Printf("%s node added", name)
		addEdge(g, t, name, tooltip, kind, strict, polysemous)
		log.Printf("%s -> %s [%s]: edge added", t, name, kind)

		if len(meanings[idx].Hyperonyms) > 0 {
			stack.push(name, meanings[idx].Hyperonyms)
			log.Printf("%s own: %s", name, meanings[idx].Hyperonyms)
		}
		if lang != wikt.Russian {
			hs, pp, err := predict(h, lang, meanings[idx], strict, presets, meanings[idx].Hyperonyms, pool)
			if err != nil {
				return nil, err
			}
			if len(hs) > 0 {
				stack.push2(name, hs, pp)
				log.Printf("%s predicted: %s", name, hs)
			}
		}
	}

	log.Printf("=== done %s ===", titles)
	return g, nil
}

func index(title, lang string, meanings parser.Meanings, presets map[string]int) int {
	if i, ok := presets[title]; ok {
		return i
	}

	if strings.Contains(title, "->") {
		title = strings.Split(title, "->")[1]
	}

	if i, ok := global[title]; ok && lang == wikt.Russian {
		return i
	}

	if len(meanings) > 1 && strings.HasPrefix(meanings[0].Value, "действие по значению гл.") {
		return 1
	}

	return 0
}

func addEdge(g *dot.Graph, from, to, tooltip string, kind kind, strict, polysemous bool) {
	attrs := map[string]string{
		"penwidth": "3",
		"tooltip":  glue(tooltip),
	}
	if !strict && polysemous {
		switch kind {
		case own:
			attrs["color"] = "green"
		case predicted:
			attrs["color"] = "blue"
		}
	}
	_ = g.AddEdge(glue(from), glue(to), true, attrs)
}

func predict(title, lang string, meaning *parser.Meaning, strict bool, presets map[string]int, existing []string, pool *redis.Pool) ([]string, []*predictedParams, error) {
	var hs []string
	var params []*predictedParams
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
			tooltip := fmt.Sprintf("%s:%d->%s", tru, idx, hru)
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
				idx2 = index(tooltip, wikt.Russian, wh.ByLanguage(wikt.Russian), presets)
			}

			for _, t := range wh.ByLanguage(wikt.Russian)[idx2].Translations.ByLanguage(lang) {
				if !contains(existing, t) && !contains(hs, t) {
					hs = append(hs, t)
					params = append(params, &predictedParams{
						ru:         hru,
						tooltip:    tooltip + ":" + strconv.Itoa(idx2),
						polysemous: len(wh.ByLanguage(wikt.Russian)) > 1,
					})
				}
			}
		}
	}

	return hs, params, nil
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

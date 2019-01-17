package parser

import (
	"bufio"
	"log"
	"strings"

	"github.com/jaytaylor/html2text"
	"github.com/stillpiercer/wikitologies/utils"
	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

func ParseText(text string) *utils.OrderedMap {
	number := 1
	l1sections := parseSections(text, wikt.Level1)
	for _, l1name := range l1sections.Keys() {
		l1section := l1sections.ByKey(l1name).(*Section)
		l2sections := parseSections(l1section.Content.(string), wikt.Level2)
		if len(l2sections.Keys()) > 0 {
			l1section.Content = &SubSections{Level: wikt.Level2, Sections: l2sections}
			l1section.Number = number
			number++

			for _, l2name := range l2sections.Keys() {
				l2section := l2sections.ByKey(l2name).(*Section)
				l3sections := parseSections(l2section.Content.(string), wikt.Level3)
				if len(l3sections.Keys()) > 0 {
					l2section.Content = &SubSections{Level: wikt.Level3, Sections: l3sections}
					l2section.Number = number
					number++

					for _, l3name := range l3sections.Keys() {
						l3section := l3sections.ByKey(l3name).(*Section)
						l4sections := parseSections(l3section.Content.(string), wikt.Level4)
						if len(l4sections.Keys()) > 0 {
							l3section.Content = &SubSections{Level: wikt.Level4, Sections: l4sections}
							l3section.Number = number
							number++

							for _, l4name := range l4sections.Keys() {
								l4sections.ByKey(l4name).(*Section).Number = number
								number++
							}
						} else {
							l3section.Number = number
							number++
						}
					}
				} else {
					l4sections := parseSections(l2section.Content.(string), wikt.Level4)
					if len(l4sections.Keys()) > 0 {
						l2section.Content = &SubSections{Level: wikt.Level4, Sections: l4sections}
						l2section.Number = number
						number++

						for _, l4name := range l4sections.Keys() {
							l4sections.ByKey(l4name).(*Section).Number = number
							number++
						}
					} else {
						l2section.Number = number
						number++
					}
				}
			}
		} else {
			l3sections := parseSections(l1section.Content.(string), wikt.Level3)
			if len(l3sections.Keys()) > 0 {
				l1section.Content = &SubSections{Level: wikt.Level3, Sections: l3sections}
				l1section.Number = number
				number++

				for _, l3name := range l3sections.Keys() {
					l3section := l3sections.ByKey(l3name).(*Section)
					l4sections := parseSections(l3section.Content.(string), wikt.Level4)
					if len(l4sections.Keys()) > 0 {
						l3section.Content = &SubSections{Level: wikt.Level4, Sections: l4sections}
						l3section.Number = number
						number++

						for _, l4name := range l4sections.Keys() {
							l4sections.ByKey(l4name).(*Section).Number = number
							number++
						}
					} else {
						l3section.Number = number
						number++
					}
				}
			} else {
				l4sections := parseSections(l1section.Content.(string), wikt.Level4)
				if len(l4sections.Keys()) > 0 {
					l1section.Content = &SubSections{Level: wikt.Level4, Sections: l4sections}
					l1section.Number = number
					number++

					for _, l4name := range l4sections.Keys() {
						l4sections.ByKey(l4name).(*Section).Number = number
						number++
					}
				} else {
					l1section.Number = number
					number++
				}
			}
		}
	}

	return l1sections
}

func parseSections(text string, level wikt.Level) *utils.OrderedMap {
	sections := utils.NewOrderedMap()

	matches := wikt.Levels[level].FindAllStringSubmatch(text, -1)
	contents := wikt.Levels[level].Split(text, -1)

	for i := range matches {
		sections.Set(matches[i][1], &Section{Content: contents[i+1]})
	}

	return sections
}

func ParseSemantics(section *Section, translationsHTML string) *utils.OrderedMap {
	var semantics *utils.OrderedMap
	switch content := section.Content.(type) {
	case string:
		semantics = parseII(content)
	case *SubSections:
		semantics = parseI(content.Sections)
	}

	translations := make([]*utils.OrderedMap, len(semantics.Keys()))
	for i := range translations {
		translations[i] = utils.NewOrderedMap()
	}

	if translationsHTML != "" {
		parseTranslations(translationsHTML, translations)
	}
	for i := range semantics.Keys() {
		semantics.ByIndex(i).(*Content).Translations = translations[i]
	}

	return semantics
}

func parseI(sections *utils.OrderedMap) *utils.OrderedMap {
	meaningSection := sections.ByKey(wikt.Meanings).(*Section)
	l, meanings, examples := parseMeanings(strings.TrimSpace(meaningSection.Content.(string)))
	synonyms, antonyms, hyperonyms, hyponyms := make([][]string, l), make([][]string, l), make([][]string, l), make([][]string, l)

	if synonymsSection := sections.ByKey(wikt.Synonyms); synonymsSection != nil {
		parseRelation(strings.TrimSpace(synonymsSection.(*Section).Content.(string)), synonyms)
	}
	if antonymsSection := sections.ByKey(wikt.Antonyms); antonymsSection != nil {
		parseRelation(strings.TrimSpace(antonymsSection.(*Section).Content.(string)), antonyms)
	}
	if hyperonymsSection := sections.ByKey(wikt.Hyperonyms); hyperonymsSection != nil {
		parseRelation(strings.TrimSpace(hyperonymsSection.(*Section).Content.(string)), hyperonyms)
	}
	if hyponymsSection := sections.ByKey(wikt.Hyponyms); hyponymsSection != nil {
		parseRelation(strings.TrimSpace(hyponymsSection.(*Section).Content.(string)), hyponyms)
	}

	result := utils.NewOrderedMap()
	for i, meaning := range meanings {
		result.Set(meaning, &Content{
			Examples:     examples[i],
			Synonyms:     synonyms[i],
			Antonyms:     antonyms[i],
			Hyperonyms:   hyperonyms[i],
			Hyponyms:     hyponyms[i],
			Translations: nil,
		})
	}

	return result
}

func parseMeanings(text string) (int, []string, [][]string) {
	const empty = "Отсутствует пример употребления (см. рекомендации)."

	// Remove proto for now. TODO: keep it?
	proto := wikt.Templates[wikt.Proto].FindString(text)
	text = strings.Replace(text, proto, "", -1)

	lines := strings.Split(text, "\n")
	l := len(lines)
	meanings := make([]string, l)
	examples := make([][]string, l)

	for i := 0; i < l; i++ {
		split := strings.Split(lines[i], "◆")
		meanings[i] = split[0]
		if len(split) > 1 {
			for _, example := range split[1:] {
				example = strings.TrimSpace(example)
				if !strings.Contains(empty, example) {
					examples[i] = append(examples[i], example)
				}
			}
		}
	}

	return l, meanings, examples
}

func parseRelation(text string, relation [][]string) {
	const empty = "-?—"

	lines := strings.Split(text, "\n")
	for i := 0; i < len(relation) && i < len(lines); i++ {
		for _, word := range strings.FieldsFunc(lines[i], func(r rune) bool {
			return r == ',' || r == ';'
		}) {
			word = strings.TrimSpace(word)
			if !strings.Contains(empty, word) {
				relation[i] = append(relation[i], word)
			}
		}
	}
}

func parseII(text string) *utils.OrderedMap {
	result := utils.NewOrderedMap()
	lines := strings.Split(strings.TrimSpace(text), "\n")
	for _, line := range lines {
		split := strings.Split(line, " § ")
		content := &Content{}
		if len(split) > 1 {
			examplesStr := wikt.Templates[wikt.T2Examples].FindStringSubmatch(split[1])
			content.Examples = strings.Split(examplesStr[1], "◆")

			contentStr := strings.Replace(split[1], examplesStr[0], "", -1)
			matches := wikt.Templates[wikt.T2Content].FindAllStringSubmatch(contentStr, -1)
			values := wikt.Templates[wikt.T2Content].Split(contentStr, -1)
			for i := range matches {
				switch matches[i][0] {
				case "синонимы:":
					for _, word := range strings.Split(values[i+1], ",") {
						content.Synonyms = append(content.Synonyms, strings.TrimSpace(word))
					}
				case "антонимы:":
					for _, word := range strings.Split(values[i+1], ",") {
						content.Antonyms = append(content.Antonyms, strings.TrimSpace(word))
					}
				case "гиперонимы:":
					for _, word := range strings.Split(values[i+1], ",") {
						content.Hyperonyms = append(content.Hyperonyms, strings.TrimSpace(word))
					}
				case "гипонимы:":
					for _, word := range strings.Split(values[i+1], ",") {
						content.Hyponyms = append(content.Hyponyms, strings.TrimSpace(word))
					}
				}
			}

		}
		result.Set(split[0], content)
	}

	return result
}

func parseTranslations(translationsHTML string, translations []*utils.OrderedMap) {
	escape := []string{"м.", "ж.", "ср.", "-s", "-e"}

	text, err := html2text.FromString(translationsHTML)
	if err != nil {
		log.Print(err)
	}

	i := -1
	scanner := bufio.NewScanner(strings.NewReader(text))

	// Skip 'Перевод' and '-------'
	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if line[:2] == "* " {
			split := strings.Split(line[2:], ":")
			lang := split[0][:strings.IndexRune(split[0], ' ')]
			wordsStr := wikt.Templates[wikt.TranslLinks].ReplaceAllString(split[1], "")
			for _, esc := range escape {
				if strings.Contains(wordsStr, esc) {
					wordsStr = strings.Replace(wordsStr, esc, "", -1)
				}
			}
			var words []string
			for _, word := range strings.FieldsFunc(wordsStr, func(r rune) bool {
				return r == ',' || r == ';'
			}) {
				word = strings.TrimSpace(word)
				if word != "" {
					words = append(words, word)
				}
			}

			if prevWords := translations[i].ByKey(lang); prevWords != nil {
				words = append(prevWords.([]string), words...)
			}

			translations[i].Set(lang, words)
		} else {
			i++
			if i >= len(translations) {
				return
			}
		}
	}
}

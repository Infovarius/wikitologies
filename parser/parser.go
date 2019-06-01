package parser

import (
	"fmt"
	"strings"

	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

func Parse(title string) (Word, error) {
	text, err := wikt.GetText(title)
	if err != nil {
		return nil, err
	}

	numbers, err := wikt.GetSectionNumbers(title)
	if err != nil {
		return nil, err
	}

	var word Word
	for _, s := range parseText(text, numbers) {
		if s.SubSections == nil {
			continue
		}

		var meanings Meanings
		foreign := s.Header != wikt.Russian
		if s.SubSections[0].Level == wikt.L2 {
			for _, s2 := range s.SubSections {
				ms, err := parseMeanings(title, foreign, s2)
				if err != nil {
					return nil, err
				}

				for _, m := range ms {
					m.Value = fmt.Sprintf("%s: %s", s2.Header, m.Value)
					meanings = append(meanings, m)
				}
			}
		} else {
			var err error
			meanings, err = parseMeanings(title, foreign, s)
			if err != nil {
				return nil, err
			}
		}

		word = append(word, struct {
			Language string
			Meanings Meanings
		}{Language: s.Header, Meanings: meanings})
	}

	return word, nil
}

func parseText(text string, numbers []int) Sections {
	var sections Sections
	stack := make(stack, 0)
	for lvl := wikt.L1; lvl <= wikt.L4; lvl++ {
		if sections = parseSection(text, lvl); len(sections) > 0 {
			stack.push(sections)
			break
		}
	}

	var i int
	for !stack.empty() {
		section := stack.pop()
		section.Number = numbers[i]
		i++

		for lvl := section.Level + 1; lvl <= wikt.L4; lvl++ {
			if subs := parseSection(section.Text, lvl); len(subs) > 0 {
				section.SubSections = subs
				stack.push(subs)
				break
			}
		}
	}

	return sections
}

func parseSection(text string, lvl wikt.Level) Sections {
	headers := wikt.HeadersRE[lvl].FindAllStringSubmatch(text, -1)
	texts := wikt.HeadersRE[lvl].Split(text, -1)

	var sections Sections
	for i := range headers {
		sections = append(sections, &Section{Header: headers[i][1], Text: texts[i+1], Level: lvl})
	}

	return sections
}

func parseMeanings(title string, foreign bool, section *Section) (Meanings, error) {
	var meanings Meanings
	semProps := section.SubSections.ByHeader(wikt.SemProps)
	if semProps == nil {
		return meanings, nil
	}

	if semProps.SubSections != nil {
		if len(semProps.SubSections) > 1 && semProps.SubSections[1].Number == 0 {
			var err error
			meanings, err = parseType3(title, semProps.SubSections[0])
			if err != nil {
				return nil, err
			}
		} else {
			meanings = parseType1(semProps.SubSections)
		}
	} else {
		meanings = parseType2(semProps.Text)
	}

	if foreign {
		if mSection := semProps.SubSections.ByHeader(wikt.Meanings); mSection != nil {
			if err := parseTranslationsForeign(title, mSection.Number, meanings); err != nil {
				return nil, err
			}
		}
	} else {
		if tSection := section.SubSections.ByHeader(wikt.Translations); tSection != nil {
			if err := parseTranslationsRu(title, tSection.Number, meanings); err != nil {
				return nil, err
			}
		}
	}

	return meanings, nil
}

func parseType1(sections Sections) Meanings {
	mSection := sections.ByHeader(wikt.Meanings)
	if mSection == nil {
		return nil
	}

	values, examples := parseMeaningsSection(mSection.Text)
	l := len(values)
	synonyms, antonyms, hyperonyms, hyponyms := make([][]string, l), make([][]string, l), make([][]string, l), make([][]string, l)

	if sSection := sections.ByHeader(wikt.Synonyms); sSection != nil {
		parseRelationsSection(sSection.Text, synonyms)
	}
	if aSection := sections.ByHeader(wikt.Antonyms); aSection != nil {
		parseRelationsSection(aSection.Text, antonyms)
	}
	if hyperSection := sections.ByHeader(wikt.Hyperonyms); hyperSection != nil {
		parseRelationsSection(hyperSection.Text, hyperonyms)
	}
	if hypoSection := sections.ByHeader(wikt.Hyponyms); hypoSection != nil {
		parseRelationsSection(hypoSection.Text, hyponyms)
	}

	var meanings Meanings
	for i, value := range values {
		meanings = append(meanings, &Meaning{
			Value:      value,
			Examples:   examples[i],
			Synonyms:   synonyms[i],
			Antonyms:   antonyms[i],
			Hyperonyms: hyperonyms[i],
			Hyponyms:   hyponyms[i],
		})
	}

	return meanings
}

func parseType2(text string) Meanings {
	var meanings Meanings
	for _, line := range strings.Split(trim(text), "\n") {
		split := strings.Split(line, " § ")
		meaning := &Meaning{Value: split[0]}
		if len(split) > 1 {
			headers := wikt.TemplatesRE[wikt.T2Content].FindAllStringSubmatch(split[1], -1)
			values := wikt.TemplatesRE[wikt.T2Content].Split(split[1], -1)

			for _, example := range strings.Split(values[0], wikt.ExampleSep) {
				example = trim(example)
				if example != "" && example != wikt.MissingExample {
					meaning.Examples = append(meaning.Examples, example)
				}
			}

			for i := range headers {
				switch headers[i][0] {
				case "синонимы:":
					for _, word := range strings.Split(values[i+1], ",") {
						meaning.Synonyms = append(meaning.Synonyms, trim(word))
					}
				case "антонимы:":
					for _, word := range strings.Split(values[i+1], ",") {
						meaning.Antonyms = append(meaning.Antonyms, trim(word))
					}
				case "гиперонимы:":
					for _, word := range strings.Split(values[i+1], ",") {
						meaning.Hyperonyms = append(meaning.Hyperonyms, trim(word))
					}
				case "гипонимы:":
					for _, word := range strings.Split(values[i+1], ",") {
						meaning.Hyponyms = append(meaning.Hyponyms, trim(word))
					}
				}
			}
		}

		meanings = append(meanings, meaning)
	}

	return meanings
}

func parseType3(title string, mSection *Section) (Meanings, error) {
	wikitext, err := wikt.GetWikitext(title, mSection.Number)
	if err != nil {
		return nil, err
	}

	values, examples := parseMeaningsSection(mSection.Text)
	l := len(values)
	synonyms, antonyms, hyperonyms, hyponyms := make([][]string, l), make([][]string, l), make([][]string, l), make([][]string, l)

	for i, match := range wikt.TemplatesRE[wikt.Semantics].FindAllStringSubmatch(wikitext, -1) {
		match[1] = wikt.TemplatesRE[wikt.Brackets].ReplaceAllString(match[1], "")
		for _, eq := range strings.Split(match[1], "|") {
			var values []string
			split := strings.Split(eq, "=")
			for _, v := range strings.FieldsFunc(split[1], func(r rune) bool {
				return r == ',' || r == ';'
			}) {
				v = trim(v)
				v = strings.Replace(v, "[[", "", -1)
				v = strings.Replace(v, "]]", "", -1)
				if !strings.Contains("-?—", v) {
					values = append(values, v)
				}
			}

			switch split[0] {
			case "синонимы":
				synonyms[i] = values
			case "антонимы":
				antonyms[i] = values
			case "гиперонимы":
				hyperonyms[i] = values
			case "гипонимы":
				hyponyms[i] = values
			}
		}
	}

	var meanings Meanings
	for i, value := range values {
		meanings = append(meanings, &Meaning{
			Value:      value,
			Examples:   examples[i],
			Synonyms:   synonyms[i],
			Antonyms:   antonyms[i],
			Hyperonyms: hyperonyms[i],
			Hyponyms:   hyponyms[i],
		})
	}

	return meanings, nil
}

func parseMeaningsSection(text string) ([]string, [][]string) {
	var meanings []string
	var examples [][]string
	for _, line := range strings.Split(trim(text), "\n") {
		if line == "" || strings.Contains(line, wikt.Proto) {
			continue
		}

		split := strings.Split(line, wikt.ExampleSep)
		meanings = append(meanings, split[0])
		examples = append(examples, make([]string, 0))
		last := len(examples) - 1
		if len(split) > 1 {
			for _, example := range split[1:] {
				example = trim(example)
				if !strings.Contains(wikt.MissingExample, example) {
					examples[last] = append(examples[last], example)
				}
			}
		}
	}

	return meanings, examples
}

func parseTranslationsRu(title string, number int, meanings Meanings) error {
	wikitext, err := wikt.GetWikitext(title, number)
	if err != nil {
		return err
	}

	for wikt.TemplatesRE[wikt.Brackets].MatchString(wikitext) {
		wikitext = wikt.TemplatesRE[wikt.Brackets].ReplaceAllString(wikitext, "")
	}
	for wikt.TemplatesRE[wikt.HTMLcomment].MatchString(wikitext) {
		wikitext = wikt.TemplatesRE[wikt.HTMLcomment].ReplaceAllString(wikitext, "")
	}

	for i, block := range wikt.TemplatesRE[wikt.TranslationsRU].Split(wikitext, -1)[1:] {
		if i >= len(meanings) {
			return nil
		}

		for wikt.TemplatesRE[wikt.Template].MatchString(block) {
			block = wikt.TemplatesRE[wikt.Template].ReplaceAllString(block, "")
		}

		for _, line := range strings.Split(block, "\n") {
			if line == "" || line[0] != '|' {
				continue
			}

			split := strings.Split(line[1:], "=")
			lang, ok := languages[split[0]]
			if !ok {
				continue
			}

			var values []string
			for _, v := range strings.FieldsFunc(split[1], func(r rune) bool {
				return r == ',' || r == ';'
			}) {
				matches := wikt.TemplatesRE[wikt.Link].FindAllStringSubmatch(v, -1)
				if len(matches) == 1 {
					values = append(values, matches[0][1])
				}
			}

			if len(values) > 0 {
				meanings[i].Translations = append(meanings[i].Translations, &Translation{
					Language: lang,
					Values:   values,
				})
			}
		}
	}

	return nil
}

func parseTranslationsForeign(title string, number int, meanings Meanings) error {
	wikitext, err := wikt.GetWikitext(title, number)
	if err != nil {
		return err
	}

	for wikt.TemplatesRE[wikt.Brackets].MatchString(wikitext) {
		wikitext = wikt.TemplatesRE[wikt.Brackets].ReplaceAllString(wikitext, "")
	}
	for wikt.TemplatesRE[wikt.HTMLcomment].MatchString(wikitext) {
		wikitext = wikt.TemplatesRE[wikt.HTMLcomment].ReplaceAllString(wikitext, "")
	}

	for i, match := range wikt.TemplatesRE[wikt.Meaning].FindAllStringSubmatch(wikitext, -1) {
		if i >= len(meanings) {
			return nil
		}

		for wikt.TemplatesRE[wikt.Template].MatchString(match[1]) {
			match[1] = wikt.TemplatesRE[wikt.Template].ReplaceAllString(match[1], "")
		}

		var values []string
		for _, v := range strings.FieldsFunc(match[1], func(r rune) bool {
			return r == ',' || r == ';'
		}) {
			matches := wikt.TemplatesRE[wikt.Link].FindAllStringSubmatch(v, -1)
			if len(matches) == 1 {
				values = append(values, matches[0][1])
			}
		}

		if len(values) > 0 {
			meanings[i].Translations = append(meanings[i].Translations, &Translation{
				Language: wikt.Russian,
				Values:   values,
			})
		}
	}

	return nil
}

func parseRelationsSection(text string, relations [][]string) {
	lines := strings.Split(trim(text), "\n")
	for i := 0; i < len(relations) && i < len(lines); i++ {
		for _, word := range strings.FieldsFunc(lines[i], func(r rune) bool {
			return r == ',' || r == ';'
		}) {
			word = trim(word)
			if !strings.Contains("-?—", word) {
				relations[i] = append(relations[i], word)
			}
		}
	}
}

func trim(s string) string {
	return strings.TrimSpace(s)
}

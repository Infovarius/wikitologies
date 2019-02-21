package parser

import (
	"bufio"
	"log"
	"strings"

	"github.com/jaytaylor/html2text"

	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

func ParseText(text string) Sections {
	number := 1
	l1sections := parseSection(text, wikt.L1)

	for _, l1section := range l1sections {
		l2sections := parseSection(l1section.Text, wikt.L2)
		if len(l2sections) > 0 {
			l1section.SubSections = l2sections
			l1section.Number = number
			number++

			for _, l2section := range l2sections {
				l3sections := parseSection(l2section.Text, wikt.L3)
				if len(l3sections) > 0 {
					l2section.SubSections = l3sections
					l2section.Number = number
					number++

					for _, l3section := range l3sections {
						l4sections := parseSection(l3section.Text, wikt.L4)
						if len(l4sections) > 0 {
							l3section.SubSections = l4sections
							l3section.Number = number
							number++

							for _, l4section := range l4sections {
								l4section.Number = number
								number++
							}
						} else {
							l3section.Number = number
							number++
						}
					}
				} else {
					l4sections := parseSection(l2section.Text, wikt.L4)
					if len(l4sections) > 0 {
						l2section.SubSections = l4sections
						l2section.Number = number
						number++

						for _, l4section := range l4sections {
							l4section.Number = number
							number++
						}
					} else {
						l2section.Number = number
						number++
					}
				}
			}
		} else {
			l3sections := parseSection(l1section.Text, wikt.L3)
			if len(l3sections) > 0 {
				l1section.SubSections = l3sections
				l1section.Number = number
				number++

				for _, l3section := range l3sections {
					l4sections := parseSection(l3section.Text, wikt.L4)
					if len(l4sections) > 0 {
						l3section.SubSections = l4sections
						l3section.Number = number
						number++

						for _, l4section := range l4sections {
							l4section.Number = number
							number++
						}
					} else {
						l3section.Number = number
						number++
					}
				}
			} else {
				l4sections := parseSection(l1section.Text, wikt.L4)
				if len(l4sections) > 0 {
					l1section.SubSections = l4sections
					l1section.Number = number
					number++

					for _, l4section := range l4sections {
						l4section.Number = number
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

func ParseMeanings(semProps *Section) Meanings {
	if semProps.SubSections != nil {
		return parseI(semProps.SubSections)
	} else {
		return parseII(semProps.Text)
	}
}

func ParseTranslations(meanings Meanings, foreign bool, html string) {
	text, err := html2text.FromString(html)
	if err != nil {
		log.Print(err)
		return
	}

	translations := make([]Translations, len(meanings))
	if foreign {
		parseTranslationsForeign(text, translations)
	} else {
		parseTranslations(text, translations)
	}

	for i := range meanings {
		meanings[i].Translations = translations[i]
	}
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

func parseI(sections Sections) Meanings {
	section := sections.ByHeader(wikt.Meanings)
	if section == nil {
		return nil
	}
	values, examples := parseMeanings(section.Text)

	l := len(values)
	synonyms, antonyms, hyperonyms, hyponyms := make([][]string, l), make([][]string, l), make([][]string, l), make([][]string, l)

	if section = sections.ByHeader(wikt.Synonyms); section != nil {
		parseRelations(section.Text, synonyms)
	}
	if section := sections.ByHeader(wikt.Antonyms); section != nil {
		parseRelations(section.Text, antonyms)
	}
	if section := sections.ByHeader(wikt.Hyperonyms); section != nil {
		parseRelations(section.Text, hyperonyms)
	}
	if section := sections.ByHeader(wikt.Hyponyms); section != nil {
		parseRelations(section.Text, hyponyms)
	}

	var meanings Meanings
	for i, value := range values {
		meanings = append(meanings, &Meaning{
			Value:        value,
			Examples:     examples[i],
			Synonyms:     synonyms[i],
			Antonyms:     antonyms[i],
			Hyperonyms:   hyperonyms[i],
			Hyponyms:     hyponyms[i],
			Translations: nil,
		})
	}

	return meanings
}

func parseII(text string) Meanings {
	var meanings Meanings
	for _, line := range strings.Split(trim(text), "\n") {
		split := strings.Split(line, " § ")
		meaning := &Meaning{Value: split[0]}
		if len(split) > 1 {
			examples := wikt.TemplatesRE[wikt.T2Examples].FindStringSubmatch(split[1])
			meaning.Examples = strings.Split(examples[1], wikt.ExampleSep)

			content := strings.Replace(split[1], examples[0], "", -1)
			headers := wikt.TemplatesRE[wikt.T2Content].FindAllStringSubmatch(content, -1)
			values := wikt.TemplatesRE[wikt.T2Content].Split(content, -1)
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

func parseMeanings(text string) ([]string, [][]string) {
	var meanings []string
	var examples [][]string
	for _, line := range strings.Split(trim(text), "\n") {
		if line == "" || line[0] == '(' || strings.Contains(line, wikt.Proto) {
			continue
		}

		split := strings.Split(line, wikt.ExampleSep)
		meanings = append(meanings, split[0])
		examples = append(examples, make([]string, 0))
		i := len(examples) - 1
		if len(split) > 1 {
			for _, example := range split[1:] {
				example = trim(example)
				if !strings.Contains(wikt.MissingExample, example) {
					examples[i] = append(examples[i], example)
				}
			}
		}
	}

	return meanings, examples
}

func parseRelations(text string, relations [][]string) {
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

func parseTranslations(text string, ts []Translations) {
	i := -1
	scanner := bufio.NewScanner(strings.NewReader(text))

	// Skip 'Перевод' and '-------'
	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case line == "":
			continue
		case strings.HasPrefix(line, "* "):
			matches := wikt.TemplatesRE[wikt.LinkedWord].FindAllStringSubmatch(line, -1)
			if len(matches) == 0 {
				continue
			}

			var values []string
			for _, value := range matches[1:] {
				values = append(values, value[1])
			}

			if len(values) > 0 {
				ts[i] = append(ts[i], &Translation{Language: matches[0][1], Values: values})
			}
		default:
			i++
			if i >= len(ts) {
				return
			}
		}
	}
}

func parseTranslationsForeign(text string, ts []Translations) {
	idx := strings.Index(text, wikt.Meanings+" * ")
	if idx == -1 {
		return
	}

	i := 0
	text = strings.TrimPrefix(text[idx:], wikt.Meanings+" ")
	scanner := bufio.NewScanner(strings.NewReader(text))

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "* ") {
			return
		}

		// Cut off examples
		if i := strings.Index(line, wikt.ExampleSep); i != -1 {
			line = line[:i]
		}

		var values []string
		for _, field := range strings.FieldsFunc(line, func(r rune) bool {
			return r == ',' || r == ';'
		}) {
			field = strings.Replace(field, "сленг", "", -1)
			field = strings.Replace(field, "табу", "", -1)

			matches := wikt.TemplatesRE[wikt.LinkedWordRu].FindAllStringSubmatch(field, -1)
			if len(matches) != 1 {
				continue
			}

			values = append(values, matches[0][1])
		}

		if len(values) > 0 {
			ts[i] = append(ts[i], &Translation{Language: wikt.Russian, Values: values})
		}

		i++
	}
}

func trim(text string) string {
	return strings.TrimSpace(text)
}

package wiktionary

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type numbersResponse struct {
	Parse struct {
		Sections []struct {
			Line   string
			Level  string
			Number string
			Index  string
		}
	}
}

type textResponse struct {
	Query struct {
		Pages []struct {
			Title   string
			Extract string
			Missing bool
		}
	}
}

type wikitextResponse struct {
	Parse struct {
		Wikitext string
	}
}

const apiUrl = "https://ru.wiktionary.org/w/api.php?"

var ErrMissing = errors.New("page is missing")

func GetSectionNumbers(title string) ([]int, error) {
	params := url.Values{}
	params.Add("action", "parse")
	params.Add("prop", "sections")
	params.Add("redirects", "1")
	params.Add("format", "json")
	params.Add("formatversion", "2")
	params.Add("page", title)
	params.Add("disablelimitreport", "1")
	params.Add("disableeditsection", "1")
	params.Add("disablestylededuplication", "1")

	bytes, err := get(apiUrl + params.Encode())
	if err != nil {
		return nil, err
	}

	var data numbersResponse
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	l := len(data.Parse.Sections)
	numbers := make([]int, l)
	for i, s := range data.Parse.Sections {
		if s.Index != "" {
			numbers[i], _ = strconv.Atoi(s.Index)
		}
	}

	return numbers, nil
}

func GetText(title string) (string, error) {
	params := url.Values{}
	params.Add("action", "query")
	params.Add("prop", "extracts")
	params.Add("explaintext", "1")
	params.Add("redirects", "1")
	params.Add("format", "json")
	params.Add("formatversion", "2")
	params.Add("titles", title)

	bytes, err := get(apiUrl + params.Encode())
	if err != nil {
		return "", err
	}

	var data textResponse
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return "", err
	}

	if data.Query.Pages[0].Missing {
		return "", ErrMissing
	}

	return data.Query.Pages[0].Extract, nil
}

func GetWikitext(title string, number int) (string, error) {
	params := url.Values{}
	params.Add("action", "parse")
	params.Add("prop", "wikitext")
	params.Add("redirects", "1")
	params.Add("format", "json")
	params.Add("formatversion", "2")
	params.Add("page", title)
	params.Add("section", strconv.Itoa(number))
	params.Add("disablelimitreport", "1")
	params.Add("disableeditsection", "1")
	params.Add("disablestylededuplication", "1")

	bytes, err := get(apiUrl + params.Encode())
	if err != nil {
		return "", err
	}

	var data wikitextResponse
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return "", err
	}

	return data.Parse.Wikitext, nil
}

func get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

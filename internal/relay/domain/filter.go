package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Filter struct {
	IDs     []string `json:"ids,omitempty"`
	Authors []string `json:"authors,omitempty"`
	Kinds   []int    `json:"kinds,omitempty"`

	Tags map[string][]string `json:"-"` // tag filters: "#e", "#p", "#t" など

	Since *int64 `json:"since,omitempty"`
	Until *int64 `json:"until,omitempty"`
	Limit *int   `json:"limit,omitempty"`

	Raw map[string]any `json:"-"` // for unknown key
}

func NewFilterFromRaw(raw map[string]any) (Filter, error) {
	var f Filter
	b, err := json.Marshal(raw)
	if err != nil {
		return f, err
	}

	// IDs, Authors, Kinds, Since, Until, Limit はこれで定義する
	if err := json.Unmarshal(b, &f); err != nil {
		return f, err
	}

	// タグフィルター (#e, #p, #t など) をパース
	f.Tags = make(map[string][]string)
	for key, value := range raw {
		if len(key) > 0 && key[0] == '#' {
			tagName := key[1:] // "#e" -> "e"

			// 値を []string に変換
			if arr, ok := value.([]any); ok {
				strArr := make([]string, 0, len(arr))
				for _, v := range arr {
					if str, ok := v.(string); ok {
						strArr = append(strArr, str)
					}
				}
				f.Tags[tagName] = strArr
			}
		}
	}

	f.Raw = raw
	return f, nil
}

// NewFiltersFromRaw parses multiple raw filter JSON messages into domain Filters.
// Returns successfully parsed filters and an error if any parsing failed.
// If some filters fail to parse, successfully parsed filters are still returned,
// but an error containing all parsing errors is also returned.
func NewFiltersFromRaw(rawFilters []json.RawMessage) ([]Filter, error) {
	var filters []Filter
	var parseErrors []error

	for i, raw := range rawFilters {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("filter %d: invalid JSON: %w", i, err))
			continue
		}

		f, err := NewFilterFromRaw(m)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("filter %d: %w", i, err))
			continue
		}

		filters = append(filters, f)
	}

	// 複数エラーを結合
	if len(parseErrors) > 0 {
		return filters, errors.Join(parseErrors...)
	}

	return filters, nil
}

// Matches returns whether the event satisfies this filter.
// 1フィルターは AND 条件であることに注意
func (f Filter) Matches(evt Event) bool {
	// IDs filter
	if len(f.IDs) > 0 {
		found := false
		for _, id := range f.IDs {
			if evt.ID == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Authors filter
	if len(f.Authors) > 0 {
		found := false
		for _, author := range f.Authors {
			if evt.PubKey == author {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Kinds filter
	if len(f.Kinds) > 0 {
		found := false
		for _, kind := range f.Kinds {
			if evt.Kind == kind {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Since filter
	if f.Since != nil && evt.CreatedAt < *f.Since {
		return false // since より古い場合
	}

	// Until filter
	if f.Until != nil && evt.CreatedAt > *f.Until {
		return false // until より新しい場合
	}

	// Tags filter (#e, #p, #t, etc.)
	for fillterTagName, filterTagValues := range f.Tags {
		// ex. tagName = "e" .. #e
		// ex. tagValues = ["event1", "event2"]
		found := false
		for _, tag := range evt.Tags {
			evtTagName := tag[0]
			evtTagValue := tag[1]
			if len(tag) >= 2 && evtTagName == fillterTagName {
				// Check if any of the tag values match
				for _, value := range filterTagValues {
					if evtTagValue == value {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
		}
		if !found {
			return false
		}
	}

	return true
}

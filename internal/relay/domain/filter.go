package domain

import "encoding/json"

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

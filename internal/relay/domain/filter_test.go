package domain_test

import (
	"encoding/json"
	"nostar/internal/relay/domain"
	"reflect"
	"testing"
)

func inttoPtr(i int) *int {
	return &i
}
func int64toPtr(i int64) *int64 {
	return &i
}
func TestNewFilterFromRaw(t *testing.T) {
	tests := []struct {
		name    string
		raw     map[string]any
		want    domain.Filter
		wantErr bool
	}{
		{
			name: "normal (no tags)",
			raw: map[string]any{
				"ids":     []string{"id1", "id2"},
				"authors": []string{"author1", "author2"},
				"kinds":   []int64{1, 2},
				"since":   1633072800,
				"until":   1633159200,
				"limit":   10,
			},
			want: domain.Filter{
				IDs:     []string{"id1", "id2"},
				Authors: []string{"author1", "author2"},
				Kinds:   []int{1, 2},
				Tags:    map[string][]string{},
				Since:   int64toPtr(1633072800),
				Until:   int64toPtr(1633159200),
				Limit:   inttoPtr(10),
				Raw: map[string]any{
					"ids":     []string{"id1", "id2"},
					"authors": []string{"author1", "author2"},
					"kinds":   []int64{1, 2},
					"since":   1633072800,
					"until":   1633159200,
					"limit":   10,
				},
			},
		},
		{
			name: "with tags",
			raw: map[string]any{
				"ids":     []any{"id1"},
				"authors": []any{"author1"},
				"#e":      []any{"event1", "event2"},
				"#p":      []any{"pubkey1"},
				"#t":      []any{"nostr", "bitcoin"},
				"limit":   5,
			},
			want: domain.Filter{
				IDs:     []string{"id1"},
				Authors: []string{"author1"},
				Tags: map[string][]string{
					"e": {"event1", "event2"},
					"p": {"pubkey1"},
					"t": {"nostr", "bitcoin"},
				},
				Limit: inttoPtr(5),
				Raw: map[string]any{
					"ids":     []any{"id1"},
					"authors": []any{"author1"},
					"#e":      []any{"event1", "event2"},
					"#p":      []any{"pubkey1"},
					"#t":      []any{"nostr", "bitcoin"},
					"limit":   5,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := domain.NewFilterFromRaw(tt.raw)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewFilterFromRaw() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewFilterFromRaw() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFilterFromRaw() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_Matches(t *testing.T) {
	// Helper function to create test event
	createEvent := func(id, pubkey string, createdAt int64, kind int, tags [][]string) domain.Event {
		return domain.Event{
			ID:        id,
			PubKey:    pubkey,
			CreatedAt: createdAt,
			Kind:      kind,
			Tags:      tags,
		}
	}

	tests := []struct {
		name   string
		filter domain.Filter
		event  domain.Event
		want   bool
	}{
		{
			name:   "empty filter matches any event",
			filter: domain.Filter{},
			event:  createEvent("id1", "pub1", 1000, 1, nil),
			want:   true,
		},
		{
			name: "IDs filter - match",
			filter: domain.Filter{
				IDs: []string{"id1", "id2"},
			},
			event: createEvent("id1", "pub1", 1000, 1, nil),
			want:  true,
		},
		{
			name: "IDs filter - no match",
			filter: domain.Filter{
				IDs: []string{"id1", "id2"},
			},
			event: createEvent("id3", "pub1", 1000, 1, nil),
			want:  false,
		},
		{
			name: "Authors filter - match",
			filter: domain.Filter{
				Authors: []string{"pub1", "pub2"},
			},
			event: createEvent("id1", "pub1", 1000, 1, nil),
			want:  true,
		},
		{
			name: "Authors filter - no match",
			filter: domain.Filter{
				Authors: []string{"pub1", "pub2"},
			},
			event: createEvent("id1", "pub3", 1000, 1, nil),
			want:  false,
		},
		{
			name: "Kinds filter - match",
			filter: domain.Filter{
				Kinds: []int{1, 2},
			},
			event: createEvent("id1", "pub1", 1000, 1, nil),
			want:  true,
		},
		{
			name: "Kinds filter - no match",
			filter: domain.Filter{
				Kinds: []int{1, 2},
			},
			event: createEvent("id1", "pub1", 1000, 3, nil),
			want:  false,
		},
		{
			name: "Since filter - match",
			filter: domain.Filter{
				Since: int64toPtr(500),
			},
			event: createEvent("id1", "pub1", 1000, 1, nil),
			want:  true,
		},
		{
			name: "Since filter - no match (too old)",
			filter: domain.Filter{
				Since: int64toPtr(1500),
			},
			event: createEvent("id1", "pub1", 1000, 1, nil),
			want:  false,
		},
		{
			name: "Until filter - match",
			filter: domain.Filter{
				Until: int64toPtr(1500),
			},
			event: createEvent("id1", "pub1", 1000, 1, nil),
			want:  true,
		},
		{
			name: "Until filter - no match (too new)",
			filter: domain.Filter{
				Until: int64toPtr(500),
			},
			event: createEvent("id1", "pub1", 1000, 1, nil),
			want:  false,
		},
		{
			name: "Tags filter - match #e tag",
			filter: domain.Filter{
				Tags: map[string][]string{
					"e": {"event1", "event2"},
				},
			},
			event: createEvent("id1", "pub1", 1000, 1, [][]string{
				{"e", "event1"},
				{"p", "pubkey1"},
			}),
			want: true,
		},
		{
			name: "Tags filter - match #p tag",
			filter: domain.Filter{
				Tags: map[string][]string{
					"p": {"pubkey1", "pubkey2"},
				},
			},
			event: createEvent("id1", "pub1", 1000, 1, [][]string{
				{"e", "event1"},
				{"p", "pubkey1"},
			}),
			want: true,
		},
		{
			name: "Tags filter - no match",
			filter: domain.Filter{
				Tags: map[string][]string{
					"e": {"event1", "event2"},
				},
			},
			event: createEvent("id1", "pub1", 1000, 1, [][]string{
				{"e", "event3"},
				{"p", "pubkey1"},
			}),
			want: false,
		},
		{
			name: "Multiple tags - all match",
			filter: domain.Filter{
				Tags: map[string][]string{
					"e": {"event1"},
					"p": {"pubkey1"},
				},
			},
			event: createEvent("id1", "pub1", 1000, 1, [][]string{
				{"e", "event1"},
				{"p", "pubkey1"},
			}),
			want: true,
		},
		{
			name: "Multiple tags - one doesn't match",
			filter: domain.Filter{
				Tags: map[string][]string{
					"e": {"event1"},
					"p": {"pubkey2"}, // doesn't match
				},
			},
			event: createEvent("id1", "pub1", 1000, 1, [][]string{
				{"e", "event1"},
				{"p", "pubkey1"},
			}),
			want: false,
		},
		{
			name: "Complex filter - all conditions match",
			filter: domain.Filter{
				IDs:     []string{"id1"},
				Authors: []string{"pub1"},
				Kinds:   []int{1},
				Since:   int64toPtr(500),
				Until:   int64toPtr(1500),
				Tags: map[string][]string{
					"e": {"event1"},
				},
			},
			event: createEvent("id1", "pub1", 1000, 1, [][]string{
				{"e", "event1"},
			}),
			want: true,
		},
		{
			name: "Complex filter - one condition fails",
			filter: domain.Filter{
				IDs:     []string{"id1"},
				Authors: []string{"pub1"},
				Kinds:   []int{2}, // doesn't match (event kind is 1)
				Since:   int64toPtr(500),
				Until:   int64toPtr(1500),
				Tags: map[string][]string{
					"e": {"event1"},
				},
			},
			event: createEvent("id1", "pub1", 1000, 1, [][]string{
				{"e", "event1"},
			}),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Matches(tt.event)
			if got != tt.want {
				t.Errorf("Filter.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewFiltersFromRaw(t *testing.T) {
	// Helper function to create json.RawMessage from JSON string
	toRawMessage := func(jsonStr string) json.RawMessage {
		return json.RawMessage(jsonStr)
	}

	tests := []struct {
		name        string
		rawFilters  []json.RawMessage
		wantFilters []domain.Filter
		wantErr     bool
	}{
		{
			name: "multiple valid filters",
			rawFilters: []json.RawMessage{
				toRawMessage(`{"ids": ["id1"], "authors": ["author1"]}`),
				toRawMessage(`{"kinds": [1], "since": 1000}`),
			},
			wantFilters: []domain.Filter{
				{
					IDs:     []string{"id1"},
					Authors: []string{"author1"},
					Tags:    map[string][]string{},
				},
				{
					Kinds: []int{1},
					Since: int64toPtr(1000),
					Tags:  map[string][]string{},
				},
			},
			wantErr: false,
		},
		{
			name: "filters with tags",
			rawFilters: []json.RawMessage{
				toRawMessage(`{"#e": ["event1"], "#p": ["pubkey1"]}`),
				toRawMessage(`{"#t": ["nostr", "bitcoin"]}`),
			},
			wantFilters: []domain.Filter{
				{
					Tags: map[string][]string{
						"e": {"event1"},
						"p": {"pubkey1"},
					},
				},
				{
					Tags: map[string][]string{
						"t": {"nostr", "bitcoin"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "some invalid JSON",
			rawFilters: []json.RawMessage{
				toRawMessage(`{"ids": ["id1"]}`), // valid
				toRawMessage(`invalid json`),     // invalid
				toRawMessage(`{"kinds": [1]}`),   // valid
			},
			wantFilters: []domain.Filter{
				{
					IDs:  []string{"id1"},
					Tags: map[string][]string{},
				},
				{
					Kinds: []int{1},
					Tags:  map[string][]string{},
				},
			},
			wantErr: true, // 1つのフィルタが失敗
		},
		{
			name: "all invalid filters",
			rawFilters: []json.RawMessage{
				toRawMessage(`not json`),
				toRawMessage(`also not json`),
			},
			wantFilters: []domain.Filter{}, // 成功したフィルタなし
			wantErr:     true,
		},
		{
			name:        "empty filters",
			rawFilters:  []json.RawMessage{},
			wantFilters: []domain.Filter{},
			wantErr:     false,
		},
		{
			name: "invalid filter structure",
			rawFilters: []json.RawMessage{
				toRawMessage(`{"ids": ["id1"]}`),       // valid
				toRawMessage(`{"kinds": ["not_int"]}`), // kinds should be []int, not []string
			},
			wantFilters: []domain.Filter{
				{
					IDs:  []string{"id1"},
					Tags: map[string][]string{},
				},
			},
			wantErr: true, // kinds の型変換エラー
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFilters, gotErr := domain.NewFiltersFromRaw(tt.rawFilters)

			if gotErr != nil != tt.wantErr {
				t.Errorf("NewFiltersFromRaw() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			if len(gotFilters) != len(tt.wantFilters) {
				t.Errorf("NewFiltersFromRaw() returned %d filters, want %d", len(gotFilters), len(tt.wantFilters))
				return
			}

			for i, got := range gotFilters {
				want := tt.wantFilters[i]
				// Raw field is not important for equality comparison
				got.Raw = nil
				want.Raw = nil
				if !reflect.DeepEqual(got, want) {
					t.Errorf("NewFiltersFromRaw() filter[%d] = %v, want %v", i, got, want)
				}
			}
		})
	}
}

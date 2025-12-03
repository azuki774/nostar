package domain_test

import (
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

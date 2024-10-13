package def

import (
	"encoding/json"
	"testing"
)

func TestImportanceLevelUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    ImportanceLevel
		wantErr bool
	}{
		{
			name:  "string value",
			input: `"domain_general"`,
			want:  DomainGeneral,
		},
		{
			name:  "numeric value",
			input: `36`,
			want:  GlobalMasterPiece,
		},
		{
			name:    "unknown string",
			input:   `"not_a_level"`,
			wantErr: true,
		},
		{
			name:    "unknown numeric value",
			input:   `4`,
			wantErr: true,
		},
		{
			name:    "fractional numeric value",
			input:   `1.5`,
			wantErr: true,
		},
		{
			name:    "negative numeric value",
			input:   `-1`,
			wantErr: true,
		},
		{
			name:    "numeric overflow",
			input:   `256`,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got ImportanceLevel
			err := json.Unmarshal([]byte(tc.input), &got)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Fatalf("json.Unmarshal(%s, ImportanceLevel) error = %v, want error presence = %t", tc.input, err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got != tc.want {
				t.Errorf("json.Unmarshal(%s, ImportanceLevel) = %v, want %v", tc.input, got, tc.want)
			}
			if !got.Valid() {
				t.Errorf("ImportanceLevel(%d).Valid() = false, want true", got)
			}
		})
	}
}

func TestImportanceLevelValid(t *testing.T) {
	testCases := []struct {
		name  string
		input ImportanceLevel
		want  bool
	}{
		{
			name:  "declared value",
			input: GlobalKey,
			want:  true,
		},
		{
			name:  "zero value",
			input: ImportanceLevel(0),
			want:  false,
		},
		{
			name:  "gap value",
			input: ImportanceLevel(4),
			want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.Valid()
			if got != tc.want {
				t.Errorf("ImportanceLevel(%d).Valid() = %t, want %t", tc.input, got, tc.want)
			}
		})
	}
}

package def

import (
	"encoding/json"
	"testing"
)

func TestDifficultyLevelUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    DifficultyLevel
		wantErr bool
	}{
		{
			name:  "string value",
			input: `"novice_normal"`,
			want:  NoviceNormal,
		},
		{
			name:  "numeric value",
			input: `68`,
			want:  MasterExtreme,
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
			var got DifficultyLevel
			err := json.Unmarshal([]byte(tc.input), &got)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Fatalf("json.Unmarshal(%s, DifficultyLevel) error = %v, want error presence = %t", tc.input, err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got != tc.want {
				t.Errorf("json.Unmarshal(%s, DifficultyLevel) = %v, want %v", tc.input, got, tc.want)
			}
			if !got.Valid() {
				t.Errorf("DifficultyLevel(%d).Valid() = false, want true", got)
			}
		})
	}
}

func TestDifficultyLevelValid(t *testing.T) {
	testCases := []struct {
		name  string
		input DifficultyLevel
		want  bool
	}{
		{
			name:  "declared value",
			input: MasterChallenge,
			want:  true,
		},
		{
			name:  "zero value",
			input: DifficultyLevel(0),
			want:  false,
		},
		{
			name:  "gap value",
			input: DifficultyLevel(4),
			want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.Valid()
			if got != tc.want {
				t.Errorf("DifficultyLevel(%d).Valid() = %t, want %t", tc.input, got, tc.want)
			}
		})
	}
}

package def

import (
	"encoding/json"
	"testing"
)

func TestDungeonTypeUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    DungeonType
		wantErr bool
	}{
		{
			name:  "string value",
			input: `"campaign"`,
			want:  DungeonTypeCampaign,
		},
		{
			name:  "numeric value",
			input: `33`,
			want:  DungeonTypeInstance,
		},
		{
			name:    "unknown string",
			input:   `"not_a_type"`,
			wantErr: true,
		},
		{
			name:    "unknown numeric value",
			input:   `4`,
			wantErr: true,
		},
		{
			name:    "rounded-up fractional numeric value",
			input:   `1.0000000000000001`,
			wantErr: true,
		},
		{
			name:    "rounded fractional numeric value below one",
			input:   `0.99999999999999999`,
			wantErr: true,
		},
		{
			name:    "integer exponent numeric value",
			input:   `1e0`,
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
			var got DungeonType
			err := json.Unmarshal([]byte(tc.input), &got)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Fatalf("json.Unmarshal(%s, DungeonType) error = %v, want error presence = %t", tc.input, err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got != tc.want {
				t.Errorf("json.Unmarshal(%s, DungeonType) = %v, want %v", tc.input, got, tc.want)
			}
			if !got.Valid() {
				t.Errorf("DungeonType(%d).Valid() = false, want true", got)
			}
		})
	}
}

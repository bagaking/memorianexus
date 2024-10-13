package utils

import "testing"

func TestPercentageFromNormalizedFloat(t *testing.T) {
	testCases := []struct {
		name  string
		input float64
		want  Percentage
	}{
		{
			name:  "zero",
			input: 0,
			want:  0,
		},
		{
			name:  "fraction",
			input: 0.42,
			want:  42,
		},
		{
			name:  "one",
			input: 1,
			want:  100,
		},
		{
			name:  "negative clamps to zero",
			input: -0.25,
			want:  0,
		},
		{
			name:  "above one clamps to full percentage",
			input: 1.25,
			want:  100,
		},
		{
			name:  "large input clamps before uint8 conversion",
			input: 3,
			want:  100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got Percentage
			got.FromNormalizedFloat(tc.input)

			if got != tc.want {
				t.Errorf("Percentage.FromNormalizedFloat(%v) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

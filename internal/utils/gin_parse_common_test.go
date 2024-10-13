package utils

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGinGetPagerFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name       string
		rawQuery   string
		wantOffset int
		wantLimit  int
	}{
		{
			name:       "defaults",
			wantOffset: 0,
			wantLimit:  10,
		},
		{
			name:       "invalid strings fallback",
			rawQuery:   "page=bad&offset=bad&limit=bad",
			wantOffset: 0,
			wantLimit:  10,
		},
		{
			name:       "negative limit falls back to default",
			rawQuery:   "page=2&limit=-1",
			wantOffset: 10,
			wantLimit:  10,
		},
		{
			name:       "zero limit falls back to default",
			rawQuery:   "page=2&limit=0",
			wantOffset: 10,
			wantLimit:  10,
		},
		{
			name:       "zero page falls back to first page",
			rawQuery:   "page=0&limit=5",
			wantOffset: 0,
			wantLimit:  5,
		},
		{
			name:       "negative page falls back to first page",
			rawQuery:   "page=-1&limit=5",
			wantOffset: 0,
			wantLimit:  5,
		},
		{
			name:       "invalid page uses offset and clamps negative offset",
			rawQuery:   "page=bad&offset=-1&limit=5",
			wantOffset: 0,
			wantLimit:  5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					RawQuery: tc.rawQuery,
				},
			}

			got := GinGetPagerFromQuery(c)

			if got.Offset != tc.wantOffset || got.Limit != tc.wantLimit {
				t.Fatalf("GinGetPagerFromQuery() = offset %d limit %d, want offset %d limit %d",
					got.Offset, got.Limit, tc.wantOffset, tc.wantLimit)
			}
		})
	}
}

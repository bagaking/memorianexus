package analytic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAnalyticsScaffoldHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name        string
		route       string
		wantMessage string
		assertData  func(t *testing.T, data json.RawMessage)
	}{
		{
			name:        "study patterns",
			route:       "studyPatterns",
			wantMessage: "study patterns scaffold",
			assertData: func(t *testing.T, data json.RawMessage) {
				t.Helper()

				var got StudyPatternsResponse
				if err := json.Unmarshal(data, &got); err != nil {
					t.Fatalf("json.Unmarshal(%q) error = %v, want nil", data, err)
				}
				if got.Patterns == nil {
					t.Errorf("GetStudyPatterns() patterns = nil, want empty slice")
				}
				if len(got.Patterns) != 0 {
					t.Errorf("GetStudyPatterns() patterns length = %d, want 0", len(got.Patterns))
				}
			},
		},
		{
			name:        "time spent",
			route:       "timeSpent",
			wantMessage: "time spent scaffold",
			assertData: func(t *testing.T, data json.RawMessage) {
				t.Helper()

				var got TimeSpentResponse
				if err := json.Unmarshal(data, &got); err != nil {
					t.Fatalf("json.Unmarshal(%q) error = %v, want nil", data, err)
				}
				if got.Buckets == nil {
					t.Errorf("GetTimeSpent() buckets = nil, want empty slice")
				}
				if len(got.Buckets) != 0 {
					t.Errorf("GetTimeSpent() buckets length = %d, want 0", len(got.Buckets))
				}
				if got.TotalMinutes != 0 {
					t.Errorf("GetTimeSpent() total_minutes = %d, want 0", got.TotalMinutes)
				}
			},
		},
	}

	router := gin.New()
	svc, err := Init(nil)
	if err != nil {
		t.Fatalf("Init(nil) error = %v, want nil", err)
	}
	svc.ApplyMux(router.Group(routePath("analytic")))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			requestPath := routePath("analytic", tc.route)
			request := httptest.NewRequest(http.MethodGet, requestPath, nil)

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", requestPath, recorder.Code, http.StatusOK)
			}

			var got struct {
				Message string          `json:"message"`
				Data    json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
				t.Fatalf("json.Unmarshal(%q) error = %v, want nil", recorder.Body.String(), err)
			}
			if got.Message != tc.wantMessage {
				t.Errorf("GET %s message = %q, want %q", requestPath, got.Message, tc.wantMessage)
			}
			if len(got.Data) == 0 {
				t.Fatalf("GET %s data length = 0, want structured payload", requestPath)
			}

			tc.assertData(t, got.Data)
		})
	}
}

func routePath(segments ...string) string {
	return "/" + strings.Join(segments, "/")
}

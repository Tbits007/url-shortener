package redirect

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tbits007/url-shortener/internal/lib/logger/handlers/slogdiscard"
	"github.com/Tbits007/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestRedirectHandler(t *testing.T) {
	cases := []struct {
		name         string
		alias        string
		mockURL      string
		mockError    error
		expectedCode int
		expectedURL  string
	}{
		{
			name:         "successful redirect",
			alias:        "test_alias",
			mockURL:      "https://chat.deepseek.com/",
			expectedCode: http.StatusFound,
			expectedURL:  "https://chat.deepseek.com/",
		},
		{
			name:         "empty alias",
			alias:        "",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "url not found",
			alias:        "missing_alias",
			mockError:    storage.ErrURLNotFound,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "internal error",
			alias:        "test_error",
			mockError:    errors.New("database error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

    mockLog := slogdiscard.NewDiscardLogger()

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            mockURLGetter := NewMockURLGetter(t)
			if tc.alias != "" {
				mockURLGetter.On("GetURL", tc.alias).Return(tc.mockURL, tc.mockError).Once()
			}

			handler := New(mockLog, mockURLGetter)
			
            target := fmt.Sprintf("/%s", tc.alias)
            req := httptest.NewRequest(http.MethodGet, target, nil)
            
			r := chi.NewRouter()
            r.Get("/{alias}", handler)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			
			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.expectedCode == http.StatusFound {
				location := w.Header().Get("Location")
				assert.Equal(t, tc.expectedURL, location)
			}

		})
	}
}
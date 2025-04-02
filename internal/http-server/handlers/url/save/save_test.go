package save

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tbits007/url-shortener/internal/lib/logger/handlers/slogdiscard"
	"github.com/Tbits007/url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)


func TestSaveHandler(t *testing.T) {
    cases := []struct {
        name         string
        url          string
        alias        string
        expectedCode int
        mockError    error 
    }{
        {
            name:         "success: with alias",
            url:          "https://github.com/",
            alias:        "test_alias",
            expectedCode: http.StatusOK,
        },
        {
            name:         "success: generate alias",
            url:          "https://google.com/",
            alias:        "",
            expectedCode: http.StatusOK,
        },
        {
            name:         "error: url exists",
            url:          "https://exists.com/",
            alias:        "exists",
            expectedCode: http.StatusBadRequest,
            mockError:   storage.ErrURLExists,
        },
    }

    mockLog := slogdiscard.NewDiscardLogger()

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            mockURLsaver := NewMockURLSaver(t)
            
            if tc.alias == "" {
                // Для случая с генерацией алиаса
                mockURLsaver.On("SaveURL", tc.url, mock.AnythingOfType("string")).
                    Return(tc.mockError).
                    Once()
            } else {
                // Для случая с указанным алиасом
                mockURLsaver.On("SaveURL", tc.url, tc.alias).
                    Return(tc.mockError).
                    Once()
            }

            handler := New(mockLog, mockURLsaver)

            body := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)
            req := httptest.NewRequest(http.MethodPost, "/saveURL", bytes.NewReader([]byte(body)))
            w := httptest.NewRecorder()

            handler(w, req)

            assert.Equal(t, tc.expectedCode, w.Code)
        })
    }
}
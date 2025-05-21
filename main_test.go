package main

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	city := "moscow"
	total := len(cafeList[city])

	requests := []struct {
		count int // значение count в запросе
		want  int // ожидаемое количество кафе в ответе
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, total},
	}

	for _, tc := range requests {
		t.Run("count="+strconv.Itoa(tc.count), func(t *testing.T) {
			req := httptest.NewRequest("GET", "/cafe", nil)
			q := url.Values{}
			q.Add("city", city)
			q.Add("count", strconv.Itoa(tc.count))
			req.URL.RawQuery = q.Encode()

			rec := httptest.NewRecorder()
			mainHandle(rec, req)

			resp := rec.Result()
			require.Equal(t, http.StatusOK, resp.StatusCode, "должен быть статус 200 OK")

			body := rec.Body.String()
			got := 0
			if strings.TrimSpace(body) != "" {
				got = len(strings.Split(body, ","))
			}

			assert.Equal(t, tc.want, got, "неверное количество кафе при count=%d", tc.count)
		})
	}
}

func TestCafeSearch(t *testing.T) {
	requests := []struct {
		search    string // значение параметра search
		wantCount int    // ожидаемое количество кафе
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, tc := range requests {
		t.Run("search="+tc.search, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/cafe", nil)
			q := url.Values{}
			q.Add("city", "moscow")
			q.Add("search", tc.search)
			req.URL.RawQuery = q.Encode()

			rec := httptest.NewRecorder()
			mainHandle(rec, req)

			resp := rec.Result()
			require.Equal(t, http.StatusOK, resp.StatusCode, "ожидается статус 200 OK")

			body := strings.TrimSpace(rec.Body.String())
			gotCafes := []string{}
			if body != "" {
				gotCafes = strings.Split(body, ",")
			}

			assert.Equal(t, tc.wantCount, len(gotCafes), "неверное количество кафе при search=%q", tc.search)

			// Проверка: каждое название содержит подстроку search
			searchLower := strings.ToLower(tc.search)
			for _, name := range gotCafes {
				assert.True(t,
					strings.Contains(strings.ToLower(name), searchLower),
					"в названии %q не найдено подстроки %q",
					name, tc.search,
				)
			}
		})
	}
}

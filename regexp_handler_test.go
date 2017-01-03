package conductor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestRegexpHandler(t *testing.T) {
	testcases := []struct {
		Method string
		Path   string
		Body   url.Values
		Status int
	}{
		{
			Method: http.MethodGet,
			Path:   "/posts",
			Status: http.StatusOK,
		},
		{
			Method: http.MethodGet,
			Path:   "/posts/1",
			Status: http.StatusOK,
		},
		{
			Method: http.MethodPost,
			Path:   "/posts",
			Body: url.Values{
				"title": {"sample post"},
				"body":  {"post body"},
			},
			Status: http.StatusOK,
		},
		{
			Method: http.MethodPut,
			Path:   "/posts/1",
			Body: url.Values{
				"body": {"updated post body"},
			},
			Status: http.StatusOK,
		},
		{
			Method: http.MethodDelete,
			Path:   "/posts/1",
			Status: http.StatusOK,
		},
		{
			Method: http.MethodGet,
			Path:   "posts/23/comments",
			Status: http.StatusNotFound,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	// Create RESTful /posts resource
	routes := RegexpRouteMap{}
	routes.AddRoute(http.MethodGet, `/posts[/]?$`, handler)
	routes.AddRoute(http.MethodGet, `/posts/[0-9]+$`, handler)
	routes.AddRoute(http.MethodPost, `/posts[/]?$`, handler)
	routes.AddRoute(http.MethodPut, `/posts/[0-9]+$`, handler)
	routes.AddRoute(http.MethodDelete, `/posts/[0-9]+$`, handler)
	rh := RegexpHandler(routes)

	for _, tc := range testcases {
		body, err := json.Marshal(tc.Body)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(tc.Method, tc.Path, bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		rh.ServeHTTP(w, req)

		if status := w.Code; status != tc.Status {
			t.Errorf("handler status code %v, expected %v", status, tc.Status)
		}
	}
}

package conductor

import (
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

	// Create RESTful /posts resource
	routes := RegexpRouteMap{}
	routes.AddRoute(http.MethodGet, `/posts[/]?$`, indexHandler)
	routes.AddRoute(http.MethodGet, `/posts/[0-9]+$`, showHandler)
	routes.AddRoute(http.MethodPost, `/posts[/]?$`, createHandler)
	routes.AddRoute(http.MethodPut, `/posts/[0-9]+$`, updateHandler)
	routes.AddRoute(http.MethodDelete, `/posts/[0-9]+$`, destroyHandler)
	handler := RegexpHandler(routes)

	for _, tc := range testcases {
		req, err := http.NewRequest(tc.Method, tc.Path, tc.Body)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if status := w.Code; status != tc.Status {
			t.Errorf("handler status code %v, expected %v", status, tc.Status)
		}
	}
}

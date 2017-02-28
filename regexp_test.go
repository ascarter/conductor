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

type testResponse struct {
	Method  string
	Path    string
	Matches []string
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	response := testResponse{Method: r.Method, Path: r.URL.Path}

	matches, ok := RegexpMatchesFromContext(r.Context())
	if ok {
		response.Matches = matches
	}

	output, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

func TestInvalidRegexpMux(t *testing.T) {
	t.Parallel()
	defer func() {
		err := recover()
		t.Logf("recovered panic %v", err)
	}()

	pattern := `/posts/([0-9]+$`
	rh := NewRegexpMux()
	rh.HandleFunc(pattern, testHandler)
	t.Fatalf("Panic expected for %s", pattern)
}

func TestDuplicateRegexpHandler(t *testing.T) {
	t.Parallel()
	defer func() {
		err := recover()
		t.Logf("recovered panic %v", err)
	}()

	pattern := `/posts[/]?$`
	rh := NewRegexpMux()
	rh.HandleFunc(pattern, testHandler)
	rh.HandleFunc(pattern, testHandler)
	t.Fatalf("Panic expected for %s", pattern)
}

func TestRegexpMux(t *testing.T) {
	testcases := []struct {
		Path    string
		Method  string
		Status  int
		Matches []string
		Body    url.Values
	}{
		{
			Path:    "/posts",
			Method:  http.MethodGet,
			Status:  http.StatusOK,
			Matches: []string{"/posts"},
		},
		{
			Path:    "/posts/1",
			Method:  http.MethodGet,
			Status:  http.StatusOK,
			Matches: []string{"/posts/1", "1"},
		},
		{
			Path:    "/posts",
			Method:  http.MethodPost,
			Status:  http.StatusOK,
			Matches: []string{"/posts"},
			Body: url.Values{
				"title": {"sample post"},
				"body":  {"post body"},
			},
		},
		{
			Path:    "/posts/1",
			Method:  http.MethodPut,
			Status:  http.StatusOK,
			Matches: []string{"/posts/1", "1"},
			Body: url.Values{
				"body": {"updated post body"},
			},
		},
		{
			Path:    "/posts/1",
			Method:  http.MethodDelete,
			Status:  http.StatusOK,
			Matches: []string{"/posts/1", "1"},
		},
		{
			Path:    "/posts/23/comments",
			Method:  http.MethodGet,
			Status:  http.StatusOK,
			Matches: []string{"/posts/23/comments", "23"},
		},
		{
			Path:   "/posts/23/foo",
			Method: http.MethodGet,
			Status: http.StatusNotFound,
		},
	}

	rh := NewRegexpMux()
	rh.HandleFunc(`/posts[/]?$`, testHandler)
	rh.HandleFunc(`/posts/([0-9]+)$`, testHandler)
	rh.HandleFunc(`/posts/([0-9]+)/comments$`, testHandler)

	for _, tc := range testcases {
		tc := tc // capture range var
		name := fmt.Sprintf("%s %s", tc.Method, tc.Path)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

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

			status := w.Code
			if status != tc.Status {
				t.Errorf("handler status code %v, expected %v", status, tc.Status)
			}

			if status == http.StatusOK {
				var data testResponse
				err := json.Unmarshal(w.Body.Bytes(), &data)
				if err != nil {
					t.Fatalf("%v: %s", err, w.Body.String())
				}

				if data.Method != tc.Method {
					t.Errorf("handler path %s method %s, expected %s", tc.Path, data.Method, tc.Method)
				}

				if data.Path != tc.Path {
					t.Errorf("handler path %s, expected %s", data.Path, tc.Path)
				}

				if len(data.Matches) != len(tc.Matches) {
					t.Errorf("handler regexp matches %+v, expected %+v", data.Matches, tc.Matches)
				} else {
					for i := 0; i < len(data.Matches); i++ {
						if data.Matches[i] != tc.Matches[i] {
							t.Errorf("handler match %v, expected %v", data.Matches[i], tc.Matches[i])
						}
					}
				}
			}
		})
	}
}

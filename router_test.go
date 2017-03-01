package conductor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type fakeMiddleware struct {
	id int
}

func (fm *fakeMiddleware) Next(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "middleware %d before\n", fm.id)
		h.ServeHTTP(w, r)
		fmt.Fprintf(w, "middleware %d after\n", fm.id)
	})
}

type testResponse struct {
	Method  string
	Path    string
	Matches map[string]string
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	response := testResponse{Method: r.Method, Path: r.URL.Path}

	matches, ok := RouteParamsFromContext(r.Context())
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

func TestNewRouter(t *testing.T) {
	router := NewRouter()
	router.Use(&fakeMiddleware{1})
	router.Use(&fakeMiddleware{2})
	if len(router.components) != 2 {
		t.Error("NewRouter did not add all handlers")
	}
}

func TestDefaultRouter(t *testing.T) {
	expected := "testrequest"

	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, expected)
	}

	router := NewRouter()
	router.HandleFunc("/", fn)

	server := httptest.NewServer(router)
	defer server.Close()

	res, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		t.Fatal(err)
	}

	if string(body) != expected {
		t.Errorf("%v != %v", string(body), expected)
	}
}

func TestUse(t *testing.T) {
	expected := "testrequest"

	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expected)
	}

	mod := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "modified-")
			h.ServeHTTP(w, r)
		})
	}

	router := NewRouter()
	router.Use(ComponentFunc(mod))
	router.HandleFunc("/", fn)

	server := httptest.NewServer(router)
	defer server.Close()

	res, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		t.Fatal(err)
	}

	expectedMessage := fmt.Sprintf("modified-%s", expected)
	if string(body) != expectedMessage {
		t.Errorf("%v != %v", string(body), expectedMessage)
	}
}

func TestInvalidRegexpPattern(t *testing.T) {
	t.Parallel()
	defer func() {
		err := recover()
		t.Logf("recovered panic %v", err)
	}()

	pattern := `/posts/([0-9]+$`
	rh := NewRouterMux()
	rh.HandleFunc(pattern, testHandler)
	t.Fatalf("Panic expected for %s", pattern)
}

func TestDuplicatePattern(t *testing.T) {
	t.Parallel()
	defer func() {
		err := recover()
		t.Logf("recovered panic %v", err)
	}()

	pattern := `/posts[/]?$`
	rh := NewRouterMux()
	rh.HandleFunc(pattern, testHandler)
	rh.HandleFunc(pattern, testHandler)
	t.Fatalf("Panic expected for %s", pattern)
}

func TestRegexpRouterMux(t *testing.T) {
	testcases := []struct {
		Path    string
		Method  string
		Status  int
		Matches map[string]string
		Body    url.Values
	}{
		{
			Path:    "/posts",
			Method:  http.MethodGet,
			Status:  http.StatusOK,
			Matches: map[string]string{"0": "/posts"},
		},
		{
			Path:   "/posts/1",
			Method: http.MethodGet,
			Status: http.StatusOK,
			Matches: map[string]string{
				"0": "/posts/1",
				"1": "1",
			},
		},
		{
			Path:    "/posts",
			Method:  http.MethodPost,
			Status:  http.StatusOK,
			Matches: map[string]string{"0": "/posts"},
			Body: url.Values{
				"title": {"sample post"},
				"body":  {"post body"},
			},
		},
		{
			Path:   "/posts/1",
			Method: http.MethodPut,
			Status: http.StatusOK,
			Matches: map[string]string{
				"0": "/posts/1",
				"1": "1",
			},
			Body: url.Values{
				"body": {"updated post body"},
			},
		},
		{
			Path:   "/posts/1",
			Method: http.MethodDelete,
			Status: http.StatusOK,
			Matches: map[string]string{
				"0": "/posts/1",
				"1": "1",
			},
		},
		{
			Path:   "/posts/23/comments",
			Method: http.MethodGet,
			Status: http.StatusOK,
			Matches: map[string]string{
				"0": "/posts/23/comments",
				"1": "23",
			},
		},
		{
			Path:   "/posts/23/foo",
			Method: http.MethodGet,
			Status: http.StatusNotFound,
		},
	}

	rh := NewRouterMux()
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
					for k, v := range data.Matches {
						mv, ok := tc.Matches[k]
						if !ok {
							t.Errorf("handler match %s param not found", k)
						}
						if mv != v {
							t.Errorf("handler match %s %v, expected %v", k, v, mv)
						}
					}
				}
			}
		})
	}
}

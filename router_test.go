package conductor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

package conductor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testHandler struct {
	msg string
}

func (h testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, h.msg)
}

func TestUse(t *testing.T) {
	expected := "testrequest"

	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expected)
	}

	h := testHandler{expected}

	mod := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "modified-")
			h.ServeHTTP(w, r)
		})
	}

	c := New()
	c.Use(mod)

	mux := http.NewServeMux()
	mux.Handle("/testh", c.Handler(h))
	mux.Handle("/testfn", c.HandlerFunc(fn))

	server := httptest.NewServer(mux)
	defer server.Close()

	testcases := []string{
		"/testh",
		"/testfn",
	}

	for _, tc := range testcases {
		res, err := http.Get(server.URL + tc)
		if err != nil {
			t.Fatal(err)
		}

		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()

		if err != nil {
			t.Fatal(tc, err)
		}

		expectedMessage := fmt.Sprintf("modified-%s", expected)
		if string(body) != expectedMessage {
			t.Errorf("%s: %v != %v", tc, string(body), expectedMessage)
		}
	}
}

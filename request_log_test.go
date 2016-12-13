package conductor

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

var (
	validLogStart    = regexp.MustCompile(`(\d{4}[/]\d{2}[/]\d{2}\s\d{2}:\d{2}:\d{2})\sStarted GET \/`)
	validLogComplete = regexp.MustCompile(`(\d{4}[/]\d{2}[/]\d{2}\s\d{2}:\d{2}:\d{2})\sCompleted 200 OK`)
)

func TestLogHandler(t *testing.T) {
	expected := "Hello client"

	// Capture logging
	var b bytes.Buffer
	logger := log.New(&b, "", log.LstdFlags)

	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, expected)
	}

	h := RequestLogHandler(http.HandlerFunc(fn), logger)

	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("%d != %d", w.Code, http.StatusOK)
	}

	body := strings.TrimSpace(w.Body.String())
	if body != expected {
		t.Errorf("%s != %s", body, expected)
	}

	messages := strings.FieldsFunc(b.String(), func(c rune) bool {
		return c == '\n'
	})

	if len(messages) != 2 {
		t.Fatalf("Expected %d messages but got %d: %q", 2, len(messages), messages)
	}

	if !validLogStart.MatchString(messages[0]) {
		t.Errorf("%v != %s", messages[0], "(time) Started GET / from 127.0.0.1")
	}

	if !validLogComplete.MatchString(messages[1]) {
		t.Errorf("%v does not match %s", messages[1], "(time) Completed 200 OK")
	}
}

func TestLogApp(t *testing.T) {
	expected := "testrequest"

	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", expected)
	}

	// Capture logging
	var b bytes.Buffer
	logger := log.New(&b, "", log.LstdFlags)

	app := NewApp()
	app.Use(RequestLogComponent(logger))
	app.HandleFunc("/", fn)

	server := httptest.NewServer(app)
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

	results := string(body)
	if results != expected {
		t.Errorf("%v != %v", results[1], expected)
	}

	messages := strings.FieldsFunc(b.String(), func(c rune) bool {
		return c == '\n'
	})

	if !validLogStart.MatchString(messages[0]) {
		t.Errorf("%v != %s", messages[0], "(time) Started GET / from 127.0.0.1")
	}

	if !validLogComplete.MatchString(messages[1]) {
		t.Errorf("%v does not match %s", messages[1], "(time) Completed 200 OK")
	}
}

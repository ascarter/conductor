package conductor

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type testData struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
}

func TestReadJSON(t *testing.T) {
	v := testData{1, "foo"}

	handler := func(w http.ResponseWriter, r *http.Request) {
		var data testData
		if err := ReadJSON(r, &data); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(v, data) {
			t.Errorf("%v != %v", v, data)
		}
	}

	inBody, err := json.MarshalIndent(&v, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(inBody)
	req, err := http.NewRequest("GET", "/", buf)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler(w, req)
}

func TestWriteJSON(t *testing.T) {
	data := testData{1, "foo"}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if err := WriteJSON(w, data); err != nil {
			t.Fatal(err)
		}
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var v testData
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, data) {
		t.Errorf("%v != %v", v, data)
	}
}

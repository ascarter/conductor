package conductor

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// ReadJSON reads data from request body to the interface provided.
func ReadJSON(r *http.Request, data interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, data); err != nil {
		return err
	}

	return nil
}

// WriteJSON writes data as JSON to the output writer.
// Data expected to be able to be marshaled to JSON.
func WriteJSON(w http.ResponseWriter, data interface{}) error {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
	return nil
}

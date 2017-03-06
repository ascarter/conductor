package conductor

import (
	"encoding/json"
	"net/http"
)

// ReadJSON reads data from request body to the interface provided.
func ReadJSON(r *http.Request, data interface{}) error {
	body := make([]byte, r.ContentLength)

	if _, err := r.Body.Read(body); err != nil {
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

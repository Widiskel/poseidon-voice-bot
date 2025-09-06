package model

import "encoding/json"

type ApiResponse struct {
	StatusCode int
	Data       map[string]any
}

func (r *ApiResponse) Decode(v interface{}) error {
	b, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

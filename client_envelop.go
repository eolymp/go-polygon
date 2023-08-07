package polygon

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Envelop struct {
	Status  string           `json:"status"`
	Comment string           `json:"comment"`
	Result  *json.RawMessage `json:"result"`
}

func (e *Envelop) Unmarshal(v any) error {
	if e.Status != "OK" {
		return fmt.Errorf("API request failed: %v", e.Comment)
	}

	if e.Result == nil {
		return errors.New("result is not populated")
	}

	return json.Unmarshal(*e.Result, v)
}

package testtemplate

import (
	"encoding/json"
	"testing"
)

// DeepCopy copies the content of a template resource, and preserves the resource name
// and id.
func DeepCopy[T Data](t *testing.T, value T) T {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("could not encode to json: %s", err)
	}

	var result T

	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("could not decode from json: %s", err)
	}

	result.SetRName(value.RName())
	result.SetRInt(value.RInt())

	return result
}

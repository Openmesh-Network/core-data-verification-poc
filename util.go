package main

import (
	"encoding/json"
)

// saves me checking the error

func jsonOrEmptyArray(v any) []byte {
	j, err := json.Marshal(v)

	if err != nil {
		j = []byte("[]")
	}

	return j
}

func jsonOrEmptyObject(v any) []byte {
	j, err := json.Marshal(v)

	if err != nil {
		j = []byte("{}")
	}

	return j
}

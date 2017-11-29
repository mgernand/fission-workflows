package fission

import (
	"net/http"

	"errors"
	"io/ioutil"

	"encoding/json"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

func ParseRequest(r *http.Request, target map[string]*types.TypedValue) error {
	contentType := r.Header.Get("Content-Type")
	log.WithField("url", r.URL).WithField("content-type", contentType).Info("Request content-type")
	// Map Inputs to function parameters
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}

	var i interface{} = body
	if len(body) > 0 {
		err = json.Unmarshal(body, &i)
		if err != nil {
			log.WithField("body", len(body)).Debugf("Input is not json: %v", err)
			i = body
		}
	}

	parsedInput, err := typedvalues.Parse(i)
	if err != nil {
		return errors.New("failed to parse body")
	}

	log.WithField(types.INPUT_MAIN, parsedInput).Info("Parsed body")
	target[types.INPUT_MAIN] = parsedInput
	return nil
}

package upstream

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrRequestError    = errors.New("request error")
	ErrProviderError   = errors.New("version provider error")
	ErrVersionNotFound = errors.New("upstream version not found")
)

// httpGetJSON sends HTTP GET request to the given URL and writes JSON response to target struct.
// Returns error if the request of JSON decoding fails.
func httpGetJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("%w: GET %s %s", ErrRequestError, url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("%w: GET %s status %d", ErrProviderError, url, resp.StatusCode)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("%w: GET %s status %d", ErrVersionNotFound, url, resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(&target)
}

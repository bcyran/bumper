package version

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

func httpGetJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrRequestError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("%w: status %d", ErrProviderError, resp.StatusCode)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("%w: status %d", ErrVersionNotFound, resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(&target)
}

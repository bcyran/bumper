package testutils

import (
	"net/http"

	"github.com/h2non/gock"
)

func NoHeaderMatcher(header string) gock.MatchFunc {
	return func(request *http.Request, _requestMock *gock.Request) (bool, error) {
		return request.Header.Get(header) == "", nil
	}
}

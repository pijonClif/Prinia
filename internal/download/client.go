package download

import (
	"fmt"
	"net/http"
	"strconv"
)

// returns new http req
func NewRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, fmt.Errorf("\ncreate %s request: %w", method, err)
	}

	req.Header.Set("User-Agent", userAgent)
	return req, nil
}

// head req >> fetch ocntent length
func FileSize(url string) (int, error) {

	req, err := NewRequest(http.MethodHead, url)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("HEAD request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return 0, fmt.Errorf("HEAD request returned status %d", resp.StatusCode)
	}

	cl := resp.Header.Get("Content-Length")
	if cl == "" {
		return 0, fmt.Errorf("server did not return Content-Length (chunked or unknown size)")
	}

	size, err := strconv.Atoi(cl)
	if err != nil {
		return 0, fmt.Errorf("parse Content-Length: %w", err)
	}

	return size, nil
}

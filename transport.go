package supervisord

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
)

type Replace func(body []byte) []byte

// ReplaceBodyTransport use replace response body
// eg:replace illegal character code like "U+001B"
type ReplaceBodyTransport struct {
	Transport http.RoundTripper
	Replace   func(body []byte) []byte
}

func (t *ReplaceBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	tr := t.Transport
	if tr == nil {
		tr = http.DefaultTransport
	}
	res, err := tr.RoundTrip(req)
	if err != nil || res.StatusCode != http.StatusOK {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	res.Body = io.NopCloser(bytes.NewReader(t.Replace(body)))
	return res, err
}

// ReplaceXmlNotSupportedCharCode
// replace illegal character code
// If the data contains illegal character codes like U+001B, the unmarshalling process will fail, so they should be replaced.
// https://www.w3.org/TR/xml11/#charsets
func ReplaceXmlNotSupportedCharCode(body []byte) []byte {
	compile, _ := regexp.Compile(`[^\x09\x0A\x0D\x{0020}-\x{D7FF}\x{E000}-\x{FFFD}]`)
	body = compile.ReplaceAll(body, []byte("\uFFFD"))
	return body
}

package supervisord

import (
	"errors"

	"github.com/kolo/xmlrpc"
)

type (
	Client struct {
		*xmlrpc.Client
	}
)

var (
	FIXMENotImplementedError error
	ReturnedFalseError       error
	ReturnedMalformedReply   error
)

func init() {
	FIXMENotImplementedError = errors.New("Not implemented yet")
	ReturnedFalseError = errors.New("Call returned false")
	ReturnedMalformedReply = errors.New("Call returned malformed reply")
}

func (c *Client) stringCall(method string, args ...interface{}) (string, error) {
	var str string
	err := c.Call(method, args, &str)

	return str, err
}

func (c *Client) boolCall(method string, args ...interface{}) error {
	var result bool
	err := c.Call(method, args, &result)
	if err != nil {
		return err
	}

	if !result {
		return ReturnedFalseError
	}

	return nil
}

// Get a new client suitable for communicating with a supervisord.
// url must contain a real url to a supervisord RPC-service.
//
// Url for local supervisord will be http://127.0.0.1:9001/RPC2
func NewClient(url string) (*Client, error) {
	rpc, err := xmlrpc.NewClient(url, nil)
	if err != nil {
		return nil, err
	}

	return &Client{rpc}, nil
}

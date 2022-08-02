package supervisord

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/kolo/xmlrpc"
)

var DefaultRequestTimeoutSec = 30

type (
	Client struct {
		rpcUrl            string
		cl                *http.Client
		debug             bool
		requestTimeoutSec int
	}
)

var (
	// Will be returned by a few functions not yet fully implemented.
	FIXMENotImplementedError error

	// Will be returned if the API endpoint returns false without further explanation.
	ReturnedFalseError error

	// Will be returned if the API endpoint returns a reply we don't know how to parse.
	ReturnedMalformedReply error
)

func init() {
	FIXMENotImplementedError = errors.New("Not implemented yet")
	ReturnedFalseError = errors.New("Call returned false")
	ReturnedMalformedReply = errors.New("Call returned malformed reply")
}

type options struct {
	username, password string

	debug      bool
	timeoutSec int
}

func (me *options) setDefaults() {
	if me.timeoutSec == 0 {
		me.timeoutSec = DefaultRequestTimeoutSec
	}
}

// ClientOption is used to customize the client.
type ClientOption func(*options)

// WithAuthentication sets the username and password to use when authenticating against the server.
func WithAuthentication(username, password string) ClientOption {
	return func(o *options) {
		o.username = username
		o.password = password
	}
}

func WithRequestTimeout(sec int) ClientOption {
	return func(o *options) {
		o.timeoutSec = sec
	}
}
func WithDebug(debug bool) ClientOption {
	return func(o *options) {
		o.debug = debug
	}
}

func (c *Client) logf(s string, args ...interface{}) {
	if c.debug == false {
		return
	}
	fmt.Printf(s+"\n", args...)
}

func (c *Client) stringCall(method string, args ...interface{}) (string, error) {
	var str string
	c.logf("stringcall method:%s args:%q", method, args)
	err := c.Call(method, args, &str)

	return str, err
}

func (c *Client) boolCall(method string, args ...interface{}) error {
	var result bool
	c.logf("boolcall method:%s args:%q", method, args)
	err := c.Call(method, args, &result)
	if err != nil {
		return err
	}

	if !result {
		return ReturnedFalseError
	}

	return nil
}

func (c *Client) Call(method string, args interface{}, reply interface{}) error {

	// encode request
	largs := args.([]interface{})
	buf, err := xmlrpc.EncodeMethodCall(method, largs...)
	if err != nil {
		return err
	}
	c.logf("xmlrpc call method:%s timeout:%d args:%q\n", method, c.requestTimeoutSec, buf)

	reqTimeout := time.Duration(c.requestTimeoutSec) * time.Second

	ctx := context.Background()
	ctx2, cancel := context.WithTimeout(ctx, reqTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx2, "POST", c.rpcUrl, bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml")

	c.logf("xmlrpc do request method:%s url:%s", method, c.rpcUrl)
	resp, err := c.cl.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	buf2, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// decode
	c.logf("xmlrpc decode response %q", buf2)
	xres := xmlrpc.Response(buf2)

	if xres.Err() != nil {
		c.logf("ERROR: %s", xres.Err())
		return xres.Err()
	}

	err = xres.Unmarshal(reply)
	if err != nil {
		return err
	}

	return nil
}

// Get a new client suitable for communicating with a supervisord.
// url must contain a real url to a supervisord RPC-service.
//
// Url for local supervisord should be http://127.0.0.1:9001/RPC2 by default.
func NewClient(url string, opts ...ClientOption) (*Client, error) {
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}

	var tr http.RoundTripper
	tr = &http.Transport{}
	if opt.username != "" && opt.password != "" {
		tr = &basicAuthTransport{
			username: opt.username,
			password: opt.password,
			rt:       tr,
		}
	}

	cl := &http.Client{}
	cl.Transport = tr

	opt.setDefaults()

	me := &Client{
		cl:                cl,
		rpcUrl:            url,
		requestTimeoutSec: opt.timeoutSec,
		debug:             opt.debug,
	}

	return me, nil
}

// NewUnixSocketClient returns a new client which connects to supervisord
// though a local unix socket
func NewUnixSocketClient(path string, opts ...ClientOption) (*Client, error) {
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}

	// we inject this fake dialer, it will only connect
	// to the path given, and does not care about what address
	// is given to it.
	dialer := func(_, _ string) (conn net.Conn, err error) {
		return net.Dial("unix", path)
	}

	var tr http.RoundTripper = &http.Transport{
		Dial: dialer,
	}

	if opt.username != "" && opt.password != "" {
		tr = &basicAuthTransport{
			username: opt.username,
			password: opt.password,
			rt:       tr,
		}
	}
	opt.setDefaults()

	// we pass a valid url, as this is later url.Parse()'ed
	// also we need to somehow specify "/RPC2"
	cl := &http.Client{}
	cl.Transport = tr

	rpcUrl := "http://127.0.0.1:9001/RPC2"
	me := &Client{
		cl:                cl,
		rpcUrl:            rpcUrl,
		requestTimeoutSec: opt.timeoutSec,
		debug:             opt.debug,
	}
	return me, nil

}

// basicAuthTransport is an http.RoundTripper that wraps another http.RoundTripper
// and injects basic auth credentials into each request.
type basicAuthTransport struct {
	rt       http.RoundTripper
	username string
	password string
}

func (b basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(b.username, b.password)
	return b.rt.RoundTrip(req)
}

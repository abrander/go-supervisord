package supervisord

import ()

type (
	LogSegment struct {
		Payload  string `xmlrpc:"string"`
		Offset   int    `xmlrpc:"offset"`
		Overflow bool   `xmlrpc:"overflow"`
	}
)

func (c *Client) ReadProcessStdoutLog(name string, offset int, length int) (string, error) {
	return c.stringCall("supervisor.readProcessStdoutLog", name, offset, length)
}

func (c *Client) ReadProcessStderrLog(name string, offset int, length int) (string, error) {
	return c.stringCall("supervisor.readProcessStderrLog", name, offset, length)
}

func (c *Client) TailProcessStdoutLog(name string, offset int, length int) ([]LogSegment, error) {
	return nil, FIXMENotImplementedError
}

func (c *Client) TailProcessStderrLog(name string, offset int, length int) ([]LogSegment, error) {
	return nil, FIXMENotImplementedError
}

func (c *Client) ClearProcessLogs(name string) error {
	return c.boolCall("supervisor.clearProcessLogs", name)
}

func (c *Client) ClearAllProcessLogs() error {
	return c.boolCall("supervisor.clearAllProcessLogs")
}

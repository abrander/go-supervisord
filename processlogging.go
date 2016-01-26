package supervisord

type (
	// A LogSegment represents a "tail" of a log
	LogSegment struct {
		Payload  string `xmlrpc:"string"`
		Offset   int    `xmlrpc:"offset"`
		Overflow bool   `xmlrpc:"overflow"`
	}
)

// Read length bytes from name’s stdout log starting at offset.
func (c *Client) ReadProcessStdoutLog(name string, offset int, length int) (string, error) {
	return c.stringCall("supervisor.readProcessStdoutLog", name, offset, length)
}

// Read length bytes from name’s stderr log starting at offset.
func (c *Client) ReadProcessStderrLog(name string, offset int, length int) (string, error) {
	return c.stringCall("supervisor.readProcessStderrLog", name, offset, length)
}

// This is not implemented yet.
func (c *Client) TailProcessStdoutLog(name string, offset int, length int) ([]LogSegment, error) {
	return nil, FIXMENotImplementedError
}

// This is not implemented yet.
func (c *Client) TailProcessStderrLog(name string, offset int, length int) ([]LogSegment, error) {
	return nil, FIXMENotImplementedError
}

// Clear the stdout and stderr logs for the process name and reopen them.
func (c *Client) ClearProcessLogs(name string) error {
	return c.boolCall("supervisor.clearProcessLogs", name)
}

// Clear all process log files.
func (c *Client) ClearAllProcessLogs() error {
	return c.boolCall("supervisor.clearAllProcessLogs")
}

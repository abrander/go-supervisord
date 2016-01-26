package supervisord

type (
	// A numeric representation of Supervisord's internal state.
	StateCode int

	// A textual representation of Supervisord's internal state.
	StateName string

	// Represents Supervisord's internal state.
	State struct {
		Code StateCode `xmlrpc:"statecode"`
		Name StateName `xmlrpc:"statename"`
	}
)

const (
	StateCodeFatal      StateCode = 2  // Supervisor has experienced a serious error.
	StateCodeRunning    StateCode = 1  // Supervisor is working normally.
	StateCodeRestarting StateCode = 0  // Supervisor is in the process of restarting.
	StateCodeShutdown   StateCode = -1 // Supervisor is in the process of shutting down.

	StateNameFatal      StateName = "FATAL"      // Supervisor has experienced a serious error.
	StateNameRunning    StateName = "RUNNING"    // Supervisor is working normally.
	StateNameRestarting StateName = "RESTARTING" // Supervisor is in the process of restarting.
	StateNameShutdown   StateName = "SHUTDOWN"   // Supervisor is in the process of shutting down.
)

// Return the version of the RPC API used by supervisord.
//
// This API is versioned separately from Supervisor itself. The API
// version returned by getAPIVersion only changes when the API changes.
// Its purpose is to help the client identify with which version of the
// Supervisor API it is communicating.
//
// When writing software that communicates with this API, it is highly
// recommended that you first test the API version for compatibility
// before making method calls.
func (c *Client) GetAPIVersion() (string, error) {
	return c.stringCall("supervisor.getAPIVersion")
}

// Return the version of the supervisor package in use by supervisord.
func (c *Client) GetSupervisorVersion() (string, error) {
	return c.stringCall("supervisor.getSupervisorVersion")
}

// Return identifiying string of the supervisord instance.
//
// This method allows the client to identify with which Supervisor
// instance it is communicating in the case of environments where multiple
// Supervisors may be running.
//
// The identification is a string that must be set in Supervisor’s
// configuration file. This method simply returns that value back to the
// client.
func (c *Client) GetIdentification() (string, error) {
	return c.stringCall("supervisor.getIdentification")
}

// Return current state of supervisord as a struct.
//
// This is an internal value maintained by Supervisor that determines what
// Supervisor believes to be its current operational state.
//
// Some method calls can alter the current state of the Supervisor. For
// example, calling the Shutdown() while the station is in the StateCodeRunning
// state places the Supervisor in the StateCodeShutdown state while it is
// shutting down.
func (c *Client) GetState() (State, error) {
	var state State
	err := c.Call("supervisor.getState", nil, &state)

	return state, err
}

// Return the PID of supervisord.
func (c *Client) GetPID() (int, error) {
	var pid int
	err := c.Call("supervisor.getPID", nil, &pid)

	return pid, err
}

// Read length bytes from the main log starting at offset.
//
// It can either return the entire log, a number of characters from the tail
// of the log, or a part of the log specified by the offset and length
// parameters.
func (c *Client) ReadLog(offset int, length int) (string, error) {
	var log string
	err := c.Call("supervisor.readLog", []interface{}{offset, length}, &log)

	return log, err
}

// Clear the main log.
func (c *Client) ClearLog() error {
	return c.boolCall("supervisor.clearLog")
}

// Shut down the supervisor process.
//
// This method shuts down the Supervisor daemon. If any processes are
// running, they are automatically killed without warning.
//
// Unlike most other methods, if Supervisor is in the StateCodeFatal state,
// this method will still function.
func (c *Client) Shutdown() error {
	return c.boolCall("supervisor.shutdown")
}

// Restart the supervisor process.
//
// This method soft restarts the Supervisor daemon. If any processes are
// running, they are automatically killed without warning. Note that the
// actual UNIX process for Supervisor cannot restart; only Supervisor’s
// main program loop. This has the effect of resetting the internal states
// of Supervisor.
//
// Unlike most other methods, if Supervisor is in the StateCodeFatal state,
// this method will still function.
func (c *Client) Restart() error {
	return c.boolCall("supervisor.restart")
}

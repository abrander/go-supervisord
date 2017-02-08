package supervisord

import "syscall"

type (
	// A structure containing data about a process.
	ProcessInfo struct {
		Name          string       `xmlrpc:"name"`           // Name of the process
		Group         string       `xmlrpc:"group"`          // Name of the process’ group
		Start         int          `xmlrpc:"start"`          // UNIX timestamp of when the process was started
		Stop          int          `xmlrpc:"stop"`           // UNIX timestamp of when the process last ended, or 0 if the process has never been stopped
		Now           int          `xmlrpc:"now"`            // UNIX timestamp of the current time, which can be used to calculate process up-time.
		State         ProcessState `xmlrpc:"state"`          // State code, see ProcessState.
		StateName     string       `xmlrpc:"statename"`      // String description of state
		SpawnErr      string       `xmlrpc:"spawnerr"`       // Description of error that occurred during spawn, or empty string if none
		ExitStatus    int          `xmlrpc:"exitstatus"`     // Exit status (errorlevel) of process, or 0 if the process is still running
		StdoutLogfile string       `xmlrpc:"stdout_logfile"` // Absolute path and filename to the STDOUT logfile
		StderrLogfile string       `xmlrpc:"stderr_logfile"` // Absolute path and filename to the STDOUT logfile
		Pid           int          `xmlrpc:"pid"`            // UNIX process ID (PID) of the process, or 0 if the process is not running
	}

	// A process controlled by supervisord will be in one of the below
	// states at any given time. You may see these state names in various
	// user interface elements in clients.
	ProcessState int
)

const (
	StateStopped  ProcessState = 0    // The process has been stopped due to a stop request or has never been started
	StateStarting ProcessState = 10   // The process is starting due to a start request
	StateRunning  ProcessState = 20   // The process is running
	StateBackoff  ProcessState = 30   // The process entered the StateStarting state but subsequently exited too quickly to move to the StateRunning state
	StateStopping ProcessState = 40   // The process is stopping due to a stop request
	StateExited   ProcessState = 100  // The process exited from the StateRunning state (expectedly or unexpectedly)
	StateFatal    ProcessState = 200  // The process could not be started successfully
	StateUnknown  ProcessState = 1000 // The process is in an unknown state (supervisord programming error)
)

func (c *Client) processInfoArrayCall(method string, args ...interface{}) ([]ProcessInfo, error) {
	var processinfo []ProcessInfo
	err := c.Call(method, args, &processinfo)
	if err != nil {
		return nil, err
	}

	return processinfo, nil
}

// Get info about a process named name.
func (c *Client) GetProcessInfo(name string) (*ProcessInfo, error) {
	var processinfo ProcessInfo
	err := c.Call("supervisor.getProcessInfo", name, &processinfo)
	if err != nil {
		return nil, err
	}

	return &processinfo, nil
}

// Get info about all processes.
func (c *Client) GetAllProcessInfo() ([]ProcessInfo, error) {
	return c.processInfoArrayCall("supervisor.getAllProcessInfo")
}

// SignalProcess sends a signal to a process
// name: Name of the process to signal (or ‘group:name’)
// signal: Signal to send
// Requires supervisord >= 3.2.0
// http://supervisord.org/changes.html#id6
func (c *Client) SignalProcess(name string, signal syscall.Signal) error {
	return c.boolCall("supervisor.signalProcess", name, int(signal))
}

// SignalAllProcesses sends a signal to all processes in the process list
// Requires supervisord >= 3.2.0
// http://supervisord.org/changes.html#id6
func (c *Client) SignalAllProcesses(signal syscall.Signal) ([]ProcessInfo, error) {
	return c.processInfoArrayCall("supervisor.signalAllProcesses", int(signal))
}

// Start a process.
func (c *Client) StartProcess(name string, wait bool) error {
	return c.boolCall("supervisor.startProcess", name, wait)
}

// Start all processes listed in the configuration file.
//
// Set wait to true if the call should wait for completion before returning.
func (c *Client) StartAllProcesses(wait bool) ([]ProcessInfo, error) {
	return c.processInfoArrayCall("supervisor.startAllProcesses", wait)
}

// Start all processes in the group named name.
//
// Set wait to true if the call should wait for completion before returning.
func (c *Client) StartProcessGroup(name string, wait bool) ([]ProcessInfo, error) {
	return c.processInfoArrayCall("supervisor.startProcessGroup", name, wait)
}

// Stop a process named by name.
//
// Set wait to true if the call should wait for completion before returning.
func (c *Client) StopProcess(name string, wait bool) error {
	return c.boolCall("supervisor.stopProcess", name, wait)
}

// Stop all processes in the process group named name.
//
// Set wait to true if the call should wait for completion before returning.
func (c *Client) StopProcessGroup(name string, wait bool) ([]ProcessInfo, error) {
	return c.processInfoArrayCall("supervisor.stopProcessGroup", name, wait)
}

// Stop all processes in the process list.
//
// Set wait to true if the call should wait for completion before returning.
func (c *Client) StopAllProcesses(wait bool) ([]ProcessInfo, error) {
	return c.processInfoArrayCall("supervisor.stopAllProcesses", wait)
}

// Send a string to the stdin of the process name. If the process’s
// stdin cannot accept input (e.g. it was closed by the child process),
// return non-nil error.
func (c *Client) SendProcessStdin(name string, chars string) error {
	return c.boolCall("supervisor.sendProcessStdin", name, chars)
}

// This is not implemented yet.
func (c *Client) SendRemoteCommEvent(typ, data interface{}) error {
	return FIXMENotImplementedError
}

// Reload supervisord configuration.
//
// This will not change, start or stop any running processes. It will only
// read new configuration. See Update() for an all-in-one solution.
func (c *Client) ReloadConfig() ([]string, []string, []string, error) {
	result := make([][][]string, 0)

	err := c.Call("supervisor.reloadConfig", nil, &result)
	if err != nil {
		return nil, nil, nil, err
	}

	if len(result) != 1 {
		return nil, nil, nil, ReturnedMalformedReply
	}

	if len(result[0]) != 3 {
		return nil, nil, nil, ReturnedMalformedReply
	}

	return result[0][0], result[0][1], result[0][2], err
}

// This will reload configuration and adapt running processes to the new
// configuration. Changed program groups will be restarted.
// Should behave like "supervisorctl update".
func (c *Client) Update() error {
	added, changed, removed, err := c.ReloadConfig()
	if err != nil {
		return err
	}

	toStart := append(added, changed...)
	toStop := append(changed, removed...)

	for _, name := range toStop {
		_, err = c.StopProcessGroup(name, true)
		if err != nil {
			return err
		}

		err = c.RemoveProcessGroup(name)
		if err != nil {
			return err
		}
	}

	for _, name := range toStart {
		err = c.AddProcessGroup(name)
		if err != nil {
			return err
		}
	}

	return nil
}

// Update the config for a running process from config file.
func (c *Client) AddProcessGroup(name string) error {
	return c.boolCall("supervisor.addProcessGroup", name)
}

// Remove a stopped process from the active configuration.
func (c *Client) RemoveProcessGroup(name string) error {
	return c.boolCall("supervisor.removeProcessGroup", name)
}

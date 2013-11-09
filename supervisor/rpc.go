package supervisor

import (
	"errors"
	"fmt"
	"github.com/kolo/xmlrpc"
)

const (
	apiVersion string = "3.0"
)

func makeParams(params ...interface{}) xmlrpc.Params {
	return xmlrpc.Params{params}
}

type SupervisorState struct {
	StateCode int64
	StateName string
}

func newSupervisorState(result xmlrpc.Struct) *SupervisorState {
	state := new(SupervisorState)
	state.StateCode = result["statecode"].(int64)
	state.StateName = result["statename"].(string)
	return state
}

func (state SupervisorState) String() string {
	return fmt.Sprintf(`SupervisorState{%d, "%s"}`, state.StateCode, state.StateName)
}

type ProcessInfo struct {
	Name          string
	Description   string
	Group         string
	Start         int64
	Stop          int64
	Now           int64
	State         int64
	StateName     string
	SpawnErr      string
	ExitStatus    int64
	Logfile       string
	StdoutLogfile string
	StderrLogfile string
	PID           int64
}

func newProcessInfo(result xmlrpc.Struct) *ProcessInfo {
	info := new(ProcessInfo)
	info.Name = result["name"].(string)
	info.Description = result["description"].(string)
	info.Group = result["group"].(string)
	info.Start = result["start"].(int64)
	info.Stop = result["stop"].(int64)
	info.Now = result["now"].(int64)
	info.State = result["state"].(int64)
	info.StateName = result["statename"].(string)
	info.SpawnErr = result["spawnerr"].(string)
	info.ExitStatus = result["exitstatus"].(int64)
	info.Logfile = result["logfile"].(string)
	info.StdoutLogfile = result["stdout_logfile"].(string)
	info.StderrLogfile = result["stderr_logfile"].(string)
	info.PID = result["pid"].(int64)
	return info
}

func (info ProcessInfo) String() string {
	return fmt.Sprintf(`ProcessInfo{"%s", %d, "%s"}`, info.Name, info.PID, info.StateName)
}

type ProcessStatus struct {
	Name        string
	Description string
	Group       string
	Status      int64
}

func newProcessStatus(result xmlrpc.Struct) *ProcessStatus {
	status := new(ProcessStatus)
	status.Name = result["name"].(string)
	status.Description = result["description"].(string)
	status.Group = result["group"].(string)
	status.Status = result["status"].(int64)
	return status
}

func (status ProcessStatus) String() string {
	return fmt.Sprintf(`ProcessStatus("%s", %d)`, status.Name, status.Status)
}

type ProcessTail struct {
	Log      string
	Offset   int64
	Overflow bool
}

func newProcessTail(result []interface{}) *ProcessTail {
	tail := new(ProcessTail)
	tail.Log = result[0].(string)
	tail.Offset = result[1].(int64)
	tail.Overflow = result[2].(bool)
	return tail
}

func (tail ProcessTail) String() string {
	return tail.Log
}

type Client struct {
	RpcClient  *xmlrpc.Client
	ApiVersion string
}

// NewClient creates a new supervisor RPC client.
func NewClient(url string) (client Client, err error) {
	var rpc *xmlrpc.Client
	if rpc, err = xmlrpc.NewClient(url, nil); err != nil {
		return
	}

	version := ""
	if err = rpc.Call("supervisor.getAPIVersion", nil, &version); err != nil {
		return
	}
	if version != apiVersion {
		err = errors.New(fmt.Sprintf("want Supervisor API version %s, got %s instead", apiVersion, version))
		return
	}
	client = Client{rpc, version}
	return
}

// Close the client.
func (client Client) Close() error {
	return client.RpcClient.Close()
}

// GetSupervisorVersion returns the Supervisor version we connect to.
func (client Client) GetSupervisorVersion() (version string, err error) {
	err = client.RpcClient.Call("supervisor.getSupervisorVersion", nil, &version)
	return
}

// GetIdentification returns the Supervisor ID string.
func (client Client) GetIdentification() (id string, err error) {
	err = client.RpcClient.Call("supervisor.getIdentification", nil, &id)
	return
}

// GetState returns the Supervisor process state.
func (client Client) GetState() (state *SupervisorState, err error) {
	result := xmlrpc.Struct{}
	if err = client.RpcClient.Call("supervisor.getState", nil, &result); err == nil {
		state = newSupervisorState(result)
	}
	return
}

// GetPID returns the Supervisor process PID.
func (client Client) GetPID() (pid int64, err error) {
	err = client.RpcClient.Call("supervisor.getPID", nil, &pid)
	return
}

// ClearLog clears the Supervisor process log.
func (client Client) ClearLog() (result bool, err error) {
	err = client.RpcClient.Call("supervisor.clearLog", nil, &result)
	return
}

// Shutdown shuts down the Supervisor process.
func (client Client) Shutdown() (result bool, err error) {
	err = client.RpcClient.Call("supervisor.shutdown", nil, &result)
	return
}

// Restart restarts the Supervisor process.
func (client Client) Restart() (result bool, err error) {
	err = client.RpcClient.Call("supervisor.restart", nil, &result)
	return
}

// GetProcessInfo retrieves information for a particular Supervisor process.
func (client Client) GetProcessInfo(name string) (info *ProcessInfo, err error) {
	result := xmlrpc.Struct{}
	if err = client.RpcClient.Call("supervisor.getProcessInfo", name, &result); err == nil {
		info = newProcessInfo(result)
	}
	return
}

// GetAllProcessInfo retrieves information for all Supervisor processes.
func (client Client) GetAllProcessInfo() (info []*ProcessInfo, err error) {
	var results []interface{}
	if err = client.RpcClient.Call("supervisor.getAllProcessInfo", nil, &results); err == nil {
		info = make([]*ProcessInfo, len(results))
		for i, result := range results {
			info[i] = newProcessInfo(result.(xmlrpc.Struct))
		}
	}
	return
}

// StartProcess tells Supervisor to start the named process.
func (client Client) StartProcess(name string, wait bool) (result bool, err error) {
	params := makeParams(name, wait)
	err = client.RpcClient.Call("supervisor.startProcess", params, &result)
	return
}

// StopProcess tells Supervisor to stop the named process.
func (client Client) StopProcess(name string, wait bool) (result bool, err error) {
	params := makeParams(name, wait)
	err = client.RpcClient.Call("supervisor.stopProcess", params, &result)
	return
}

// StartAllProcesses tells Supervisor to start all stopped processes.
func (client Client) StartAllProcesses(wait bool) (info []*ProcessStatus, err error) {
	var results []interface{}
	if err = client.RpcClient.Call("supervisor.startAllProcesses", wait, &results); err == nil {
		info = make([]*ProcessStatus, len(results))
		for i, result := range results {
			info[i] = newProcessStatus(result.(xmlrpc.Struct))
		}
	}
	return
}

// StopAllProcesses teslls Supervisor to stop all running processes.
func (client Client) StopAllProcesses(wait bool) (info []*ProcessStatus, err error) {
	var results []interface{}
	if err = client.RpcClient.Call("supervisor.stopAllProcesses", wait, &results); err == nil {
		info = make([]*ProcessStatus, len(results))
		for i, result := range results {
			info[i] = newProcessStatus(result.(xmlrpc.Struct))
		}
	}
	return
}

// StartProcessGroup tells Supervisor to start all stopped processes in the named group.
func (client Client) StartProcessGroup(name string, wait bool) (result bool, err error) {
	params := makeParams(name, wait)
	err = client.RpcClient.Call("supervisor.startProcessGroup", params, &result)
	return
}

// StopProcessGroup tells Supervisor to start all stopped processes in the named group.
func (client Client) StopProcessGroup(name string, wait bool) (result bool, err error) {
	params := makeParams(name, wait)
	err = client.RpcClient.Call("supervisor.stopProcessGroup", params, &result)
	return
}

// SendProcessStdin send data to the stdin of a running process.
func (client Client) SendProcessStdin(name string, chars string) (result bool, err error) {
	params := makeParams(name, chars)
	err = client.RpcClient.Call("supervisor.sendProcessStdin", params, &result)
	return
}

// SendRemoteCommEvent sends an event to Supervisor processes listening to RemoveCommunicationEvents..
func (client Client) SendRemoteCommEvent(typeKey string, data string) (result bool, err error) {
	params := makeParams(typeKey, data)
	err = client.RpcClient.Call("supervisor.sendRemoteCommEvent", params, &result)
	return
}

// AddProcessGroup adds a configured process group to Supervisor.
func (client Client) AddProcessGroup(name string) (result bool, err error) {
	err = client.RpcClient.Call("supervisor.addProcessGroup", name, &result)
	return
}

// RemoveProcessGroup removes a configured process group from Supervisor.
func (client Client) RemoveProcessGroup(name string) (result bool, err error) {
	err = client.RpcClient.Call("supervisor.removeProcessGroup", name, &result)
	return
}

// ReadLog reads the Supervisor process log.
func (client Client) ReadLog(offset int64, length int64) (log string, err error) {
	params := makeParams(offset, length)
	err = client.RpcClient.Call("supervisor.readLog", params, &log)
	return
}

// ReadProcessStdoutLog reads the stdout log for the named process.
func (client Client) ReadProcessStdoutLog(name string, offset int64, length int64) (log string, err error) {
	params := makeParams(name, offset, length)
	err = client.RpcClient.Call("supervisor.readProcessStdoutLog", params, &log)
	return
}

// ReadProcessStderrLog reads the stderr log for the named process.
func (client Client) ReadProcessStderrLog(name string, offset int64, length int64) (log string, err error) {
	params := makeParams(name, offset, length)
	err = client.RpcClient.Call("supervisor.readProcessStderrLog", params, &log)
	return
}

// TailProcessStdoutLog reads the stdout log for the named process.
func (client Client) TailProcessStdoutLog(name string, offset int64, length int64) (tail *ProcessTail, err error) {
	params := makeParams(name, offset, length)
	result := make([]interface{}, 0, 3)
	if err = client.RpcClient.Call("supervisor.tailProcessStdoutLog", params, &result); err == nil {
		tail = newProcessTail(result)
	}
	return
}

// TailProcessStderrLog reads the stderr log for the named process.
func (client Client) TailProcessStderrLog(name string, offset int64, length int64) (tail *ProcessTail, err error) {
	params := makeParams(name, offset, length)
	result := make([]interface{}, 0, 3)
	if err = client.RpcClient.Call("supervisor.tailProcessStderrLog", params, &result); err == nil {
		tail = newProcessTail(result)
	}
	return
}

// ClearProcessLogs clears all logs for the named process.
func (client Client) ClearProcessLogs(name string) (result bool, err error) {
	err = client.RpcClient.Call("supervisor.clearProcessLogs", name, &result)
	return
}

// ClearAllProcessLogs clears all logs all processes.
func (client Client) ClearAllProcessLogs(name string) (result bool, err error) {
	err = client.RpcClient.Call("supervisor.clearAllProcessLogs", name, &result)
	return
}

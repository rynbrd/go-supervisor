package supervisor

import (
	"errors"
	"strconv"
)

type Process struct {
	Name  string
	Group string
	State string
	PID   int
}

// update the process from a listener event
func (proc *Process) updateFromListener(event Event) error {
	name, ok := event.Meta["processname"]
	if !ok {
		return errors.New("processname not found in metadata")
	}
	group, ok := event.Meta["groupname"]
	if !ok {
		return errors.New("groupname not found in metadata")
	}

	var pid int
	var str string
	var err error

	if str, ok = event.Meta["pid"]; ok {
		if pid, err = strconv.Atoi(str); err != nil {
			return err
		}
	}

	state := event.State()
	if state == Stopped {
		pid = 0
	}

	proc.Name = name
	proc.Group = group
	proc.State = state
	proc.PID = pid
	return nil
}

// update the process from an rpc response
func (proc *Process) updateFromRpc(info ProcessInfo) {
	proc.Name = info.Name
	proc.Group = info.Group
	proc.State = info.StateName
	proc.PID = int(info.PID)
}

// update the process from process data
func (proc *Process) update(data interface{}) error {
	switch data.(type) {
	case Event:
		return proc.updateFromListener(data.(Event))
	case ProcessInfo:
		proc.updateFromRpc(data.(ProcessInfo))
	default:
		return errors.New("invalid data type")
	}
	return nil
}

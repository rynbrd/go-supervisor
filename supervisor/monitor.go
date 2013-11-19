package supervisor

import (
	"errors"
	"io"
	"strconv"
)

const (
	Backoff  string = "BACKOFF"
	Exited   string = "EXITED"
	Fatal    string = "FATAL"
	Running  string = "RUNNING"
	Starting string = "STARTING"
	Stopped  string = "STOPPED"
	Stopping string = "STOPPING"
	Unknown  string = "UNKNOWN"
)

// Get name from process data.
func getProcessName(data interface{}) (name string, err error) {
	switch data.(type) {
	case Event:
		var ok bool
		if name, ok = (data.(Event)).Meta["processname"]; !ok {
			err = errors.New("processname not found in event metadata")
		}
	case ProcessInfo:
		name = (data.(ProcessInfo)).Name
	default:
		err = errors.New("invalid data type")
	}
	return
}

// SupervisorStateEvent is emitted when the Supervisor instance changes state.
type SupervisorStateEvent struct {
	Supervisor Supervisor
	FromName   string
	FromState  string
}

// ProcessAddEvent is emitted when a process is added to Supervisor.
type ProcessAddEvent struct {
	Supervisor Supervisor
	Process    Process
}

// ProcessRemoveEvent is emitted when a process is removed from Supervisor.
type ProcessRemoveEvent ProcessAddEvent

// ProcessStateEvent is emitted when a process changes state.
type ProcessStateEvent struct {
	Supervisor Supervisor
	Process    Process
	FromState  string
	Tries      int
}

type Monitor struct {
	Client     Client
	Listener   Listener
	Supervisor *Supervisor
	Processes  map[string]*Process
	events     chan interface{}
}

// NewMonitor creates a new Supervisor monitor.
func NewMonitor(url string, in io.Reader, out io.Writer, events chan interface{}) (mon Monitor, err error) {
	client, err := NewClient(url)
	if err != nil {
		return
	}

	listener := NewListener(in, out)

	mon = Monitor{
		client,
		listener,
		NewSupervisor(),
		make(map[string]*Process),
		events,
	}
	return
}

// Close the monitor.
func (mon Monitor) Close() error {
	return mon.Client.Close()
}

// Update supervisor struct with a new name and state.
func (mon Monitor) updateSupervisor(name string, state string) {
	if mon.events != nil && (mon.Supervisor.State != state || mon.Supervisor.Name != name) {
		oldName := mon.Supervisor.Name
		oldState := mon.Supervisor.State
		mon.Supervisor.Name = name
		mon.Supervisor.State = state
		mon.events <- SupervisorStateEvent{*mon.Supervisor, oldName, oldState}
	} else {
		mon.Supervisor.Name = name
		mon.Supervisor.State = state
	}
}

// Update a process with an event or info struct.
func (mon Monitor) updateProcess(data interface{}) error {
	name, err := getProcessName(data)
	if err != nil {
		return err
	}

	if proc, ok := mon.Processes[name]; ok {
		emit := false
		tries := 0
		fromState := ""

		switch data.(type) {
		case Event:
			event := data.(Event)
			fromState = event.Meta["from_state"]
			proc.updateFromListener(data.(Event))
			emit = true

			if val, ok := event.Meta["tries"]; ok {
				if tries, err = strconv.Atoi(val); err != nil {
					return err
				}
			}
		case ProcessInfo:
			fromState = proc.State
			fromPid := proc.PID
			proc.updateFromRpc(data.(ProcessInfo))
			emit = fromState != proc.State || fromPid != proc.PID
		default:
			return errors.New("invalid data type")
		}

		if emit && mon.events != nil {
			mon.events <- ProcessStateEvent{*mon.Supervisor, *proc, fromState, tries}
		}
	} else {
		proc := &Process{}
		if err := proc.update(data); err != nil {
			return err
		}

		mon.Processes[proc.Name] = proc

		if mon.events != nil {
			mon.events <- ProcessAddEvent{*mon.Supervisor, *proc}
		}
	}
	return nil
}

func (mon Monitor) removeProcess(proc *Process) {
	delete(mon.Processes, proc.Name)
	if mon.events != nil {
		mon.events <- ProcessRemoveEvent{*mon.Supervisor, *proc}
	}
}

// Refresh polls the Supervisor instance for the current state.
func (mon Monitor) Refresh() (err error) {
	name, err := mon.Client.GetIdentification()
	if err != nil {
		return
	}
	state, err := mon.Client.GetState()
	if err != nil {
		return
	}
	allInfo, err := mon.Client.GetAllProcessInfo()
	if err != nil {
		return
	}

	// update supervisor
	mon.updateSupervisor(name, state.StateName)

	// add or update processes
	allInfoMap := make(map[string]*ProcessInfo, len(allInfo))
	for _, info := range allInfo {
		allInfoMap[info.Name] = &info
		mon.updateProcess(info)
	}

	// remove processes
	for name, proc := range mon.Processes {
		if _, ok := allInfoMap[name]; !ok {
			mon.removeProcess(proc)
		}
	}
	return
}

// Run monitors the status of the Supervisor instance and sends events to the provided channel.
func (mon Monitor) Run() error {
	done := make(chan bool)
	events := make(chan Event)

	defer func() {
		close(events)
		<-done
	}()

	go func() {
		for event := range events {
			parent := event.Parent()
			switch parent {
			case "PROCESS_STATE":
				mon.updateProcess(event)
			case "SUPERVISOR_STATE_CHANGE":
				state := event.State()
				mon.updateSupervisor(mon.Supervisor.Name, state)
			case "TICK":
				mon.Refresh()
			}
		}
		done <- true
	}()

	return mon.Listener.Run(events)
}

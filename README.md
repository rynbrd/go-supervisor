Go Supervisor API Toolkit
=========================
API toolkit for Supervisor written in Go!

Event Listener
--------------
Listener implements a Supervisor event listener. The following code illustrates basic usage of the listener:

```
func main() {
	done := make(chan bool)
	events := make(chan supervisor.Event)
	evl := supervisor.NewListener(os.Stdin, os.Stdout)

	go func() {
		for event := range events {
			fmt.Fprintf(os.Stderr, "Got event: %s\n", event)
		}
		done <- true
	}()

	evl.Run(events)
	close(events)
	<-done
}
```

RPC Client
----------
Client implements an HTML RPC client to communicate with Supervisor. The following code demonstrates using the client to start a service:

```
func main() {
	client, err := supervisor.NewClient("http://localhost:9001/RPC2")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	if result, err = client.StartProcess("nginx", true); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	} else if !result {
		fmt.Printf("Failed to start process.\n")
	}

	fmt.Printf("Process started.\n")
}
```

Stateful Monitor
----------------
Monitor implements a stateful monitoring system. It maintains the current state of all processes and emits events when processes are added, removed, or change state. It will also emit events when the Supervisor instance changes state. It will also do a full state refresh on any TICK event it receives. For full functionality it requires the PROCESS_STATE, SUPERVISOR_STATE_CHANGE, and a TICK event. If the TICK event is removed then no events will be emitted for process removal.

As an example:

```
func main() {
	url := "http://localhost:9001/RPC2"
	events := make(chan interface{})
	mon, err := supervisor.NewMonitor(url, os.Stdin, os.Stdout, events)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	done := make(chan bool)
	go func() {
		for event := range events {
			switch event.(type) {
			case supervisor.ProcessAddEvent:
				process := (event.(supervisor.ProcessAddEvent)).Process
				fmt.Fprintf(os.Stderr, "Process %s added\n", process.Name)
			case supervisor.ProcessRemoveEvent:
				process := (event.(supervisor.ProcessRemoveEvent)).Process
				fmt.Fprintf(os.Stderr, "Process %s added\n", process.Name)
			case supervisor.ProcessStateEvent:
				process := (event.(supervisor.ProcessStateEvent)).Process
				from := (event.(supervisor.ProcessStateEvent)).FromState
				fmt.Fprintf(os.Stderr, "Process %s state change %s => %s\n", process.Name, from, process.State)
			case supervisor.SupervisorStateEvent:
				supervisor := (event.(supervisor.SupervisorStateEvent)).Supervisor
				from := (event.(supervisor.SupervisorStateEvent)).FromState
				fmt.Fprintf(os.Stderr, "Supervisor \"%s\" state change %s => %s\n", supervisor.Name, from, supervisor.State)
			default:
				fmt.Fprintf(os.Stderr, "Unchecked Event: %+v\n", event)
			}
		}
		done <- true
	}()

	mon.Refresh()
	mon.Run()

	close(events)
	mon.Close()
	<- done
}
```

License
-------
This software project is licensed under the BSD-derived license and is copyright (c) 2013 Ryan Bourgeois. A copy of the license is included in the LICENSE file. If it is missing a copy can be found on the project page.

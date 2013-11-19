package supervisor

type Supervisor struct {
	Name  string
	State string
}

func NewSupervisor() *Supervisor {
	return &Supervisor{"supervisor", Unknown}
}

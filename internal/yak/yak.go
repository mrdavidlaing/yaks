package yak

type State string

const (
	StateTodo State = "todo"
	StateDone State = "done"
)

func (s State) String() string {
	return string(s)
}

type Yak struct {
	Name        string
	State       State
	Children    []Yak
	ContextPath string
}

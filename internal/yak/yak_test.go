package yak

import "testing"

func TestState_TodoString(t *testing.T) {
	s := StateTodo
	if s.String() != "todo" {
		t.Errorf("expected 'todo', got '%s'", s.String())
	}
}

func TestState_DoneString(t *testing.T) {
	s := StateDone
	if s.String() != "done" {
		t.Errorf("expected 'done', got '%s'", s.String())
	}
}

func TestYak_HasNameAndState(t *testing.T) {
	y := Yak{Name: "my task", State: StateTodo}
	if y.Name != "my task" {
		t.Errorf("expected name 'my task', got '%s'", y.Name)
	}
	if y.State != StateTodo {
		t.Errorf("expected state todo, got '%s'", y.State)
	}
}

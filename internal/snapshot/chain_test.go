package snapshot

import (
	"errors"
	"testing"
)

type fakeCmd struct {
	name       string
	failOnExec bool
	failOnUndo bool
	executed   bool
	undone     bool
}

func (f *fakeCmd) Execute() error {
	if f.failOnExec {
		return errors.New("exec failed: " + f.name)
	}
	f.executed = true
	return nil
}

func (f *fakeCmd) Undo() error {
	if f.failOnUndo {
		return errors.New("undo failed: " + f.name)
	}
	f.undone = true
	return nil
}

func (f *fakeCmd) Describe() string { return f.name }

func TestChain_RunsAllOnSuccess(t *testing.T) {
	a, b, c := &fakeCmd{name: "a"}, &fakeCmd{name: "b"}, &fakeCmd{name: "c"}
	if err := NewChain(a, b, c).Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, cmd := range []*fakeCmd{a, b, c} {
		if !cmd.executed {
			t.Errorf("%s: expected executed", cmd.name)
		}
		if cmd.undone {
			t.Errorf("%s: expected not undone", cmd.name)
		}
	}
}

func TestChain_RollsBackPriorStepsInReverseOnFailure(t *testing.T) {
	a, b, c, d := &fakeCmd{name: "a"}, &fakeCmd{name: "b"}, &fakeCmd{name: "c", failOnExec: true}, &fakeCmd{name: "d"}

	err := NewChain(a, b, c, d).Run()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !a.undone || !b.undone {
		t.Error("expected a and b (already-executed steps) to be undone")
	}
	if c.undone {
		t.Error("c never executed successfully — should never be undone")
	}
	if d.executed {
		t.Error("d comes after the failed step — should never have run")
	}
}

func TestChain_StopsIfRollbackItselfFails(t *testing.T) {
	a := &fakeCmd{name: "a"}
	b := &fakeCmd{name: "b", failOnUndo: true}
	c := &fakeCmd{name: "c", failOnExec: true}

	err := NewChain(a, b, c).Run()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if a.undone {
		t.Error("rollback of b failed — a's undo should never be attempted")
	}
}

func TestChain_Plan(t *testing.T) {
	plan := NewChain(&fakeCmd{name: "a"}, &fakeCmd{name: "b"}).Plan()
	if len(plan) != 2 || plan[0] != "a" || plan[1] != "b" {
		t.Fatalf("unexpected plan: %v", plan)
	}
}

package delta

import (
	"path/filepath"
	"testing"
)

func TestCorpusStatesAndExactDeltaStatuses(t *testing.T) {
	root := filepath.Join("..", "..")
	checks := []struct {
		name string
		state string
		statuses map[string]string
	}{
		{name: "normal", state: StatusClosed, statuses: map[string]string{"ENTITY:entity:evaluator": StatusChanged, "ENTITY:entity:optimizer": StatusAdded, "EFFECT:effect:stderr": StatusRemoved, "ENTITY:entity:parser": StatusUnchanged}},
		{name: "unknown", state: StatusUnknown, statuses: map[string]string{"ENTITY:entity:parser": StatusUnknown}},
		{name: "refuted", state: StatusRefuted, statuses: map[string]string{"ENTITY:entity:parser": StatusRefuted}},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			fixture, err := LoadFixture(filepath.Join(root, "fixtures", "cases", check.name+".json"))
			if err != nil { t.Fatal(err) }
			evidence := Evaluate(fixture.Input, Inventory{})
			if evidence.State != check.state { t.Fatalf("state = %s, want %s", evidence.State, check.state) }
			for key, expected := range check.statuses {
				found := false
				for _, cell := range evidence.DeltaCells {
					if cell.Domain+":"+cell.Identity == key { found = true; if cell.Status != expected { t.Errorf("%s = %s, want %s", key, cell.Status, expected) } }
				}
				if !found { t.Errorf("missing cell %s", key) }
			}
		})
	}
}

func TestUnknownPreservesSixCoordinates(t *testing.T) {
	unknown := Unknown{Stage: "S", Step: "P", Reason: "R", UnknownClass: "C", NextOperation: "N", BlockedBy: []string{"B"}}
	if err := validUnknown(&unknown); err != nil { t.Fatal(err) }
}

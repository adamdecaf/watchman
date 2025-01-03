// Copyright The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package prepare

import (
	"testing"
)

func TestPipeline__normalizeStep(t *testing.T) {
	nn := &Name{Processed: "Nicolás Maduro"}

	step := &normalizeStep{}
	if err := step.apply(nn); err != nil {
		t.Fatal(err)
	}

	if nn.Processed != "nicolas maduro" {
		t.Errorf("nn.Processed=%v", nn.Processed)
	}
}

// TestLowerAndRemovePunctuation ensures we are trimming and UTF-8 normalizing strings
// as expected. This is needed since our datafiles are normalized for us.
func TestLowerAndRemovePunctuation(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{"remove accents", "nicolás maduro", "nicolas maduro"},
		{"convert IAcute", "Delcy Rodríguez", "delcy rodriguez"},
		{"issue 58", "Raúl Castro", "raul castro"},
		{"remove hyphen", "ANGLO-CARIBBEAN ", "anglo caribbean"},
		// Issue 483
		{"issue 483 #1", "11420 CORP.", "11420 corp"},
		{"issue 483 #2", "11,420.2-1 CORP.", "114202 1 corp"},
	}
	for i, tc := range tests {
		guess := LowerAndRemovePunctuation(tc.input)
		if guess != tc.expected {
			t.Errorf("case: %d name: %s LowerAndRemovePunctuation(%q)=%q expected %q", i, tc.name, tc.input, guess, tc.expected)
		}
	}
}

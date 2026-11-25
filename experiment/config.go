// Copyright (2026) Thiago Lizier Rissi
//
// Reduced replication of Cooper & Shepard (1973), "The time required to
// prepare for a rotated stimulus", Memory & Cognition 1(3).

package main

import "fmt"

// Kind is the type of advance information a condition gives.
type Kind int

const (
	// KindN: no information. Aperture alone (warning field), then the test.
	KindN Kind = iota
	// KindB: identity cue (outline of the normal upright character), then an
	// arrow showing where the top of the character will be, then the test.
	KindB
	// KindC: outline of the normal character already rotated to the test
	// orientation, then the empty aperture, then the test.
	KindC
)

// ConditionDef describes one condition of the paper.
type ConditionDef struct {
	Name    string
	Kind    Kind
	ArrowMs int // planned arrow duration (KindB only)
}

// conditionCatalog lists every condition of the original study. Only those
// named in Config.Conditions are run; adding "B100", "B700" or "C" back to that
// list is all it takes to run them.
var conditionCatalog = map[string]ConditionDef{
	"N":     {Name: "N", Kind: KindN},
	"B100":  {Name: "B100", Kind: KindB, ArrowMs: 100},
	"B400":  {Name: "B400", Kind: KindB, ArrowMs: 400},
	"B700":  {Name: "B700", Kind: KindB, ArrowMs: 700},
	"B1000": {Name: "B1000", Kind: KindB, ArrowMs: 1000},
	"C":     {Name: "C", Kind: KindC},
}

// Timing holds every duration of the trial timeline, in milliseconds. Each is
// converted to a whole number of display frames (Frames) before use.
type Timing struct {
	WarningMs       int // N: aperture alone before the test (paper: 2000)
	IdentityMs      int // B: outline of the normal upright character
	CombinedCueMs   int // C: outline already rotated to the test orientation
	CombinedBlankMs int // C: empty aperture between the cue and the test
	TimeoutMs       int // test stays up until the response or this timeout
	FeedbackMs      int // "Erreur" / "Trop lent"
	ITIMs           int // aperture alone after each trial
}

// Config is the whole design in one place.
type Config struct {
	Conditions           []string // run order is counterbalanced, see ConditionOrder
	Departures           []int    // angular departures from upright, degrees
	Characters           []string // test characters
	Repetitions          int      // per condition x departure x version cell
	BlocksPerCondition   int      // 2 gives the ABC-CBA order
	PracticePerCondition int      // practice trials per condition
	QuickTrialsPerBlock  int      // -quick: one block of this many per condition
	Timing               Timing
}

// Versions of the test character.
const (
	Normal   = "normal"
	Backward = "backward"
)

// Versions in the order cells are enumerated.
var Versions = []string{Normal, Backward}

// DefaultConfig is the design run by default: 3 conditions x 4 departures x 2
// versions x 6 repetitions = 144 main trials, in 6 blocks of 24.
func DefaultConfig() Config {
	return Config{
		Conditions:           []string{"N", "B400", "B1000"},
		Departures:           []int{0, 60, 120, 180},
		Characters:           []string{"G", "J", "R", "2", "5", "7"},
		Repetitions:          6,
		BlocksPerCondition:   2,
		PracticePerCondition: 2,
		QuickTrialsPerBlock:  6,
		Timing: Timing{
			WarningMs:       1000,
			IdentityMs:      2000,
			CombinedCueMs:   2000,
			CombinedBlankMs: 1000,
			TimeoutMs:       3000,
			FeedbackMs:      400,
			ITIMs:           500,
		},
	}
}

// Validate reports the first inconsistency in the configuration.
func (c Config) Validate() error {
	if len(c.Conditions) == 0 {
		return fmt.Errorf("config: no conditions")
	}
	seen := map[string]bool{}
	for _, name := range c.Conditions {
		if _, ok := conditionCatalog[name]; !ok {
			return fmt.Errorf("config: unknown condition %q", name)
		}
		if seen[name] {
			return fmt.Errorf("config: condition %q listed twice", name)
		}
		seen[name] = true
	}
	if len(c.Departures) == 0 {
		return fmt.Errorf("config: no departures")
	}
	for _, d := range c.Departures {
		if d < 0 || d > 180 {
			return fmt.Errorf("config: departure %d outside 0..180", d)
		}
	}
	if len(c.Characters) == 0 {
		return fmt.Errorf("config: no characters")
	}
	for _, ch := range c.Characters {
		if len([]rune(ch)) != 1 {
			return fmt.Errorf("config: character %q is not a single rune", ch)
		}
	}
	if c.Repetitions < 1 {
		return fmt.Errorf("config: repetitions must be >= 1")
	}
	if c.BlocksPerCondition < 1 || c.BlocksPerCondition > c.Repetitions {
		return fmt.Errorf("config: blocks per condition must be in 1..repetitions")
	}
	if c.PracticePerCondition < 0 || c.QuickTrialsPerBlock < 1 {
		return fmt.Errorf("config: bad practice or quick trial count")
	}
	t := c.Timing
	for _, v := range []int{t.WarningMs, t.IdentityMs, t.CombinedCueMs, t.CombinedBlankMs,
		t.TimeoutMs, t.FeedbackMs, t.ITIMs} {
		if v <= 0 {
			return fmt.Errorf("config: every duration must be > 0 ms")
		}
	}
	for _, name := range c.Conditions {
		if d := conditionCatalog[name]; d.Kind == KindB && d.ArrowMs <= 0 {
			return fmt.Errorf("config: condition %s has no arrow duration", name)
		}
	}
	return nil
}

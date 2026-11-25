// Copyright (2026) Thiago Lizier Rissi

package main

// Trial-list generation. Everything here is pure (no SDL), so it is covered by
// design_test.go.

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/chrplr/goxpyriment/design"
)

// Trial is one presentation of a test character.
type Trial struct {
	Practice     bool
	Block        int // 1-based; 0 for practice
	TrialInBlock int // 1-based
	Global       int // 1-based over the main trials (over practice trials for practice)
	Condition    string
	Character    string
	Version      string // Normal or Backward
	Orientation  int    // displayed orientation, degrees clockwise, 0..359
	Departure    int    // angular departure from upright, 0..180
}

// Block is a run of trials of a single condition.
type Block struct {
	Index     int // 1-based
	Condition string
	Trials    []Trial
}

// Session is everything a participant runs.
type Session struct {
	Seed     uint64
	Order    []string // condition order of the first pass (row of the Latin square)
	Practice []Trial
	Blocks   []Block
}

// seedSalt is mixed into the PCG stream so that the seed is not just the
// subject number fed to the generator twice.
const seedSalt = 0xC00A_5E9A_1973

// Seed is the random seed of a subject: the subject number itself, so that a
// session can be regenerated from the data file alone.
func Seed(subject int) uint64 { return uint64(subject) }

func newRNG(seed uint64) *rand.Rand { return rand.New(rand.NewPCG(seed, seedSalt)) }

// ConditionOrder returns the condition order for a subject: row
// (subject-1) mod rows of the balanced Latin square of the conditions (for 3
// conditions the square has 6 rows, one per permutation).
func ConditionOrder(conditions []string, subject int) ([]string, error) {
	square, err := design.LatinSquare(conditions, design.PBalancedLatinSquare)
	if err != nil {
		return nil, err
	}
	if len(square) == 0 {
		return nil, fmt.Errorf("empty Latin square")
	}
	row := (subject - 1) % len(square)
	if row < 0 {
		row += len(square)
	}
	return slices.Clone(square[row]), nil
}

// DepartureOf returns the angular departure from upright of an orientation.
func DepartureOf(orientation int) int {
	o := ((orientation % 360) + 360) % 360
	if o > 180 {
		return 360 - o
	}
	return o
}

// Orientations lists every displayed orientation the departures produce
// (60 gives 60 and 300, 0 and 180 give themselves), sorted.
func Orientations(departures []int) []int {
	var out []int
	for _, d := range departures {
		for _, o := range []int{d % 360, (360 - d) % 360} {
			if !slices.Contains(out, o) {
				out = append(out, o)
			}
		}
	}
	slices.Sort(out)
	return out
}

// hasDirection reports whether a departure can be shown clockwise or
// counterclockwise (0 and 180 cannot).
func hasDirection(dep int) bool { return dep%180 != 0 }

// assignDirections chooses clockwise (+1) or counterclockwise (-1) for each
// repetition of a cell. blockOf[r] is the block of repetition r. The totals are
// exactly nCW and nCCW, and within each block the two directions alternate so
// that each block gets as close to half and half as its size allows. cwFirst
// breaks the first tie, so that neighbouring cells start on opposite sides.
func assignDirections(blockOf []int, nBlocks, nCW, nCCW int, cwFirst bool) []int {
	dirs := make([]int, len(blockOf))
	remCW, remCCW := nCW, nCCW
	last := 1
	if cwFirst {
		last = -1
	}
	for b := 0; b < nBlocks; b++ {
		for r, bb := range blockOf {
			if bb != b {
				continue
			}
			var d int
			switch {
			case remCW > remCCW:
				d = 1
			case remCCW > remCW:
				d = -1
			default:
				d = -last
			}
			if d == 1 {
				remCW--
			} else {
				remCCW--
			}
			dirs[r], last = d, d
		}
	}
	return dirs
}

// conditionTrials builds the trials of one condition, already split into
// cfg.BlocksPerCondition blocks (not yet shuffled).
//
//   - Every departure x version cell gets cfg.Repetitions trials. Repetition r
//     of cell c goes to block (r+c) mod B, so with 6 repetitions and 2 blocks
//     each block holds every cell 3 times (and with an odd count the blocks
//     still have equal sizes).
//   - 60 and 120 are shown clockwise on half the repetitions of the cell and
//     counterclockwise on the other half; with an odd count the extra one is
//     clockwise for normal and counterclockwise for backward, so each
//     condition x departure stays balanced.
//   - Characters are dealt in a cycle over the cells of the condition, so each
//     character appears equally often (8 times for 48 trials and 6 characters;
//     with 6 repetitions, once per cell, hence evenly over departures and
//     versions), then shuffled within each cell.
func conditionTrials(cfg Config, cond string, rng *rand.Rand) [][]Trial {
	nB, nR, nChars := cfg.BlocksPerCondition, cfg.Repetitions, len(cfg.Characters)
	blocks := make([][]Trial, nB)
	charOffset := rng.IntN(nChars)
	cell := 0
	for _, dep := range cfg.Departures {
		for _, ver := range Versions {
			blockOf := make([]int, nR)
			for r := range nR {
				blockOf[r] = (r + cell) % nB
			}
			dirs := make([]int, nR)
			if hasDirection(dep) {
				nCW := nR / 2
				if nR%2 == 1 && ver == Normal {
					nCW++
				}
				dirs = assignDirections(blockOf, nB, nCW, nR-nCW, cell%2 == 0)
			}
			chars := make([]string, nR)
			for r := range nR {
				chars[r] = cfg.Characters[(cell*nR+r+charOffset)%nChars]
			}
			rng.Shuffle(nR, func(i, j int) { chars[i], chars[j] = chars[j], chars[i] })
			for r := range nR {
				orient := dep % 360
				if dirs[r] < 0 {
					orient = (360 - dep) % 360
				}
				blocks[blockOf[r]] = append(blocks[blockOf[r]], Trial{
					Condition:   cond,
					Character:   chars[r],
					Version:     ver,
					Orientation: orient,
					Departure:   dep,
				})
			}
			cell++
		}
	}
	return blocks
}

// BuildPractice returns the practice trials: cfg.PracticePerCondition per
// condition, with random characters, orientations and versions, in random
// order.
func BuildPractice(cfg Config, rng *rand.Rand) []Trial {
	orients := Orientations(cfg.Departures)
	var out []Trial
	for _, cond := range cfg.Conditions {
		for range cfg.PracticePerCondition {
			o := orients[rng.IntN(len(orients))]
			out = append(out, Trial{
				Practice:    true,
				Condition:   cond,
				Character:   cfg.Characters[rng.IntN(len(cfg.Characters))],
				Version:     Versions[rng.IntN(len(Versions))],
				Orientation: o,
				Departure:   DepartureOf(o),
			})
		}
	}
	rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	for i := range out {
		out[i].TrialInBlock, out[i].Global = i+1, i+1
	}
	return out
}

// BuildSession generates the session of a subject. The same configuration and
// subject always give the same session.
//
// Full session: each condition's trials are split into BlocksPerCondition
// blocks; the blocks run in passes, the first pass in the subject's condition
// order (ConditionOrder), the second in the reverse order, and so on (ABC-CBA
// for 3 conditions and 2 blocks each). Trials are shuffled within each block.
//
// Quick session (-quick): one block per condition in the subject's order, each
// of QuickTrialsPerBlock trials drawn at random from that condition's trials.
func BuildSession(cfg Config, subject int, quick bool) (Session, error) {
	if err := cfg.Validate(); err != nil {
		return Session{}, err
	}
	if subject < 1 {
		return Session{}, fmt.Errorf("subject ID must be a positive integer, got %d", subject)
	}
	order, err := ConditionOrder(cfg.Conditions, subject)
	if err != nil {
		return Session{}, err
	}
	seed := Seed(subject)
	rng := newRNG(seed)

	// Generated in configuration order, not run order, so that a condition's
	// trials do not depend on which counterbalancing row the subject is in.
	perCond := map[string][][]Trial{}
	for _, cond := range cfg.Conditions {
		perCond[cond] = conditionTrials(cfg, cond, rng)
	}

	var blocks []Block
	if quick {
		for _, cond := range order {
			all := slices.Concat(perCond[cond]...)
			rng.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
			n := min(cfg.QuickTrialsPerBlock, len(all))
			blocks = append(blocks, Block{Condition: cond, Trials: all[:n]})
		}
	} else {
		for pass := 0; pass < cfg.BlocksPerCondition; pass++ {
			seq := slices.Clone(order)
			if pass%2 == 1 {
				slices.Reverse(seq)
			}
			for _, cond := range seq {
				trials := slices.Clone(perCond[cond][pass])
				rng.Shuffle(len(trials), func(i, j int) { trials[i], trials[j] = trials[j], trials[i] })
				blocks = append(blocks, Block{Condition: cond, Trials: trials})
			}
		}
	}
	global := 0
	for b := range blocks {
		blocks[b].Index = b + 1
		for i := range blocks[b].Trials {
			global++
			t := &blocks[b].Trials[i]
			t.Block, t.TrialInBlock, t.Global = b+1, i+1, global
		}
	}

	return Session{
		Seed:     seed,
		Order:    order,
		Practice: BuildPractice(cfg, rng),
		Blocks:   blocks,
	}, nil
}

// BlockOrder is the sequence of conditions of the blocks, e.g. "N,B400,B1000,B1000,B400,N".
func (s Session) BlockOrder() string {
	names := make([]string, len(s.Blocks))
	for i, b := range s.Blocks {
		names[i] = b.Condition
	}
	return strings.Join(names, ",")
}

// NumMainTrials counts the trials of all blocks.
func (s Session) NumMainTrials() int {
	n := 0
	for _, b := range s.Blocks {
		n += len(b.Trials)
	}
	return n
}

// KeyMapping says which key means which response. Keys are "f" and "j".
type KeyMapping struct {
	NormalKey, BackwardKey string
}

// MappingFor counterbalances the response keys by subject, independently of
// the condition order. The order is row (subject-1) mod 6 of the Latin square
// (ConditionOrder); the keys alternate every 6 subjects: subjects 1-6 press J
// for normal and F for backward (as in the paper), 7-12 the reverse, and so
// on. Subjects 1-12 thus cover each order with each mapping exactly once, and
// for a random subject number the two are independent.
func MappingFor(subject int) KeyMapping {
	if ((subject-1)/6)%2 == 0 {
		return KeyMapping{NormalKey: "j", BackwardKey: "f"}
	}
	return KeyMapping{NormalKey: "f", BackwardKey: "j"}
}

// String is what goes into the key_mapping column, e.g. "J=normal;F=backward".
func (m KeyMapping) String() string {
	return fmt.Sprintf("%s=normal;%s=backward", strings.ToUpper(m.NormalKey), strings.ToUpper(m.BackwardKey))
}

// Response maps a key name ("f", "j", or "" for none) to a response.
func (m KeyMapping) Response(key string) string {
	switch key {
	case m.NormalKey:
		return Normal
	case m.BackwardKey:
		return Backward
	}
	return "none"
}

// Frames converts a duration in milliseconds to a whole number of display
// frames (at least 1).
func Frames(ms int, frame time.Duration) int {
	if frame <= 0 {
		frame = time.Second / 60
	}
	n := int(math.Round(float64(ms) * float64(time.Millisecond) / float64(frame)))
	return max(n, 1)
}

// framePeriod estimates the display's frame period from the timestamps (ns) of
// consecutive flips: the span from the first to the last flip divided by the
// number of frame periods it holds. Each interval is counted as a whole number
// of periods (rounded against the median interval), so a dropped frame (an
// interval of two periods) lengthens the span and the count together and does
// not bias the estimate. Rounding each interval, not the whole span, matters:
// the median is itself only good to a clock tick, and over 120 flips that error
// adds up to a whole frame.
//
// The median of single intervals, which is what Screen.CalibrateRefresh
// returns, is not precise enough in a browser: without cross-origin isolation
// its clock ticks in 0.1 ms steps, so on a 144 Hz display (6.945 ms) single
// intervals read 6.9 or 7.0 ms and the median says 7.000 ms. Counted in those
// frames, a 2000 ms cue came out 286 frames = 1986 ms long. Over 120 flips the
// span-based estimate is good to about a microsecond.
func framePeriod(ts []uint64) (time.Duration, error) {
	if len(ts) < 3 {
		return 0, fmt.Errorf("framePeriod: need at least 3 flips, got %d", len(ts))
	}
	iv := make([]float64, len(ts)-1)
	for i := range iv {
		if ts[i+1] <= ts[i] {
			return 0, fmt.Errorf("framePeriod: timestamps not increasing at flip %d", i+1)
		}
		iv[i] = float64(ts[i+1] - ts[i])
	}
	slices.Sort(iv)
	median := iv[len(iv)/2]
	span := float64(ts[len(ts)-1] - ts[0])
	periods := 0.0
	for i := 1; i < len(ts); i++ {
		periods += max(1, math.Round(float64(ts[i]-ts[i-1])/median))
	}
	if periods < 1 {
		return 0, fmt.Errorf("framePeriod: degenerate timestamps")
	}
	return time.Duration(math.Round(span / periods)), nil
}

// TrialDuration is the expected duration of one trial of a condition, given
// the frame duration, a mean response time and an error-plus-timeout rate.
func TrialDuration(cfg Config, cond string, frame time.Duration, meanRT time.Duration, errRate float64) time.Duration {
	t := cfg.Timing
	f := func(ms int) time.Duration { return time.Duration(Frames(ms, frame)) * frame }
	var pre time.Duration
	switch d := conditionCatalog[cond]; d.Kind {
	case KindN:
		pre = f(t.WarningMs)
	case KindB:
		pre = f(t.IdentityMs) + f(d.ArrowMs)
	case KindC:
		pre = f(t.CombinedCueMs) + f(t.CombinedBlankMs)
	}
	fb := time.Duration(errRate * float64(f(t.FeedbackMs)))
	return pre + meanRT + fb + f(t.ITIMs)
}

// EstimateTrials is the expected total duration of the practice and main
// trials of a session (instruction screens excluded).
func EstimateTrials(cfg Config, s Session, frame, meanRT time.Duration, errRate float64) time.Duration {
	var total time.Duration
	for _, t := range s.Practice {
		total += TrialDuration(cfg, t.Condition, frame, meanRT, errRate)
	}
	for _, b := range s.Blocks {
		for _, t := range b.Trials {
			total += TrialDuration(cfg, t.Condition, frame, meanRT, errRate)
		}
	}
	return total
}

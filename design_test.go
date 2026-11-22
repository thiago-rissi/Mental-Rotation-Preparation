// Copyright (2026) Thiago Lizier Rissi

package main

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func mustSession(t *testing.T, cfg Config, subject int, quick bool) Session {
	t.Helper()
	s, err := BuildSession(cfg, subject, quick)
	if err != nil {
		t.Fatalf("BuildSession(subject %d): %v", subject, err)
	}
	return s
}

func allTrials(s Session) []Trial {
	var out []Trial
	for _, b := range s.Blocks {
		out = append(out, b.Trials...)
	}
	return out
}

func TestTotalTrialsAndBlocks(t *testing.T) {
	for subject := 1; subject <= 12; subject++ {
		s := mustSession(t, DefaultConfig(), subject, false)
		if n := s.NumMainTrials(); n != 144 {
			t.Fatalf("subject %d: %d main trials, want 144", subject, n)
		}
		if len(s.Blocks) != 6 {
			t.Fatalf("subject %d: %d blocks, want 6", subject, len(s.Blocks))
		}
		global := 0
		for bi, b := range s.Blocks {
			if len(b.Trials) != 24 {
				t.Errorf("subject %d block %d: %d trials, want 24", subject, bi+1, len(b.Trials))
			}
			for i, tr := range b.Trials {
				global++
				if tr.Condition != b.Condition || tr.Block != bi+1 || tr.TrialInBlock != i+1 || tr.Global != global || tr.Practice {
					t.Fatalf("subject %d: bad trial bookkeeping %+v in block %d", subject, tr, bi+1)
				}
			}
		}
	}
}

func TestPerBlockCellCounts(t *testing.T) {
	cfg := DefaultConfig()
	for subject := 1; subject <= 12; subject++ {
		s := mustSession(t, cfg, subject, false)
		for _, b := range s.Blocks {
			counts := map[string]int{}
			for _, tr := range b.Trials {
				counts[fmt.Sprintf("%d/%s", tr.Departure, tr.Version)]++
			}
			for _, dep := range cfg.Departures {
				for _, ver := range Versions {
					if n := counts[fmt.Sprintf("%d/%s", dep, ver)]; n != 3 {
						t.Errorf("subject %d block %d (%s): cell %d/%s has %d trials, want 3",
							subject, b.Index, b.Condition, dep, ver, n)
					}
				}
			}
		}
	}
}

func TestOrientationsMatchDepartures(t *testing.T) {
	s := mustSession(t, DefaultConfig(), 3, false)
	for _, tr := range append(allTrials(s), s.Practice...) {
		if DepartureOf(tr.Orientation) != tr.Departure {
			t.Fatalf("orientation %d logged with departure %d", tr.Orientation, tr.Departure)
		}
		if tr.Orientation < 0 || tr.Orientation >= 360 {
			t.Fatalf("orientation %d out of 0..359", tr.Orientation)
		}
	}
}

// directionCounts counts clockwise / counterclockwise trials per
// condition x departure x version.
func directionCounts(trials []Trial) map[string][2]int {
	out := map[string][2]int{}
	for _, tr := range trials {
		if !hasDirection(tr.Departure) {
			continue
		}
		k := fmt.Sprintf("%s/%d/%s", tr.Condition, tr.Departure, tr.Version)
		c := out[k]
		if tr.Orientation == tr.Departure {
			c[0]++
		} else {
			c[1]++
		}
		out[k] = c
	}
	return out
}

func TestClockwiseBalance(t *testing.T) {
	cfg := DefaultConfig()
	for subject := 1; subject <= 12; subject++ {
		s := mustSession(t, cfg, subject, false)
		counts := directionCounts(allTrials(s))
		if len(counts) != 3*2*2 {
			t.Fatalf("subject %d: %d rotated cells, want 12", subject, len(counts))
		}
		for k, c := range counts {
			if c != [2]int{3, 3} {
				t.Errorf("subject %d %s: cw/ccw = %v, want 3/3", subject, k, c)
			}
		}
		// Within a block, each condition x departure is also split evenly
		// (3 normal + 3 backward trials: 2/1 and 1/2).
		for _, b := range s.Blocks {
			perDep := map[int][2]int{}
			for _, tr := range b.Trials {
				if hasDirection(tr.Departure) {
					c := perDep[tr.Departure]
					if tr.Orientation == tr.Departure {
						c[0]++
					} else {
						c[1]++
					}
					perDep[tr.Departure] = c
				}
			}
			for dep, c := range perDep {
				if c != [2]int{3, 3} {
					t.Errorf("subject %d block %d departure %d: cw/ccw = %v, want 3/3", subject, b.Index, dep, c)
				}
			}
		}
	}
}

func TestClockwiseBalanceFiveRepetitions(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repetitions = 5
	s := mustSession(t, cfg, 1, false)
	if n := s.NumMainTrials(); n != 120 {
		t.Fatalf("%d main trials, want 120", n)
	}
	for _, b := range s.Blocks {
		if len(b.Trials) != 20 {
			t.Errorf("block %d has %d trials, want 20", b.Index, len(b.Trials))
		}
	}
	for k, c := range directionCounts(allTrials(s)) {
		want := [2]int{2, 3}
		if strings.HasSuffix(k, "/"+Normal) {
			want = [2]int{3, 2}
		}
		if c != want {
			t.Errorf("%s: cw/ccw = %v, want %v", k, c, want)
		}
	}
}

func TestCharacterBalance(t *testing.T) {
	cfg := DefaultConfig()
	for subject := 1; subject <= 12; subject++ {
		s := mustSession(t, cfg, subject, false)
		perCond := map[string]int{}
		perDep := map[string]int{}
		perVer := map[string]int{}
		for _, tr := range allTrials(s) {
			perCond[tr.Condition+"/"+tr.Character]++
			perDep[fmt.Sprintf("%s/%s/%d", tr.Condition, tr.Character, tr.Departure)]++
			perVer[tr.Condition+"/"+tr.Character+"/"+tr.Version]++
		}
		for _, cond := range cfg.Conditions {
			for _, ch := range cfg.Characters {
				if n := perCond[cond+"/"+ch]; n != 8 {
					t.Errorf("subject %d %s: character %s appears %d times, want 8", subject, cond, ch, n)
				}
				for _, dep := range cfg.Departures {
					if n := perDep[fmt.Sprintf("%s/%s/%d", cond, ch, dep)]; n != 2 {
						t.Errorf("subject %d %s %s departure %d: %d, want 2", subject, cond, ch, dep, n)
					}
				}
				for _, ver := range Versions {
					if n := perVer[cond+"/"+ch+"/"+ver]; n != 4 {
						t.Errorf("subject %d %s %s %s: %d, want 4", subject, cond, ch, ver, n)
					}
				}
			}
		}
	}
}

func TestBlockOrder(t *testing.T) {
	cfg := DefaultConfig()
	firstPass := map[string]bool{}
	firstCount := map[string]int{}
	for subject := 1; subject <= 12; subject++ {
		s := mustSession(t, cfg, subject, false)
		order, err := ConditionOrder(cfg.Conditions, subject)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(s.Order, order) {
			t.Fatalf("subject %d: session order %v != ConditionOrder %v", subject, s.Order, order)
		}
		got := strings.Split(s.BlockOrder(), ",")
		rev := slices.Clone(order)
		slices.Reverse(rev)
		want := slices.Concat(order, rev)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("subject %d: blocks %v, want ABC-CBA %v", subject, got, want)
		}
		sorted := slices.Clone(order)
		slices.Sort(sorted)
		wantSorted := slices.Clone(cfg.Conditions)
		slices.Sort(wantSorted)
		if !reflect.DeepEqual(sorted, wantSorted) {
			t.Fatalf("subject %d: order %v is not a permutation of %v", subject, order, cfg.Conditions)
		}
		if subject <= 6 {
			firstPass[strings.Join(order, ",")] = true
			firstCount[order[0]]++
		} else {
			prev, _ := ConditionOrder(cfg.Conditions, subject-6)
			if !reflect.DeepEqual(order, prev) {
				t.Errorf("subject %d: order %v differs from subject %d's %v", subject, order, subject-6, prev)
			}
		}
		t.Logf("subject %2d: %s", subject, s.BlockOrder())
	}
	if len(firstPass) != 6 {
		t.Errorf("subjects 1-6 use %d distinct orders, want all 6 permutations", len(firstPass))
	}
	for _, c := range cfg.Conditions {
		if firstCount[c] != 2 {
			t.Errorf("condition %s comes first for %d of subjects 1-6, want 2", c, firstCount[c])
		}
	}
}

func TestKeyMapping(t *testing.T) {
	for subject := 1; subject <= 12; subject++ {
		m := MappingFor(subject)
		if subject%2 == 1 {
			if m.NormalKey != "j" || m.BackwardKey != "f" || m.String() != "J=normal;F=backward" {
				t.Errorf("subject %d (odd): mapping %+v", subject, m)
			}
		} else {
			if m.NormalKey != "f" || m.BackwardKey != "j" || m.String() != "F=normal;J=backward" {
				t.Errorf("subject %d (even): mapping %+v", subject, m)
			}
		}
		if m.Response(m.NormalKey) != Normal || m.Response(m.BackwardKey) != Backward || m.Response("") != "none" {
			t.Errorf("subject %d: Response broken", subject)
		}
	}
}

func TestReproducibleForSameSeed(t *testing.T) {
	cfg := DefaultConfig()
	for _, quick := range []bool{false, true} {
		a := mustSession(t, cfg, 7, quick)
		b := mustSession(t, cfg, 7, quick)
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("quick=%v: two sessions for subject 7 differ", quick)
		}
		if a.Seed != 7 {
			t.Errorf("seed %d, want 7", a.Seed)
		}
	}
	a := mustSession(t, cfg, 1, false)
	c := mustSession(t, cfg, 7, false) // same Latin-square row as subject 1
	if reflect.DeepEqual(allTrials(a), allTrials(c)) {
		t.Errorf("subjects 1 and 7 got identical trial sequences")
	}
}

func TestQuickSession(t *testing.T) {
	cfg := DefaultConfig()
	s := mustSession(t, cfg, 2, true)
	if len(s.Blocks) != 3 {
		t.Fatalf("%d quick blocks, want 3", len(s.Blocks))
	}
	for i, b := range s.Blocks {
		if len(b.Trials) != 6 || b.Condition != s.Order[i] {
			t.Errorf("quick block %d: %s with %d trials", i+1, b.Condition, len(b.Trials))
		}
	}
}

func TestPractice(t *testing.T) {
	cfg := DefaultConfig()
	s := mustSession(t, cfg, 5, false)
	if len(s.Practice) != 6 {
		t.Fatalf("%d practice trials, want 6", len(s.Practice))
	}
	per := map[string]int{}
	for i, tr := range s.Practice {
		per[tr.Condition]++
		if !tr.Practice || tr.Block != 0 || tr.TrialInBlock != i+1 {
			t.Errorf("practice trial %d: %+v", i+1, tr)
		}
	}
	for _, c := range cfg.Conditions {
		if per[c] != 2 {
			t.Errorf("condition %s: %d practice trials, want 2", c, per[c])
		}
	}
}

func TestConditionCStillSelectable(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Conditions = []string{"N", "B100", "B400", "B700", "B1000", "C"}
	s := mustSession(t, cfg, 1, false)
	if n := s.NumMainTrials(); n != 6*48 {
		t.Fatalf("%d trials, want %d", n, 6*48)
	}
	cfg.Conditions = []string{"N", "X"}
	if _, err := BuildSession(cfg, 1, false); err == nil {
		t.Error("unknown condition accepted")
	}
}

func TestSubjectMustBePositive(t *testing.T) {
	for _, id := range []int{0, -3} {
		if _, err := BuildSession(DefaultConfig(), id, false); err == nil {
			t.Errorf("subject %d accepted", id)
		}
	}
}

func TestFrames(t *testing.T) {
	hz60 := time.Second / 60
	hz144 := time.Second / 144
	cases := []struct {
		ms    int
		frame time.Duration
		want  int
	}{
		{400, hz60, 24}, {1000, hz60, 60}, {2000, hz60, 120}, {100, hz144, 14},
		{400, hz144, 58}, {1, hz60, 1}, {500, time.Duration(16_683_350), 30}, // 59.94 Hz
	}
	for _, c := range cases {
		if got := Frames(c.ms, c.frame); got != c.want {
			t.Errorf("Frames(%d, %v) = %d, want %d", c.ms, c.frame, got, c.want)
		}
	}
}

func TestEstimate(t *testing.T) {
	cfg := DefaultConfig()
	s := mustSession(t, cfg, 1, false)
	est := EstimateTrials(cfg, s, time.Second/60, 700*time.Millisecond, 0)
	// 48 x (2.2 + 3.6 + 4.2) s for the main trials, 2 x that for the practice.
	want := 50 * 10 * time.Second
	if d := est - want; d < -time.Millisecond || d > time.Millisecond {
		t.Errorf("estimate %v, want %v", est, want)
	}
}

// framePeriod recovers the true period from timestamps quantised to 0.1 ms
// (the browser clock without cross-origin isolation), with a dropped frame.
func TestFramePeriod(t *testing.T) {
	const period = 6_945_400.0 // 143.98 Hz, in ns
	var ts []uint64
	frame := 0
	for i := 0; i < 120; i++ {
		if i == 50 {
			frame++ // a dropped frame: one interval of two periods
		}
		exact := 1e9 + float64(frame)*period
		ts = append(ts, uint64(exact/100_000)*100_000) // 0.1 ms ticks
		frame++
	}
	got, err := framePeriod(ts)
	if err != nil {
		t.Fatal(err)
	}
	if d := float64(got) - period; d < -2000 || d > 2000 {
		t.Errorf("framePeriod = %v, want %.4f ms within 2 us", got, period/1e6)
	}
	if _, err := framePeriod([]uint64{1, 2}); err == nil {
		t.Error("two flips accepted")
	}
}

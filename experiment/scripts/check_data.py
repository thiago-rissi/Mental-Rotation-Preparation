#!/usr/bin/env python3
# Copyright (2026) Thiago Lizier Rissi
"""Sanity checks for the Mental-Rotation-Preparation data files.

For every session in the data directory (default: ./data), whether saved by
the desktop program (a .csv and its -info.txt) or downloaded by the browser
build (one .zip holding both):

  * trial counts per condition x angular departure x version (main trials);
  * total session duration (from the -info.txt) and duration of the trials;
  * accuracy and timeouts, per condition, for practice and main trials;
  * the timing manipulation check: the measured arrow duration
    (arrow_ms_measured - cue_ms_planned) per B condition, and the measured
    pre-test interval for N and C, flagging every trial off by more than one
    frame;
  * flags set by the program: autopilot runs and interrupted sessions.

Usage:  python scripts/check_data.py [DATA_DIR ...]
Exit status 1 if any file has a problem worth looking at. Needs pandas only.
"""

from __future__ import annotations

import argparse
import io
import re
import zipfile
import sys
from pathlib import Path

import pandas as pd

pd.set_option("display.width", 120)
pd.set_option("display.max_columns", 20)

EXPECTED_COLUMNS = [
    "subject_id", "practice", "block_index", "trial_in_block", "trial_global", "condition",
    "cue_ms_planned", "cue_frames", "character", "version", "orientation_deg",
    "angular_departure", "key_mapping", "response_key", "response", "correct",
    "timeout", "rt_ms", "test_onset_ns", "key_ts_ns", "identity_onset_ns",
    "arrow_onset_ns", "blank_onset_ns", "arrow_ms_measured", "frame_ms", "seed",
]


class Session:
    """One session: its CSV and -info.txt, from a folder (desktop) or from the
    single .zip a browser session downloads."""

    def __init__(self, path: Path):
        self.name = path.name
        self.csv: io.StringIO | Path | None = None
        self.info_text = ""
        if path.suffix == ".zip":
            with zipfile.ZipFile(path) as z:
                for member in z.namelist():
                    data = z.read(member).decode("utf-8", errors="replace")
                    if member.endswith("-info.txt"):
                        self.info_text = data
                    elif member.endswith(".csv"):
                        self.csv = io.StringIO(data)
        else:
            self.csv = path
            info = path.with_name(path.stem + "-info.txt")
            if info.exists():
                self.info_text = info.read_text(encoding="utf-8", errors="replace")


def read_info(text: str) -> dict:
    """Pick the useful lines out of the -info.txt."""
    out = {"autopilot": False, "quick": False, "interrupted": None, "timing": {}}
    if not text:
        return out
    for key in ("start_time", "end_time", "duration"):
        m = re.search(rf"^# e {key}: (.+)$", text, re.M)
        if m:
            out[key] = m.group(1).strip()
    out["autopilot"] = "AUTOPILOT" in text
    out["quick"] = re.search(r"^# e quick: true", text, re.M) is not None
    m = re.search(r"^# e (INTERRUPTED.*)$", text, re.M)
    if m:
        out["interrupted"] = m.group(1)
    m = re.search(r"Timing:\{([^}]*)\}", text)
    if m:
        out["timing"] = {k: int(v) for k, v in re.findall(r"(\w+):(\d+)", m.group(1))}
    return out


def section(title: str) -> None:
    print(f"\n  -- {title}")


def indent(obj) -> str:
    return "\n".join("     " + line for line in str(obj).splitlines())


def check_interval(rows: pd.DataFrame, measured: pd.Series, planned_ms: float | pd.Series,
                   label: str, problems: list[str]) -> None:
    """Describe measured - planned (ms) and flag rows off by more than one frame."""
    diff = (measured - planned_ms).dropna()
    if diff.empty:
        return
    print(f"     {label}: n={len(diff)}  mean={diff.mean():+.2f} ms  "
          f"min={diff.min():+.2f}  max={diff.max():+.2f}")
    off = diff[diff.abs() > rows.loc[diff.index, "frame_ms"]]
    for i in off.index:
        r = rows.loc[i]
        print(f"       ! block {r.block_index} trial {r.trial_in_block} (practice={r.practice}): "
              f"off by {diff[i]:+.2f} ms")
    if len(off):
        problems.append(f"{len(off)} {label} trial(s) off by more than one frame")


def check_file(sess: Session) -> list[str]:
    problems: list[str] = []
    print("=" * 100)
    print(sess.name)
    if sess.csv is None:
        print("  ! no CSV in the archive")
        return ["no CSV"]
    try:
        df = pd.read_csv(sess.csv)
    except pd.errors.EmptyDataError:
        print("  (empty file: -sheet run or a session that ended before the first trial)")
        return problems
    if df.empty:
        print("  (no trials)")
        return problems
    missing = [c for c in EXPECTED_COLUMNS if c not in df.columns]
    if missing:
        print(f"  ! missing columns: {missing}")
        return [f"missing columns {missing}"]

    info = read_info(sess.info_text)
    main = df[df.practice == 0]
    prac = df[df.practice == 1]

    print(f"  subject {df.subject_id.iloc[0]}  key mapping {df.key_mapping.iloc[0]}  "
          f"seed {df.seed.iloc[0]}  frame {df.frame_ms.iloc[0]} ms")
    order = main.groupby("block_index").condition.first()
    print(f"  block order: {','.join(order)}")
    if info["autopilot"]:
        print("  ! AUTOPILOT run: responses generated by the program, not a participant")
        problems.append("autopilot run")
    if info["interrupted"]:
        print(f"  ! {info['interrupted']}")
        problems.append("session interrupted")

    section("duration")
    if "duration" in info:
        print(f"     session (program start to end): {info['duration']}  "
              f"(start {info.get('start_time')}, end {info.get('end_time')})")
    onsets = df[["identity_onset_ns", "blank_onset_ns", "test_onset_ns"]].min(axis=1)
    ends = df[["test_onset_ns", "key_ts_ns"]].max(axis=1)
    print(f"     first cue to last response: {(ends.max() - onsets.min()) / 6e10:.2f} min")

    section(f"main trials: {len(main)} (practice: {len(prac)})")
    counts = pd.crosstab([main.condition, main.angular_departure], main.version)
    print(indent(counts))
    if info["quick"]:
        print("     (-quick run: cells are not balanced by design)")
    elif not counts.empty and counts.values.min() != counts.values.max():
        print("     ! unequal cell counts")
        problems.append("unequal cell counts")
    rot = main[main.angular_departure.isin([60, 120])]
    if not rot.empty:
        cw = rot.orientation_deg == rot.angular_departure
        print(indent(pd.crosstab([rot.condition, rot.angular_departure],
                                 cw.map({True: "cw", False: "ccw"}))))

    section("accuracy and timeouts")
    for name, part in (("practice", prac), ("main", main)):
        if part.empty:
            continue
        g = part.groupby("condition").agg(
            n=("correct", "size"), accuracy=("correct", "mean"), timeouts=("timeout", "sum"))
        g.loc["ALL"] = [len(part), part.correct.mean(), part.timeout.sum()]
        g = g.astype({"n": int, "timeouts": int})
        g["accuracy"] = g.accuracy.round(3)
        print(f"     {name}:")
        print(indent(g))
    if not main.empty and main.correct.mean() < 0.8:
        problems.append(f"main accuracy {main.correct.mean():.2f} < 0.80")
    ok = main[(main.correct == 1)]
    if not ok.empty:
        print("     median RT of correct main trials (ms), condition x departure:")
        print(indent(ok.pivot_table(index="condition", columns="angular_departure",
                                    values="rt_ms", aggfunc="median").round(0)))

    section("timing manipulation check (measured - planned, ms; '!' = off by > 1 frame)")
    b = df[df.cue_ms_planned.notna()]
    for cond, rows in b.groupby("condition"):
        check_interval(rows, rows.arrow_ms_measured, rows.cue_ms_planned, f"{cond} arrow", problems)
    t = info["timing"]
    if t:
        b_ident = b[b.identity_onset_ns.notna() & b.arrow_onset_ns.notna()]
        for cond, rows in b_ident.groupby("condition"):
            check_interval(rows, (rows.arrow_onset_ns - rows.identity_onset_ns) / 1e6,
                           t.get("IdentityMs"), f"{cond} identity cue", problems)
        n = df[df.condition == "N"]
        check_interval(n, (n.test_onset_ns - n.blank_onset_ns) / 1e6,
                       t.get("WarningMs"), "N warning field", problems)
        c = df[df.condition == "C"]
        check_interval(c, (c.test_onset_ns - c.blank_onset_ns) / 1e6,
                       t.get("CombinedBlankMs"), "C blank", problems)
    else:
        print("     (no timing config in the -info.txt: N/C intervals not checked)")
    return problems


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("dirs", nargs="*", default=["data"], help="data directories (default: data)")
    args = ap.parse_args()

    # Desktop sessions (.csv + -info.txt) and browser sessions (one .zip each).
    files = sorted(f for d in args.dirs for pattern in ("*.csv", "*.zip") for f in Path(d).glob(pattern))
    if not files:
        print(f"no .csv or .zip files in {', '.join(args.dirs)}")
        return 1
    summary = {}
    for f in files:
        summary[f.name] = check_file(Session(f))
    print("=" * 100)
    print("SUMMARY")
    bad = 0
    for name, problems in summary.items():
        print(f"  {'OK  ' if not problems else 'LOOK'}  {name}" + (f": {'; '.join(problems)}" if problems else ""))
        bad += bool(problems)
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())

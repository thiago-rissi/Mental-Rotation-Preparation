Mental Rotation Preparation: the experiment
===========================================

This folder holds the program of the experiment described in the [report](https://thiago-rissi.github.io/Mental-Rotation-Preparation/). It is written in [Go](https://go.dev) with the [goxpyriment](https://chrplr.github.io/goxpyriment/) library, and runs in a web browser or as a program on Windows, macOS and Linux.

- [Running the experiment](#running-the-experiment)
- [Options](#options)
- [Data files](#data-files)
- [Source files](#source-files)
- [Building](#building)
- [Checks](#checks)


## Running the experiment

Participants do not need a number. Each session draws a random six-digit number, which sets the order of the conditions and the response keys, and appears in the names of the data files.

### In a web browser

Open <https://thiago-rissi.github.io/Mental-Rotation-Preparation/experiment/> in Chrome, Edge or Firefox, on a computer with a keyboard, and click Start. Keep the tab in front until the end screen. The results are then downloaded as a `.zip` file.

### With a prebuilt program

Download the file for your system from the [Releases page](https://github.com/thiago-rissi/Mental-Rotation-Preparation/releases):

System                    | File
------------------------- | -----------------------------------------
Windows                   | `mental-rotation-preparation-windows.exe`
macOS, Apple processor    | `mental-rotation-preparation-macos-arm64`
macOS, Intel processor    | `mental-rotation-preparation-macos-amd64`
Linux                     | `mental-rotation-preparation-linux`

On Windows, double-click the file. If SmartScreen stops it, click "More info", then "Run anyway".

On macOS, open the Terminal in the download folder and type:

    chmod +x mental-rotation-preparation-macos-arm64
    xattr -d com.apple.quarantine mental-rotation-preparation-macos-arm64
    ./mental-rotation-preparation-macos-arm64

On an Intel Mac, replace `arm64` by `amd64`.

On Linux, type:

    chmod +x mental-rotation-preparation-linux
    ./mental-rotation-preparation-linux

The program starts at once, in full screen. The last screen shows where the data were saved.

### From the source code

Install Go 1.25.7 or later from <https://go.dev/dl/>. Nothing else is needed. Then type:

    git clone https://github.com/thiago-rissi/Mental-Rotation-Preparation
    cd Mental-Rotation-Preparation/experiment
    go run . -s 1 -w

`go build .` makes an executable in this folder instead.


## Options

Option       | Effect
------------ | ------
`-s N`       | use participant number N instead of a random one
`-w`         | run in a window instead of full screen
`-lang fr`   | instructions in French
`-quick`     | three short blocks, to try the program
`-sheet`     | show all the stimuli, to check them by eye
`-autopilot` | the program answers by itself, for testing only
`-h`         | list all options

ESC stops the session at any time, and the trials done so far are saved. In the browser, options go in the address, for example `?s=1&quick`.


## Data files

Each session produces a CSV file, `Mental-Rotation-Preparation_sub-<number>_date-<date>.csv`, with one line per trial, and a file ending in `-info.txt` that describes the session and the computer. The desktop version also saves the answer to the final question in a file ending in `_debrief.txt`.

The desktop version saves these files in a folder `data`, inside the folder from which it was started. The browser version downloads them in a `.zip` file. We store the files of all participants in [data](data).

The main columns of the CSV file are:

Column              | Content
------------------- | -------
`condition`         | N, B400 or B1000
`character`         | G, J, R, 2, 5 or 7
`version`           | normal or backward
`orientation_deg`   | orientation on the screen, in degrees clockwise
`angular_departure` | 0, 60, 120 or 180
`response`          | normal, backward or none
`correct`           | 1 or 0
`rt_ms`             | response time, in milliseconds
`arrow_ms_measured` | measured duration of the arrow, in milliseconds
`practice`          | 1 for practice trials
`timeout`           | 1 when there was no response


## Source files

- [main.go](main.go): the procedure (instructions, blocks, trials and data logging)
- [config.go](config.go): the conditions, characters, angles, repetitions and durations
- [design.go](design.go): the list of trials and the counterbalancing
- [stimuli.go](stimuli.go): the drawing of the characters, outlines, arrow and frame
- [strings.go](strings.go): the texts shown to the participant, in English and French
- [richtext.go](richtext.go): the display of these texts, with bold words and keys drawn in boxes
- [textinput.go](textinput.go): the answer box of the final question
- [subject.go](subject.go): the random participant number
- [sheet.go](sheet.go): the stimulus sheet (option `-sheet`)
- [autopilot.go](autopilot.go), [keys_notjs.go](keys_notjs.go), [keys_js.go](keys_js.go): the test mode (option `-autopilot`)
- [design_test.go](design_test.go), [stimuli_test.go](stimuli_test.go), [strings_test.go](strings_test.go), [subject_test.go](subject_test.go), [richtext_test.go](richtext_test.go): the tests
- [assets](assets): the font
- [build.sh](build.sh) and [Makefile](Makefile): the build commands
- [scripts/check_data.py](scripts/check_data.py): the checks of the data files

The start page of the browser version is in [webapp](../webapp).


## Building

`build.sh` runs on Linux, macOS, and Windows in Git Bash.

Command            | Result
------------------ | ------
`./build.sh`       | the program for this computer
`./build.sh dist`  | the four files of the Releases page, in `dist/`
`./build.sh web`   | the browser version, in `dist-web/`
`./build.sh serve` | the browser version at <http://localhost:8080/>

On GitHub, [release.yml](../.github/workflows/release.yml) publishes a release when a tag such as `v1.0.0` is pushed, and [pages.yml](../.github/workflows/pages.yml) publishes the report and the browser version on every push to `main`.


## Checks

`go test ./...` runs the tests. `python scripts/check_data.py` checks the data files in `data` (it needs pandas): trial counts, accuracy, timeouts and timing.

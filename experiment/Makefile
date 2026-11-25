# Copyright (2026) Thiago Lizier Rissi
#
# make            build ./mental-rotation-preparation for this machine
# make test       go vet + unit tests
# make quick      build, then a short windowed run (subject 1)
# make sheet      build, then the stimulus check sheet
# make check-data run scripts/check_data.py on data/
# make dist       cross-compile every platform into dist/
#
# Windows without make: use ./build.sh from Git Bash.

export GOWORK := off
export CGO_ENABLED := 0

EXE := $(shell go env GOEXE)
BIN := mental-rotation-preparation$(EXE)
PYTHON ?= python3

.PHONY: build test quick sheet check-data dist clean

build:
	go build -trimpath -ldflags="-s -w" -o $(BIN) .

test:
	go vet ./...
	go test ./...

quick: build
	./$(BIN) -w -s 1 -quick

sheet: build
	./$(BIN) -w -s 0 -sheet

check-data:
	$(PYTHON) scripts/check_data.py data

dist:
	./build.sh dist

clean:
	rm -rf dist dist-web mental-rotation-preparation mental-rotation-preparation.exe

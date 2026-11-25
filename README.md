Mental Rotation Preparation
===========================

- Home page: <https://thiago-rissi.github.io/Mental-Rotation-Preparation>
- GitHub repository: <https://github.com/thiago-rissi/Mental-Rotation-Preparation>
- Run the experiment in your browser: <https://thiago-rissi.github.io/Mental-Rotation-Preparation/experiment/>

The aim of this project was to replicate, in a shorter form, an experiment by Cooper and Shepard (1973) on the mental rotation of letters and digits.

In each trial, a letter or a digit is displayed inside a circle, tilted by some angle, and the participant must indicate, by pressing one of two keys, whether it is printed normally or as a mirror image. The response time is recorded. When nothing is known in advance about the character, the response time increases with the angle between the character and its upright position. This is usually taken as evidence that participants rotate a mental image of the character back to upright before comparing it with the normal version stored in memory.

Cooper and Shepard asked whether this rotation can be carried out before the character appears. In some trials, participants were first shown which character was coming and then, for a variable duration, an arrow indicating its orientation. With one second of orientation information, the response time no longer depended on the angle. Our experiment tests whether this result holds with fewer conditions, in a ten-minute session that participants run on their own computer.

**Table of Contents**

- [The original experiment](#the-original-experiment)
- [Design](#design)
- [Stimuli and trial structure](#stimuli-and-trial-structure)
- [Differences from the original study](#differences-from-the-original-study)
- [Running the experiment](#running-the-experiment)
- [Source code](#source-code)
- [Data files](#data-files)
- [Results](#results)
- [References](#references)


## The original experiment

Cooper and Shepard (1973) tested eight participants on six asymmetric characters (G, J, R, 2, 5 and 7), each shown in its normal or mirror-image version at six orientations, from 0 to 300 degrees in steps of 60. The stimuli were presented in a tachistoscope.

Before each test character, participants received one of several kinds of advance information. In condition N, they saw only a blank field for 2000 ms. In the B conditions, they saw the outline of the upcoming character, upright and normal, for 2000 ms, followed by an arrow indicating its orientation for 100, 400, 700 or 1000 ms. The test character appeared as soon as the arrow was removed. In condition C, they saw the outline of the character already rotated to the test orientation for 2000 ms, then a blank field for 1000 ms.

In condition N, the mean response time went from about 570 ms for upright characters to about 1070 ms for characters turned by 180 degrees. The curves of the B conditions became flatter as the arrow duration increased. With 1000 ms of orientation information, the response time was about 400 ms at all orientations, close to condition C. Normal characters were classified 50 to 100 ms faster than mirror-image characters in all conditions.


## Design

The experiment has three conditions, presented in separate blocks. Every participant does all three.

Condition | What is shown before the character
--------- | ----------------------------------
N         | the empty circle, for 1000 ms
B-400     | the outline of the character for 2000 ms, then the arrow for 400 ms
B-1000    | the outline of the character for 2000 ms, then the arrow for 1000 ms

Each condition is crossed with the angular departure of the character from upright (0, 60, 120 or 180 degrees) and with its version (normal or mirror image). For 60 and 120 degrees, half of the characters are turned clockwise and half counterclockwise. Each of the 24 combinations is repeated six times, which gives 144 trials per participant, in six blocks of 24 trials.

The three conditions are presented in one order for the first three blocks and in the reverse order for the last three. This order is one of the six possible orders, chosen from the participant number, so that ten participants cover the six orders almost equally. The response keys are also counterbalanced. Participants with an odd number press J for normal and F for mirror image, and participants with an even number do the reverse.

Each participant first completes six practice trials. The session lasts about ten minutes.

We expect the response time in condition N to increase steeply with the angular departure. If participants use the advance information to prepare a rotated image of the character, the response time in condition B-1000 should depend little on the angle. Condition B-400 should lie between the two, with a cost mainly at the largest angles.


## Stimuli and trial structure

The characters are G, J, R, 2, 5 and 7, as in the original study. They are drawn in near-black on a light background, inside a circular frame. The mirror-image version is obtained by flipping the character horizontally before rotating it. The rotation is done around the centre of the visible character, so that it stays centred in the frame at every angle. The arrow crosses the circle through its centre and points to where the top of the character will be.

After the character appears, the participant has 3000 ms to respond. A wrong answer is followed by the word "Error" for 400 ms, and a missing answer by "Too slow". The next trial starts 500 ms later.

![Successive screens of a trial in each condition](figures/trial-sequences.png)

All durations are counted in screen refreshes. When the program starts, it gets the refresh rate of the screen from the system, or measures it in the browser, and converts every duration into a number of frames. Response times are computed from the time stamp of the screen refresh that displayed the character and the time stamp of the key press, both given by the system rather than measured by the program.


## Differences from the original study

To keep the session short enough for volunteers, we changed several aspects of the original procedure:

- we kept three of the six conditions (N, B-400 and B-1000) and dropped B-100, B-700 and C;
- we used four angular departures instead of six orientations, and presented 60 and 120 degrees in both directions, since the original results were symmetric around 180 degrees;
- the blank field of condition N lasts 1000 ms instead of 2000 ms;
- trials with errors are not repeated, and are excluded from the analysis of response times;
- the response keys are counterbalanced across participants;
- the size of the stimuli is relative to the screen, because the viewing distance is not controlled.


## Running the experiment

The simplest way is to open <https://thiago-rissi.github.io/Mental-Rotation-Preparation/experiment/> in Chrome, Edge or Firefox, on a computer with a keyboard, enter a participant number and click Start. When the experiment is over, the browser downloads a `.zip` file containing the results.

The experiment also runs as a program on Windows, macOS and Linux, either downloaded ready to use from the [Releases page](https://github.com/thiago-rissi/Mental-Rotation-Preparation/releases) or built from the source code. The [README of the experiment](https://github.com/thiago-rissi/Mental-Rotation-Preparation/blob/main/experiment/README.md) explains how to install and run it in each case.


## Source code

The experiment is written in [Go](https://go.dev) with the [goxpyriment](https://chrplr.github.io/goxpyriment/) library. It needs no installation and runs on Windows, macOS and Linux, and in a web browser. The source code is in the folder [experiment](https://github.com/thiago-rissi/Mental-Rotation-Preparation/tree/main/experiment), whose README describes each file.


## Data files

Each session produces a CSV file named `Mental-Rotation-Preparation_sub-<number>_date-<date>.csv`, with one line per trial, and a text file describing the computer and the screen. The files of all participants are stored in the folder [experiment/data](https://github.com/thiago-rissi/Mental-Rotation-Preparation/tree/main/experiment/data). The columns of the CSV file are described in the [README of the experiment](https://github.com/thiago-rissi/Mental-Rotation-Preparation/blob/main/experiment/README.md#data-files).


## Results

To be completed after data collection.


## References

Cooper, L. A., & Shepard, R. N. (1973). The time required to prepare for a rotated stimulus. *Memory & Cognition, 1*(3), 246-250.

Shepard, R. N., & Metzler, J. (1971). Mental rotation of three-dimensional objects. *Science, 171*(3972), 701-703.

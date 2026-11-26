Mental Rotation Preparation
===========================

- Home page: <https://thiago-rissi.github.io/Mental-Rotation-Preparation>
- GitHub repository: <https://github.com/thiago-rissi/Mental-Rotation-Preparation>
- Run the experiment in your browser: <https://thiago-rissi.github.io/Mental-Rotation-Preparation/experiment/>

The aim of this project was to replicate, in a shorter form, an experiment by Cooper and Shepard (1973) on the mental rotation of letters and digits.

In each trial, a letter or a digit is displayed inside a circle, tilted by some angle, and the participant must indicate whether it is printed normally or as a mirror image, while the response time is recorded. When nothing is known in advance about the character, the response time increases with the angle between the character and its upright position, a pattern first shown by Shepard and Metzler (1971) with three-dimensional objects. This is usually taken as evidence that participants rotate a mental image of the character back to upright before comparing it with the normal version stored in memory.

Cooper and Shepard asked whether this rotation can be carried out before the character appears. In some trials, participants were first shown which character was coming and then, for a variable duration, an arrow indicating its orientation. With one second of orientation information, the response time no longer depended on the angle. Our experiment intends to replicate their results, but on a smaller scale, with a ten-minute session.

**Table of Contents**

- [The original experiment](#the-original-experiment)
- [Design](#design)
- [Differences from the original study](#differences-from-the-original-study)
- [Stimuli and trial structure](#stimuli-and-trial-structure)
- [Results](#results)
- [Running the experiment](#running-the-experiment)
- [Source code](#source-code)
- [Data files](#data-files)
- [References](#references)


## The original experiment

Cooper and Shepard (1973) tested eight participants on six asymmetric characters (G, J, R, 2, 5 and 7), each shown in its normal or mirror-image version at six orientations, from 0 to 300 degrees in steps of 60.

Before each test character, participants received one of several kinds of advance information. In condition N, they saw only a blank field for 2000 ms. In the B conditions, they saw the outline of the upcoming character, upright and normal, for 2000 ms, followed by an arrow indicating its orientation for 100, 400, 700 or 1000 ms. The test character appeared as soon as the arrow was removed. In condition C, they saw the outline of the character already rotated to the test orientation for 2000 ms, then a blank field for 1000 ms.

In condition N, the mean response time went from about 570 ms for upright characters to about 1070 ms for characters turned by 180 degrees. The curves of the B conditions became flatter as the arrow duration increased, such that with 1000 ms of orientation information, the response time was about 400 ms at all orientations, close to condition C. This indicates that the time needed to perform the mental rotation for any target angle is no longer than 1000 ms, and this allows the participant to have a mental image of the normal character, already rotated, which helps classify the character shown afterwards.


## Design

In order to shorten the experiment, which took up to six one-hour sessions per participant in the original work, we designed three conditions for our experiment, presented in separate blocks. Every participant does all three:

Condition | What is shown before the character
--------- | ----------------------------------
N         | the empty circle, for 1000 ms
B-400     | the outline of the character for 2000 ms, then the arrow for 400 ms
B-1000    | the outline of the character for 2000 ms, then the arrow for 1000 ms

Each condition is crossed with the angular departure of the character from upright, of 0, 60, 120 or 180 degrees, and with its version (normal or mirror image). For 60 and 120 degrees, half of the characters are turned clockwise and half counterclockwise. Each of the 24 combinations is repeated six times, which gives 144 trials per participant, in six blocks of 24 trials.

The program draws a random number for each participant when it starts, and this number sets, independently and with equal probability, the order of the conditions and the response keys (J for normal and F for mirror image, or the reverse).

Each participant first completes six practice trials, and the session lasts about ten minutes.


## Differences from the original study

In summary, we present the main differences between our experiment and the original one, simplified to provide a shorter session for the participants:

- we kept three of the six conditions (N, B-400 and B-1000) and dropped B-100, B-700 and C;
- we analyse four angular departures instead of six orientations, and present 60 and 120 degrees in both directions, each in half of the trials, considering that the original results were symmetric around 180 degrees;
- the blank field of condition N lasts 1000 ms instead of 2000 ms;
- trials with errors are not repeated, and are excluded from the analysis of response times;
- the response keys are assigned at random to each participant.


## Stimuli and trial structure

The characters are G, J, R, 2, 5 and 7, as in the original study. They are drawn in near-black on a light background, inside a circular frame. The mirror-image version is obtained by flipping the character horizontally before rotating it. The rotation is done around the centre of the visible character, so that it stays centred in the frame at every angle. The arrow crosses the circle through its centre and points to where the top of the character will be.

After the character appears, the participant has 3000 ms to respond. A wrong answer is followed by the word "Error" for 400 ms, and a missing answer by "Too slow". The next trial starts 500 ms later.

![Successive screens of a trial in each condition](website/figures/trial-sequences.png)


## Results

The results below come from example data, not from participants. To test the analysis before data collection, the notebook simulates 12 participants whose mean response times follow Figure 2 of Cooper and Shepard (1973). These results will be replaced by those of the real participants.

The analysis is in [data_analysis.ipynb](https://github.com/thiago-rissi/Mental-Rotation-Preparation/blob/main/data_analysis.ipynb), and its figures and tables are saved in the folder [analysis](https://github.com/thiago-rissi/Mental-Rotation-Preparation/tree/main/analysis). Response times are computed on correct trials only, which leaves 1631 of the 1728 trials.

![Mean response time by orientation, example data](analysis/rt_by_orientation.png)

The figure is the equivalent of Figure 2 of Cooper and Shepard (1973): the mean response time as a function of the orientation of the character, for each condition, with the points at 360 degrees repeating those at 0 degrees.

For the statistical tests, we used the angular departure rather than the orientation, and all three conditions. The mean response times, by angular departure, are:

Condition | 0°   | 60°  | 120° | 180°
--------- | ---- | ---- | ---- | ----
N         | 616  | 702  | 852  | 1128
B-400     | 483  | 512  | 613  | 878
B-1000    | 430  | 435  | 486  | 495

We ran a repeated-measures analysis of variance on the mean response times of each participant, with condition, angular departure and version as factors. Each effect is tested against its own interaction with the participants, and the p-values are corrected for sphericity (Greenhouse-Geisser).

Effect                                     | F     | df     | p      | partial η²
------------------------------------------ | ----- | ------ | ------ | ----------
condition                                  | 431.0 | 2, 22  | < .001 | 0.98
angular departure                          | 473.7 | 3, 33  | < .001 | 0.98
version                                    | 594.9 | 1, 11  | < .001 | 0.98
condition x angular departure              | 131.8 | 6, 66  | < .001 | 0.92
condition x version                        | 2.19  | 2, 22  | .143   | 0.17
angular departure x version                | 0.49  | 3, 33  | .637   | 0.04
condition x angular departure x version    | 1.48  | 6, 66  | .235   | 0.12

To measure the cost of the angle, we fitted for each participant and condition a line to the response time as a function of the angular departure. The mean slope is 2.81 ms per degree in condition N, 2.14 in condition B-400 and 0.41 in condition B-1000. The three slopes differ from each other (paired t-tests with Holm correction, p < .001 for each pair). In these example data, the slope of B-1000 is still above zero (t(11) = 8.68, p < .001), because the simulated means include the small rise of the original B-1000 curve.

Mirror-image characters were classified 69 ms more slowly than normal characters (t(11) = 14.74, p < .001), and this difference does not depend on the condition or the angle.


## Running the experiment

The simplest way is to open <https://thiago-rissi.github.io/Mental-Rotation-Preparation/experiment/> in Chrome, Edge or Firefox, on a computer with a keyboard, and click Start. When the experiment is over, the browser downloads a `.zip` file containing the results.

The experiment also runs as a program on Windows, macOS and Linux, either downloaded ready to use from the [Releases page](https://github.com/thiago-rissi/Mental-Rotation-Preparation/releases) or built from the source code. The [README of the experiment](https://github.com/thiago-rissi/Mental-Rotation-Preparation/blob/main/experiment/README.md) explains how to install and run it in each case.


## Source code

The experiment is written in [Go](https://go.dev) with the [goxpyriment](https://chrplr.github.io/goxpyriment/) library. It needs no installation and runs on Windows, macOS and Linux, and in a web browser. The source code is in the folder [experiment](https://github.com/thiago-rissi/Mental-Rotation-Preparation/tree/main/experiment), whose README describes each file. The files of this site are in [website](https://github.com/thiago-rissi/Mental-Rotation-Preparation/tree/main/website), and the start page of the browser version in [webapp](https://github.com/thiago-rissi/Mental-Rotation-Preparation/tree/main/webapp).


## Data files

Each session produces a CSV file named `Mental-Rotation-Preparation_sub-<number>_date-<date>.csv`, with one line per trial, and a text file describing the computer and the screen. The files of all participants are stored in the folder [experiment/data](https://github.com/thiago-rissi/Mental-Rotation-Preparation/tree/main/experiment/data). The columns of the CSV file are described in the [README of the experiment](https://github.com/thiago-rissi/Mental-Rotation-Preparation/blob/main/experiment/README.md#data-files).


## References

Cooper, L. A., & Shepard, R. N. (1973). The time required to prepare for a rotated stimulus. *Memory & Cognition, 1*(3), 246-250.

Shepard, R. N., & Metzler, J. (1971). Mental rotation of three-dimensional objects. *Science, 171*(3972), 701-703.

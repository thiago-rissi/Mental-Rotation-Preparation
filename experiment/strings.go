// Copyright (2026) Thiago Lizier Rissi

package main

// Every string shown to the participant. English is the reference wording;
// French is a translation of the same screens.
//
// The instructions describe the task and what each block type shows, and say
// that the advance information is always valid. They deliberately never tell
// the participant to rotate anything mentally: whether people prepare on their
// own is part of what the experiment measures.

import (
	"fmt"
	"strconv"
	"strings"
)

// Texts holds the strings of one language.
type Texts struct {
	Lang string

	Welcome       func(m KeyMapping) string
	BlockTypes    string
	Practice      func(nPractice int) string
	AfterPractice func(nBlocks, nTrials int) string
	BlockIntro    func(d ConditionDef, i, n int, m KeyMapping) string
	BlockDone     func(i int) string // first line of the block screen from block 2 on
	Error         string
	TooSlow       string
	DebriefPrompt string
	DebriefHint   string
	End           func(path string) string // desktop
	EndBrowser    string                   // browser build (GOOS=js)
}

// TextsFor returns the strings of a language ("en" or "fr").
func TextsFor(lang string) (Texts, error) {
	switch lang {
	case "en":
		return english(), nil
	case "fr":
		return french(), nil
	}
	return Texts{}, fmt.Errorf("unknown language %q (use en or fr)", lang)
}

// spelled writes small counts as words ("six"), larger ones as digits.
func spelled(n int, words []string) string {
	if n >= 0 && n < len(words) {
		return words[n]
	}
	return strconv.Itoa(n)
}

var (
	enNumbers = []string{"zero", "one", "two", "three", "four", "five", "six",
		"seven", "eight", "nine", "ten", "eleven", "twelve"}
	frNumbers = []string{"zéro", "un", "deux", "trois", "quatre", "cinq", "six",
		"sept", "huit", "neuf", "dix", "onze", "douze"}
)

func up(key string) string { return strings.ToUpper(key) }

func english() Texts {
	t := Texts{Lang: "en"}
	t.Welcome = func(m KeyMapping) string {
		return "Rotated Characters\n\n" +
			"In this experiment you will see letters and digits, one at a time, inside a circle.\n" +
			"Each character may be printed normally or as a mirror image.\n\n" +
			"Press " + boxed(up(m.NormalKey)) + " if it is normal and " + boxed(up(m.BackwardKey)) + " if it is mirrored.\n\n" +
			"Respond as quickly as you can, but try not to make mistakes. " +
			"The session takes about ten minutes.\n\n" +
			"Press the " + bold("space bar") + " to continue."
	}
	t.BlockTypes = "The experiment is divided into blocks of two kinds.\n\n" +
		"In some blocks, the circle stays empty for a moment before the character appears.\n\n" +
		"In the other blocks, you first see the outline of the coming character, upright and " +
		"in its normal form. An arrow then shows where the top of the character will point, " +
		"and the character appears as soon as the arrow is gone. The arrow stays on screen " +
		"longer in some blocks than in others.\n\n" +
		"The outline and the arrow are always correct, but the character itself may still be mirrored.\n\n" +
		"Press the space bar to continue."
	t.Practice = func(n int) string {
		return "You will begin with " + spelled(n, enNumbers) + " practice trials. " +
			"Place your left index finger on F and your right index finger on J.\n\n" +
			"Press the space bar when you are ready."
	}
	t.AfterPractice = func(nBlocks, nTrials int) string {
		return "The practice is over. The experiment itself has " + spelled(nBlocks, enNumbers) +
			" blocks of " + strconv.Itoa(nTrials) + " trials, and you can rest between blocks.\n\n" +
			"Press the space bar to begin."
	}
	t.BlockIntro = func(d ConditionDef, i, n int, m KeyMapping) string {
		var what string
		switch d.Kind {
		case KindN:
			what = "In this block, the circle stays empty for a moment before the character appears."
		case KindB:
			what = "In this block, you will see the outline of the character, then the arrow, then the character."
		case KindC: // not in the default design; no wording was specified for it
			what = "In this block, you will see the outline of the character already tilted, " +
				"then an empty circle, then the character."
		}
		return fmt.Sprintf("Block %d of %d\n\n", i, n) + what + "\n\n" +
			"Press " + up(m.NormalKey) + " for normal and " + up(m.BackwardKey) + " for mirrored.\n\n" +
			"Press the space bar to begin."
	}
	t.BlockDone = func(i int) string {
		return fmt.Sprintf("End of block %d. You may rest for a moment before continuing.", i)
	}
	t.Error = "Error"
	t.TooSlow = "Too slow"
	t.DebriefPrompt = "Before the character appeared, did you try to prepare for it? If so, how?"
	t.DebriefHint = "Type your answer and press Enter to continue."
	t.End = func(path string) string {
		return "The experiment is over. Thank you for taking part.\n\n" +
			"Your responses were saved in\n" +
			path + "\n\n" +
			"This task is adapted from Cooper and Shepard (1973), \"The time required to prepare " +
			"for a rotated stimulus\", Memory & Cognition, 1(3), 246-250.\n\n" +
			"Press any key to close the window."
	}
	t.EndBrowser = "The experiment is over. Thank you for taking part.\n\n" +
		"Your browser has downloaded a file whose name ends in .zip.\n" +
		"This task is adapted from Cooper and Shepard (1973), \"The time required to prepare " +
		"for a rotated stimulus\", Memory & Cognition, 1(3), 246-250.\n\n" +
		"Press any key to finish."
	return t
}

func french() Texts {
	t := Texts{Lang: "fr"}
	t.Welcome = func(m KeyMapping) string {
		return "Caractères tournés\n\n" +
			"Dans cette expérience, vous verrez des lettres et des chiffres, un à la fois, dans un cercle.\n" +
			"Chaque caractère peut être imprimé normalement ou en miroir.\n\n" +
			"Appuyez sur " + boxed(up(m.NormalKey)) + " s'il est normal et sur " + boxed(up(m.BackwardKey)) + " s'il est en miroir.\n\n" +
			"Répondez aussi vite que possible, en essayant de ne pas vous tromper. " +
			"La séance dure environ dix minutes.\n\n" +
			"Appuyez sur la " + bold("barre d'espace") + " pour continuer."
	}
	t.BlockTypes = "L'expérience est divisée en blocs de deux sortes.\n\n" +
		"Dans certains blocs, le cercle reste vide un moment avant que le caractère apparaisse.\n\n" +
		"Dans les autres blocs, vous voyez d'abord le contour du caractère à venir, droit et " +
		"sous sa forme normale. Une flèche montre ensuite où pointera le haut du caractère, " +
		"et le caractère apparaît dès que la flèche a disparu. La flèche reste à l'écran " +
		"plus longtemps dans certains blocs que dans d'autres.\n\n" +
		"Le contour et la flèche sont toujours exacts, mais le caractère lui-même peut être en miroir.\n\n" +
		"Appuyez sur la barre d'espace pour continuer."
	t.Practice = func(n int) string {
		return "Vous allez commencer par " + spelled(n, frNumbers) + " essais d'entraînement. " +
			"Placez votre index gauche sur F et votre index droit sur J.\n\n" +
			"Appuyez sur la barre d'espace quand vous êtes prêt."
	}
	t.AfterPractice = func(nBlocks, nTrials int) string {
		return "L'entraînement est terminé. L'expérience proprement dite comporte " + spelled(nBlocks, frNumbers) +
			" blocs de " + strconv.Itoa(nTrials) + " essais, et vous pourrez vous reposer entre les blocs.\n\n" +
			"Appuyez sur la barre d'espace pour commencer."
	}
	t.BlockIntro = func(d ConditionDef, i, n int, m KeyMapping) string {
		var what string
		switch d.Kind {
		case KindN:
			what = "Dans ce bloc, le cercle reste vide un moment avant que le caractère apparaisse."
		case KindB:
			what = "Dans ce bloc, vous verrez le contour du caractère, puis la flèche, puis le caractère."
		case KindC:
			what = "Dans ce bloc, vous verrez le contour du caractère déjà incliné, " +
				"puis un cercle vide, puis le caractère."
		}
		return fmt.Sprintf("Bloc %d sur %d\n\n", i, n) + what + "\n\n" +
			"Appuyez sur " + up(m.NormalKey) + " pour normal et sur " + up(m.BackwardKey) + " pour miroir.\n\n" +
			"Appuyez sur la barre d'espace pour commencer."
	}
	t.BlockDone = func(i int) string {
		return fmt.Sprintf("Fin du bloc %d. Vous pouvez vous reposer un moment avant de continuer.", i)
	}
	t.Error = "Erreur"
	t.TooSlow = "Trop lent"
	t.DebriefPrompt = "Avant que le caractère n'apparaisse, avez-vous essayé de vous y préparer ? Si oui, comment ?"
	t.DebriefHint = "Tapez votre réponse et appuyez sur Entrée pour continuer."
	t.End = func(path string) string {
		return "L'expérience est terminée. Merci de votre participation.\n\n" +
			"Vos réponses ont été enregistrées dans\n" +
			path + "\n\n" +
			"Cette tâche est adaptée de Cooper et Shepard (1973), « The time required to prepare " +
			"for a rotated stimulus », Memory & Cognition, 1(3), 246-250.\n\n" +
			"Appuyez sur une touche pour fermer la fenêtre."
	}
	t.EndBrowser = "L'expérience est terminée. Merci de votre participation.\n\n" +
		"Votre navigateur a téléchargé un fichier dont le nom se termine par .zip.\n" +
		"Cette tâche est adaptée de Cooper et Shepard (1973), « The time required to prepare " +
		"for a rotated stimulus », Memory & Cognition, 1(3), 246-250.\n\n" +
		"Appuyez sur une touche pour terminer."
	return t
}

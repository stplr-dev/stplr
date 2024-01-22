/*
 * LURE - Linux User REpository
 * Copyright (C) 2023 Elara Musayelyan
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package cliutils

import (
	"context"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"lure.sh/lure/internal/config"
	"lure.sh/lure/internal/db"
	"lure.sh/lure/internal/pager"
	"lure.sh/lure/internal/translations"
	"lure.sh/lure/pkg/loggerctx"
)

// YesNoPrompt asks the user a yes or no question, using def as the default answer
func YesNoPrompt(ctx context.Context, msg string, interactive, def bool) (bool, error) {
	if interactive {
		var answer bool
		err := survey.AskOne(
			&survey.Confirm{
				Message: translations.Translator(ctx).TranslateTo(msg, config.Language(ctx)),
				Default: def,
			},
			&answer,
		)
		return answer, err
	} else {
		return def, nil
	}
}

// PromptViewScript asks the user if they'd like to see a script,
// shows it if they answer yes, then asks if they'd still like to
// continue, and exits if they answer no.
func PromptViewScript(ctx context.Context, script, name, style string, interactive bool) error {
	log := loggerctx.From(ctx)

	if !interactive {
		return nil
	}

	scriptPrompt := translations.Translator(ctx).TranslateTo("Would you like to view the build script for", config.Language(ctx)) + " " + name
	view, err := YesNoPrompt(ctx, scriptPrompt, interactive, false)
	if err != nil {
		return err
	}

	if view {
		err = ShowScript(script, name, style)
		if err != nil {
			return err
		}

		cont, err := YesNoPrompt(ctx, "Would you still like to continue?", interactive, false)
		if err != nil {
			return err
		}

		if !cont {
			log.Fatal(translations.Translator(ctx).TranslateTo("User chose not to continue after reading script", config.Language(ctx))).Send()
		}
	}

	return nil
}

// ShowScript uses the built-in pager to display a script at a
// given path, in the given syntax highlighting style.
func ShowScript(path, name, style string) error {
	scriptFl, err := os.Open(path)
	if err != nil {
		return err
	}
	defer scriptFl.Close()

	str, err := pager.SyntaxHighlightBash(scriptFl, style)
	if err != nil {
		return err
	}

	pgr := pager.New(name, str)
	return pgr.Run()
}

// FlattenPkgs attempts to flatten the a map of slices of packages into a single slice
// of packages by prompting the user if multiple packages match.
func FlattenPkgs(ctx context.Context, found map[string][]db.Package, verb string, interactive bool) []db.Package {
	log := loggerctx.From(ctx)
	var outPkgs []db.Package
	for _, pkgs := range found {
		if len(pkgs) > 1 && interactive {
			choice, err := PkgPrompt(ctx, pkgs, verb, interactive)
			if err != nil {
				log.Fatal("Error prompting for choice of package").Send()
			}
			outPkgs = append(outPkgs, choice)
		} else if len(pkgs) == 1 || !interactive {
			outPkgs = append(outPkgs, pkgs[0])
		}
	}
	return outPkgs
}

// PkgPrompt asks the user to choose between multiple packages.
func PkgPrompt(ctx context.Context, options []db.Package, verb string, interactive bool) (db.Package, error) {
	if !interactive {
		return options[0], nil
	}

	names := make([]string, len(options))
	for i, option := range options {
		names[i] = option.Repository + "/" + option.Name + " " + option.Version
	}

	prompt := &survey.Select{
		Options: names,
		Message: translations.Translator(ctx).TranslateTo("Choose which package to "+verb, config.Language(ctx)),
	}

	var choice int
	err := survey.AskOne(prompt, &choice)
	if err != nil {
		return db.Package{}, err
	}

	return options[choice], nil
}

// ChooseOptDepends asks the user to choose between multiple optional dependencies.
// The user may choose multiple items.
func ChooseOptDepends(ctx context.Context, options []string, verb string, interactive bool) ([]string, error) {
	if !interactive {
		return []string{}, nil
	}

	prompt := &survey.MultiSelect{
		Options: options,
		Message: translations.Translator(ctx).TranslateTo("Choose which optional package(s) to install", config.Language(ctx)),
	}

	var choices []int
	err := survey.AskOne(prompt, &choices)
	if err != nil {
		return nil, err
	}

	out := make([]string, len(choices))
	for i, choiceIndex := range choices {
		out[i], _, _ = strings.Cut(options[choiceIndex], ": ")
	}

	return out, nil
}

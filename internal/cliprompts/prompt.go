// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "LURE - Linux User REpository",
// created by Elara Musayelyan.
// It was later modified as part of "ALR - Any Linux Repository" by the ALR Authors.
// This version has been further modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) Elara Musayelyan (LURE)
// Copyright (C) 2025 The ALR Authors
// Copyright (C) 2025 The Stapler Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cliprompts

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/pkg/staplerfile"

	"go.stplr.dev/stplr/internal/app/tui/pager"
)

func newI18nDefaultKeyMap() *huh.KeyMap {
	return &huh.KeyMap{
		Quit: key.NewBinding(key.WithKeys("ctrl+c")),
		Select: huh.SelectKeyMap{
			Prev:         key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", gotext.Get("back"))),
			Next:         key.NewBinding(key.WithKeys("enter", "tab"), key.WithHelp("enter", gotext.Get("select"))),
			Submit:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", gotext.Get("submit"))),
			Up:           key.NewBinding(key.WithKeys("up", "k", "ctrl+k", "ctrl+p"), key.WithHelp("↑", gotext.Get("up"))),
			Down:         key.NewBinding(key.WithKeys("down", "j", "ctrl+j", "ctrl+n"), key.WithHelp("↓", gotext.Get("down"))),
			Left:         key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("←", gotext.Get("left")), key.WithDisabled()),
			Right:        key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("→", gotext.Get("right")), key.WithDisabled()),
			Filter:       key.NewBinding(key.WithKeys("/"), key.WithHelp("/", gotext.Get("filter"))),
			SetFilter:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", gotext.Get("set filter")), key.WithDisabled()),
			ClearFilter:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", gotext.Get("clear filter")), key.WithDisabled()),
			HalfPageUp:   key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", gotext.Get("½ page up"))),
			HalfPageDown: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", gotext.Get("½ page down"))),
			GotoTop:      key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("g/home", gotext.Get("go to start"))),
			GotoBottom:   key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("G/end", gotext.Get("go to end"))),
		},
		MultiSelect: huh.MultiSelectKeyMap{
			Prev:         key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", gotext.Get("back"))),
			Next:         key.NewBinding(key.WithKeys("enter", "tab"), key.WithHelp("enter", gotext.Get("confirm"))),
			Submit:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", gotext.Get("submit"))),
			Toggle:       key.NewBinding(key.WithKeys("space", "x"), key.WithHelp("x", gotext.Get("toggle"))),
			Up:           key.NewBinding(key.WithKeys("up", "k", "ctrl+p"), key.WithHelp("↑", gotext.Get("up"))),
			Down:         key.NewBinding(key.WithKeys("down", "j", "ctrl+n"), key.WithHelp("↓", gotext.Get("down"))),
			Filter:       key.NewBinding(key.WithKeys("/"), key.WithHelp("/", gotext.Get("filter"))),
			SetFilter:    key.NewBinding(key.WithKeys("enter", "esc"), key.WithHelp("esc", gotext.Get("set filter")), key.WithDisabled()),
			ClearFilter:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", gotext.Get("clear filter")), key.WithDisabled()),
			HalfPageUp:   key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", gotext.Get("½ page up"))),
			HalfPageDown: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", gotext.Get("½ page down"))),
			GotoTop:      key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("g/home", gotext.Get("go to start"))),
			GotoBottom:   key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("G/end", gotext.Get("go to end"))),
			SelectAll:    key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", gotext.Get("select all"))),
			SelectNone:   key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", gotext.Get("select none")), key.WithDisabled()),
		},
		Confirm: huh.ConfirmKeyMap{
			Prev:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", gotext.Get("back"))),
			Next:   key.NewBinding(key.WithKeys("enter", "tab"), key.WithHelp("enter", gotext.Get("next"))),
			Submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", gotext.Get("submit"))),
			Toggle: key.NewBinding(key.WithKeys("h", "l", "right", "left"), key.WithHelp("←/→", gotext.Get("toggle"))),
			Accept: key.NewBinding(key.WithKeys("y", "Y"), key.WithHelp("y", gotext.Get("Yes"))),
			Reject: key.NewBinding(key.WithKeys("n", "N"), key.WithHelp("n", gotext.Get("No"))),
		},
	}
}

// Based on https://github.com/charmbracelet/huh/blob/3b90d9d743964ba8ebc4db221dfa5ccbb8abf888/theme.go#L139
func themeCharm(isDark bool) *huh.Styles {
	t := huh.ThemeBase(isDark)
	lightDark := lipgloss.LightDark(isDark)

	var (
		normalFg = lightDark(lipgloss.Color("252"), lipgloss.Color("235"))
		cream    = lightDark(lipgloss.Color("#FFFDF5"), lipgloss.Color("#FFFDF5"))
		primary  = lipgloss.Color("#00a8a3")
		green    = lightDark(lipgloss.Color("#02BA84"), lipgloss.Color("#02BF87"))
		red      = lightDark(lipgloss.Color("#FF4672"), lipgloss.Color("#ED567A"))
	)

	t.Focused.Base = t.Focused.Base.BorderForeground(lipgloss.Color("238"))
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(primary).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(primary).Bold(true).MarginBottom(1)
	t.Focused.Directory = t.Focused.Directory.Foreground(primary)
	t.Focused.Description = t.Focused.Description.Foreground(lightDark(lipgloss.Color(""), lipgloss.Color("243")))
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(red)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(red)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(primary)
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(primary)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(primary)
	t.Focused.Option = t.Focused.Option.Foreground(normalFg)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(primary)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(green)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("#02CF92"), lipgloss.Color("#02A877"))).SetString("✓ ")
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color(""), lipgloss.Color("243"))).SetString("• ")
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(normalFg)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(cream).Background(primary)
	t.Focused.Next = t.Focused.FocusedButton
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(normalFg).Background(lightDark(lipgloss.Color("237"), lipgloss.Color("252")))

	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(green)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(lightDark(lipgloss.Color("248"), lipgloss.Color("238")))
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(primary)

	t.Blurred = t.Focused
	t.Blurred.Base = t.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description
	return t
}

func wrapIntoHuhForm(f huh.Field) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(f),
	).
		WithKeyMap(newI18nDefaultKeyMap()).
		WithTheme(huh.ThemeFunc(themeCharm))
}

// YesNoPrompt asks the user a yes or no question, using def as the default answer
func YesNoPrompt(ctx context.Context, msg string, interactive, def bool) (bool, error) {
	if interactive {
		answer := def

		err := wrapIntoHuhForm(huh.NewConfirm().
			Title(msg).
			Value(&answer).
			Affirmative(gotext.Get("Yes")).
			Negative(gotext.Get("No")).
			WithButtonAlignment(lipgloss.Left)).Run()
		return answer, err
	} else {
		return def, nil
	}
}

var ErrUserChoseNotContinue = errors.New("user chose not to continue after reading script")

// PromptViewScript asks the user if they'd like to see a script,
// shows it if they answer yes, then asks if they'd still like to
// continue, and exits if they answer no.
func PromptViewScript(ctx context.Context, script, name, style string, interactive bool) error {
	if !interactive {
		return nil
	}

	view, err := YesNoPrompt(ctx, gotext.Get("Would you like to view the build script for %s", name), interactive, false)
	if err != nil {
		return err
	}

	if view {
		err = ShowScript(script, name, style)
		if err != nil {
			return err
		}

		cont, err := YesNoPrompt(ctx, gotext.Get("Would you still like to continue?"), interactive, false)
		if err != nil {
			return err
		}

		if !cont {
			return ErrUserChoseNotContinue
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

	pgr := pager.NewCode(name, str)
	return pgr.Run()
}

// FlattenPkgs attempts to flatten the a map of slices of packages into a single slice
// of packages by prompting the user if multiple packages match.
func FlattenPkgs(ctx context.Context, found map[string][]staplerfile.Package, verb string, interactive bool) ([]staplerfile.Package, error) {
	var outPkgs []staplerfile.Package
	for _, pkgs := range found {
		if len(pkgs) > 1 && interactive {
			choice, err := pkgPrompt(pkgs, verb, interactive)
			if err != nil {
				return nil, fmt.Errorf("error prompting for choice of package: %w", err)
			}
			outPkgs = append(outPkgs, choice)
		} else if len(pkgs) == 1 || !interactive {
			outPkgs = append(outPkgs, pkgs[0])
		}
	}
	return outPkgs, nil
}

// pkgPrompt asks the user to choose between multiple packages.
func pkgPrompt(options []staplerfile.Package, verb string, interactive bool) (staplerfile.Package, error) {
	if !interactive {
		return options[0], nil
	}

	opts := make([]huh.Option[int], len(options))
	for i, option := range options {
		opts[i] = huh.NewOption(option.Repository+"/"+option.Name+" "+option.Version, i)
	}

	var choice int

	err := wrapIntoHuhForm(
		huh.NewSelect[int]().
			Title(gotext.Get("Choose which package to %s", verb)).
			Options(opts...).
			Value(&choice),
	).Run()
	if err != nil {
		return staplerfile.Package{}, err
	}

	return options[choice], nil
}

// ChooseOptDepends asks the user to choose between multiple optional dependencies.
// The user may choose multiple items.
func ChooseOptDepends(ctx context.Context, options []string, verb string, interactive bool) ([]string, error) {
	if !interactive {
		return []string{}, nil
	}

	var choices []string
	err := wrapIntoHuhForm(huh.NewMultiSelect[string]().
		Title(gotext.GetN("Choose which optional package to install", "Choose which optional packages to install", len(options))).
		Options(huh.NewOptions(options...)...).
		Value(&choices),
	).Run()
	if err != nil {
		return nil, err
	}

	out := make([]string, len(choices))
	for i, c := range choices {
		out[i], _, _ = strings.Cut(c, ": ")
	}

	return out, nil
}

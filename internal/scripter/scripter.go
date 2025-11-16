// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
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

package scripter

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/shlex"
	"github.com/goreleaser/nfpm/v2"
	"github.com/goreleaser/nfpm/v2/files"
	"github.com/leonelquinteros/gotext"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/build/common"
	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/internal/shutils/decoder"
	"go.stplr.dev/stplr/internal/shutils/handlers"
	"go.stplr.dev/stplr/internal/shutils/helpers"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/reqprov"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

type BuiltDep struct {
	Name string
	Path string
}

type LocalScriptExecutor struct {
	cfg commonbuild.Config
	out output.Output
}

func NewLocalScriptExecutor(cfg commonbuild.Config, out output.Output) *LocalScriptExecutor {
	return &LocalScriptExecutor{
		cfg,
		out,
	}
}

func (e *LocalScriptExecutor) Read(ctx context.Context, scriptPath string) (*staplerfile.ScriptFile, error) {
	return staplerfile.ReadFromLocal(scriptPath)
}

func (e *LocalScriptExecutor) ParsePackages(
	ctx context.Context,
	file *staplerfile.ScriptFile,
	packages []string,
	info distro.OSRelease,
) (string, []*staplerfile.Package, error) {
	return file.ParseBuildVars(ctx, &info, packages)
}

func (e *LocalScriptExecutor) PrepareDirs(
	ctx context.Context,
	input *commonbuild.BuildInput,
	basePkg string,
) error {
	dirs, err := getDirs(
		e.cfg,
		input.Script,
		basePkg,
	)
	if err != nil {
		return err
	}

	err = prepareDirs(dirs)
	if err != nil {
		return err
	}

	return nil
}

func (e *LocalScriptExecutor) ExecuteSecondPass(
	ctx context.Context,
	input *commonbuild.BuildInput,
	sf *staplerfile.ScriptFile,
	varsOfPackages []*staplerfile.Package,
	repoDeps []string,
	builtDeps []*commonbuild.BuiltDep,
	basePkg string,
) ([]*commonbuild.BuiltDep, error) {
	dirs, err := getDirs(e.cfg, sf.Path(), basePkg)
	if err != nil {
		return nil, err
	}

	env := common.CreateBuildEnvVars(input.OSRelease(), dirs)

	options := []handlers.Option{
		handlers.WithFilter(
			handlers.RestrictSandbox(dirs.SrcDir, dirs.PkgDir),
		),
	}

	disableNet := slices.ContainsFunc(varsOfPackages, func(pkg *staplerfile.Package) bool {
		return pkg.DisableNetwork.Resolved()
	})

	fakeroot, cleanup, err := handlers.SandboxHandler(
		2*time.Second,
		dirs.SrcDir,
		dirs.PkgDir,
		disableNet,
	)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	runner, err := interp.New(
		interp.Env(expand.ListEnviron(env...)),       // Устанавливаем окружение
		interp.StdIO(os.Stdin, os.Stderr, os.Stderr), // Устанавливаем стандартный ввод-вывод
		interp.ReadDirHandler2(handlers.RestrictedReadDir(options...)),
		interp.OpenHandler(handlers.RestrictedOpen(options...)),
		interp.StatHandler(handlers.RestrictedStat(options...)),
		interp.ExecHandlers(func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
			return helpers.Helpers.ExecHandler(fakeroot)
		}), // Обрабатываем выполнение через fakeroot
	)
	if err != nil {
		return nil, err
	}

	err = runner.Run(ctx, sf.File())
	if err != nil {
		return nil, err
	}

	dec := decoder.New(input.OSRelease(), runner)

	// var builtPaths []string

	err = e.ExecuteFunctions(ctx, dirs, dec)
	if err != nil {
		return nil, err
	}

	for _, vars := range varsOfPackages {
		packageName := ""
		if vars.BasePkgName != "" {
			packageName = vars.Name
		}

		pkgFormat := input.PkgFormat()

		funcOut, err := e.ExecutePackageFunctions(
			ctx,
			dec,
			dirs,
			packageName,
		)
		if err != nil {
			return nil, err
		}

		// slog.Info(gotext.Get("Building metadata for package %q", basePkg), "name", basePkg)

		e.out.Info(gotext.Get("Building metadata for package %q", basePkg))

		pkgInfo, err := buildPkgMetadata(
			ctx,
			e.out,
			input,
			vars,
			dirs,
			append(
				repoDeps,
				GetBuiltName(builtDeps)...,
			),
			funcOut.Contents,
		)
		if err != nil {
			return nil, err
		}

		packager, err := nfpm.Get(pkgFormat) // Получаем упаковщик для формата пакета
		if err != nil {
			return nil, err
		}

		pkgName := packager.ConventionalFileName(pkgInfo) // Получаем имя файла пакета
		pkgPath := filepath.Join(dirs.BaseDir, pkgName)   // Определяем путь к пакету

		pkgFile, err := os.Create(pkgPath)
		if err != nil {
			return nil, err
		}

		// utils.WriteProf("/tmp", "stplr-before")
		// google/rpmpack performs in-memory, which is critical for large rpms
		err = packager.Package(pkgInfo, pkgFile)
		if err != nil {
			return nil, err
		}
		// debug.FreeOSMemory()
		// utils.WriteProf("/tmp", "stplr-after")

		builtDeps = append(builtDeps, &commonbuild.BuiltDep{
			Name: vars.Name,
			Path: pkgPath,
		})
	}

	return builtDeps, nil
}

func buildPkgMetadata(
	ctx context.Context,
	out output.Output,
	input interface {
		commonbuild.OsInfoProvider
		commonbuild.BuildOptsProvider
		commonbuild.PkgFormatProvider
		commonbuild.RepositoryProvider
	},
	vars *staplerfile.Package,
	dirs types.Directories,
	deps []string,
	preferedContents *[]string,
) (*nfpm.Info, error) {
	pkgInfo := GetBasePkgInfo(vars, input)
	pkgInfo.Description = vars.Description.Resolved()
	pkgInfo.Platform = "linux"
	pkgInfo.Homepage = vars.Homepage.Resolved()
	pkgInfo.License = strings.Join(vars.Licenses, ", ")
	pkgInfo.Maintainer = vars.Maintainer.Resolved()
	pkgInfo.Overridables = nfpm.Overridables{
		Conflicts: append(vars.Conflicts, vars.Name),
		Replaces:  vars.Replaces,
		Provides:  append(vars.Provides, vars.Name),
		Depends:   deps,
	}
	pkgInfo.Section = vars.Group.Resolved()

	pkgFormat := input.PkgFormat()
	info := input.OSRelease()

	if pkgFormat == "apk" {
		// Alpine отказывается устанавливать пакеты, которые предоставляют сами себя, поэтому удаляем такие элементы
		pkgInfo.Provides = slices.DeleteFunc(pkgInfo.Provides, func(s string) bool {
			return s == pkgInfo.Name
		})
	}

	if pkgFormat == "rpm" {
		pkgInfo.RPM.Group = vars.Group.Resolved()

		if vars.Summary.Resolved() != "" {
			pkgInfo.RPM.Summary = vars.Summary.Resolved()
		} else {
			lines := strings.SplitN(vars.Description.Resolved(), "\n", 2)
			pkgInfo.RPM.Summary = lines[0]
		}
	}

	if vars.Epoch != 0 {
		pkgInfo.Epoch = strconv.FormatUint(uint64(vars.Epoch), 10)
	}

	setScripts(vars, pkgInfo, dirs.ScriptDir)

	if slices.Contains(vars.Architectures, "all") {
		pkgInfo.Arch = "all"
	}

	contents, err := buildContents(vars, dirs, preferedContents)
	if err != nil {
		return nil, err
	}

	normalizeContents(contents)

	if vars.FireJailed.Resolved() {
		contents, err = applyFirejailIntegration(out, vars, dirs, contents)
		if err != nil {
			return nil, err
		}
		pkgInfo.Depends = append(pkgInfo.Depends, "firejail")
	}

	pkgInfo.Contents = contents

	var f *reqprov.ReqProvService
	initFinder := func() error {
		var err error
		f, err = reqprov.New(info, pkgFormat, vars.AutoReqProvMethod.Resolved())
		if err != nil {
			return fmt.Errorf("failed to init provreq: %w", err)
		}
		return nil
	}

	autoProv := vars.AutoProv.Resolved()
	if len(autoProv) == 1 && decoder.IsTruthy(autoProv[0]) {
		if err := initFinder(); err != nil {
			return nil, err
		}
		if err := f.FindProvides(
			ctx,
			out,
			pkgInfo,
			dirs,
			vars.AutoProvSkipList.Resolved(),
			vars.AutoProvFilter.Resolved(),
		); err != nil {
			return nil, fmt.Errorf("failed to find provides: %w", err)
		}
	}

	autoReq := vars.AutoReq.Resolved()
	if len(autoReq) == 1 && decoder.IsTruthy(autoReq[0]) {
		if err := initFinder(); err != nil {
			return nil, err
		}
		if err := f.FindRequires(
			ctx,
			out,
			pkgInfo,
			dirs,
			vars.AutoReqSkipList.Resolved(),
			vars.AutoReqFilter.Resolved(),
		); err != nil {
			return nil, fmt.Errorf("failed to find requires: %w", err)
		}
	}

	return pkgInfo, nil
}

func execFunc(ctx context.Context, out output.Output, d *decoder.Decoder, name string, dirs types.Directories) error {
	fn, ok := d.GetFuncP(name, func(ctx context.Context, r *interp.Runner) error {
		// It should be done via interp.RunnerOption,
		// but due to the issues below, it cannot be done.
		// - https://github.com/mvdan/sh/issues/962
		// - https://github.com/mvdan/sh/issues/1125
		script, err := syntax.NewParser().Parse(strings.NewReader("cd $srcdir"), "")
		if err != nil {
			return err
		}
		return r.Run(ctx, script)
	})
	if ok {
		out.Info(gotext.Get("Executing %s()", name))
		// slog.Info(gotext.Get("Executing %s()", name))
		err := fn(ctx, interp.Dir(dirs.SrcDir))
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *LocalScriptExecutor) ExecuteFunctions(ctx context.Context, dirs types.Directories, dec *decoder.Decoder) error {
	if err := execFunc(ctx, e.out, dec, "prepare", dirs); err != nil {
		return err
	}
	if err := execFunc(ctx, e.out, dec, "build", dirs); err != nil {
		return err
	}

	return nil
}

func (e *LocalScriptExecutor) ExecutePackageFunctions(
	ctx context.Context,
	dec *decoder.Decoder,
	dirs types.Directories,
	packageName string,
) (*commonbuild.FunctionsOutput, error) {
	fOutput := &commonbuild.FunctionsOutput{}
	var packageFuncName string
	var filesFuncName string

	if packageName == "" {
		packageFuncName = "package"
		filesFuncName = "files"
	} else {
		packageFuncName = fmt.Sprintf("package_%s", packageName)
		filesFuncName = fmt.Sprintf("files_%s", packageName)
	}
	if err := execFunc(ctx, e.out, dec, packageFuncName, dirs); err != nil {
		return nil, err
	}

	files, ok := dec.GetFuncP(filesFuncName, func(ctx context.Context, s *interp.Runner) error {
		// It should be done via interp.RunnerOption,
		// but due to the issues below, it cannot be done.
		// - https://github.com/mvdan/sh/issues/962
		// - https://github.com/mvdan/sh/issues/1125
		script, err := syntax.NewParser().Parse(strings.NewReader("cd $pkgdir && shopt -s globstar"), "")
		if err != nil {
			return err
		}
		return s.Run(ctx, script)
	})

	if ok {
		// slog.Info(gotext.Get("Executing %s()", filesFuncName))
		e.out.Info("%s", gotext.Get("Executing %s()", filesFuncName))

		buf := &bytes.Buffer{}

		err := files(
			ctx,
			interp.Dir(dirs.PkgDir),
			interp.StdIO(os.Stdin, buf, os.Stderr),
		)
		if err != nil {
			return nil, err
		}

		contents, err := shlex.Split(buf.String())
		if err != nil {
			return nil, err
		}
		fOutput.Contents = &contents
	}

	return fOutput, nil
}

func normalizeContents(contents []*files.Content) {
	for _, content := range contents {
		content.Destination = filepath.Join("/", content.Destination)
	}
}

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

package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/leonelquinteros/gotext"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/specconv"
	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"

	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/buildns"
	"go.stplr.dev/stplr/internal/constants"
	"go.stplr.dev/stplr/internal/subid"
	"go.stplr.dev/stplr/pkg/types"

	_ "github.com/opencontainers/cgroups/devices"
	"github.com/opencontainers/cgroups/devices/config"
)

type SandboxHandlerInstance struct {
	handler   interp.ExecHandlerFunc
	container *libcontainer.Container
	cleanup   func()
}

func (i *SandboxHandlerInstance) Exec(ctx context.Context, args []string) error {
	return i.handler(ctx, args)
}

func (i *SandboxHandlerInstance) Cleanup() {
	i.cleanup()
}

func (i *SandboxHandlerInstance) Rootfs() string {
	return i.container.Config().Rootfs
}

func SandboxHandler(out output.Output, killTimeout time.Duration, dirs types.Directories, disableNetwork bool) (*SandboxHandlerInstance, error) {
	container, cleanup, err := createContainer(out, dirs, disableNetwork, true)
	if err != nil {
		return nil, err
	}

	err = startInitProcess(container)
	if err != nil && isMountError(err) {
		cleanup()

		slog.Debug("cannot create isolated /proc, retrying bind mount")

		container, cleanup, err = createContainer(out, dirs, disableNetwork, false)
		if err != nil {
			return nil, err
		}

		err = startInitProcess(container)
		if err != nil {
			cleanup()
			return nil, err
		}
	} else if err != nil {
		cleanup()
		return nil, err
	}

	handler := createHandler(container, killTimeout)

	return &SandboxHandlerInstance{
		handler,
		container,
		cleanup,
	}, nil
}

func createContainer(out output.Output, dirs types.Directories, disableNetwork, isolatedProc bool) (*libcontainer.Container, func(), error) {
	rootfsDir, containerDir, err := createTempDirs()
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		os.RemoveAll(rootfsDir)
		os.RemoveAll(containerDir)
	}

	spec, err := buildContainerSpec(out, rootfsDir, dirs, disableNetwork, isolatedProc)
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	libcontainerConfig, err := createLibcontainerConfig(spec)
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	id := fmt.Sprintf("sandbox-%d-%d", os.Getpid(), time.Now().UnixNano())
	container, err := libcontainer.Create(containerDir, id, libcontainerConfig)
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	containerCleanup := func() {
		_ = container.Destroy()
		cleanup()
	}

	return container, containerCleanup, nil
}

func createTempDirs() (string, string, error) {
	rootfsDir, err := os.MkdirTemp("", "stplr-container-rootfs-*")
	if err != nil {
		return "", "", err
	}

	tmpDir := filepath.Join(rootfsDir, "tmp")
	if err := os.Mkdir(tmpDir, 0o1777); err != nil {
		return "", "", err
	}

	containerDir, err := os.MkdirTemp("", "stplr-container-state-*")
	if err != nil {
		os.RemoveAll(rootfsDir)
		return "", "", err
	}

	return rootfsDir, containerDir, nil
}

func buildContainerSpec(out output.Output, rootfsDir string, dirs types.Directories, disableNetwork, isolatedProc bool) (*specs.Spec, error) {
	namespaces := buildNamespaces(disableNetwork)

	// When stplr runs as root, the namespace with the mappings is prepared in
	// advance; otherwise runc maps the ids via newuidmap/newgidmap.
	var uidMappings, gidMappings []specs.LinuxIDMapping
	if path, ok := buildns.InheritedPath(); ok {
		setNamespacePath(namespaces, specs.UserNamespace, path)
	} else {
		var err error
		uidMappings, gidMappings, err = generateMappings(out)
		if err != nil {
			return nil, err
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	mounts := buildMounts(homeDir, dirs, isolatedProc)
	mounts = append(mounts, specs.Mount{
		Destination: "/tmp",
		Type:        "bind",
		Source:      filepath.Join(rootfsDir, "tmp"),
		Options:     []string{"rbind", "rw"},
	})

	spec := &specs.Spec{
		Version: specs.Version,
		Root: &specs.Root{
			Path:     rootfsDir,
			Readonly: false,
		},
		Mounts: mounts,
		Linux: &specs.Linux{
			UIDMappings: uidMappings,
			GIDMappings: gidMappings,
			Namespaces:  namespaces,
			MaskedPaths: []string{"/run"},
		},
	}

	return spec, nil
}

func buildMounts(realHomeDir string, dirs types.Directories, isolatedProc bool) []specs.Mount {
	mounts := []specs.Mount{
		{
			Destination: "/dev",
			Type:        "tmpfs",
			Source:      "tmpfs",
			Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
		},
	}

	if isolatedProc {
		mounts = append(mounts, []specs.Mount{
			{
				Destination: "/proc",
				Type:        "proc",
				Source:      "proc",
			},
		}...)
		slog.Debug("mounting /proc with proc filesystem")
	} else {
		// fallback
		mounts = append(mounts, []specs.Mount{
			{
				Destination: "/proc",
				Type:        "bind",
				Source:      "/proc",
				Options:     []string{"rbind", "ro"},
			},
		}...)
		slog.Debug("mounting /proc with bind mount")
	}

	mounts = append(mounts, buildSystemMounts()...)
	mounts = append(mounts, buildTmpfsMounts(realHomeDir)...)
	mounts = append(mounts, buildWorkspaceMounts(dirs)...)

	return mounts
}

func isMountError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	if strings.Contains(errStr, "operation not permitted") ||
		strings.Contains(errStr, "permission denied") ||
		strings.Contains(errStr, "mount") && strings.Contains(errStr, "proc") {
		return true
	}

	return false
}

func buildSystemMounts() []specs.Mount {
	systemPaths := []string{
		"/bin", "/sbin", "/lib", "/lib64",
		"/usr", "/var", "/etc", "/opt", "/run",
	}

	mounts := make([]specs.Mount, 0, len(systemPaths))
	for _, path := range systemPaths {
		if path == "" {
			continue
		}
		mounts = append(mounts, specs.Mount{
			Destination: path,
			Type:        "rbind",
			Source:      path,
			Options:     []string{"rbind", "rw"},
		})
	}

	return mounts
}

func buildTmpfsMounts(homeDir string) []specs.Mount {
	tmpfsPaths := []string{
		constants.SystemCachePath,
		constants.SocketDirPath,
		homeDir,
		"/var/run",
		"/var/log",
		"/dev/shm",
	}

	mounts := make([]specs.Mount, 0, len(tmpfsPaths))
	for _, path := range tmpfsPaths {
		mounts = append(mounts, specs.Mount{
			Destination: path,
			Type:        "tmpfs",
			Source:      "tmpfs",
			Options:     []string{"nosuid", "noexec", "nodev", "size=16M", "rprivate"},
		})
	}

	return mounts
}

func buildWorkspaceMounts(dirs types.Directories) []specs.Mount {
	execPath, err := os.Executable()
	if err != nil {
		return nil
	}

	mounts := []specs.Mount{
		{
			Destination: execPath,
			Type:        "bind",
			Source:      "/bin/false",
			Options:     []string{"ro", "rbind"},
		},
	}

	for _, path := range []string{dirs.SrcDir, dirs.PkgDir, dirs.HomeDir} {
		mounts = append(mounts, specs.Mount{
			Destination: path,
			Type:        "bind",
			Source:      path,
			Options:     []string{"rbind", "rw"},
		})
	}

	return mounts
}

func setNamespacePath(namespaces []specs.LinuxNamespace, nsType specs.LinuxNamespaceType, path string) {
	for i := range namespaces {
		if namespaces[i].Type == nsType {
			namespaces[i].Path = path
		}
	}
}

func buildNamespaces(disableNetwork bool) []specs.LinuxNamespace {
	namespaces := []specs.LinuxNamespace{
		{Type: specs.PIDNamespace},
		{Type: specs.IPCNamespace},
		{Type: specs.UTSNamespace},
		{Type: specs.MountNamespace},
		{Type: specs.UserNamespace},
		{Type: specs.CgroupNamespace},
	}

	if disableNetwork {
		namespaces = append(namespaces, specs.LinuxNamespace{Type: specs.NetworkNamespace})
	}

	return namespaces
}

func createLibcontainerConfig(spec *specs.Spec) (*configs.Config, error) {
	id := fmt.Sprintf("sandbox-%d-%d", os.Getpid(), time.Now().UnixNano())

	libcontainerConfig, err := specconv.CreateLibcontainerConfig(
		&specconv.CreateOpts{
			CgroupName:       id,
			UseSystemdCgroup: false,
			NoPivotRoot:      false,
			NoNewKeyring:     false,
			Spec:             spec,
			RootlessEUID:     os.Geteuid() != 0,
			RootlessCgroups:  true,
		},
	)
	if err != nil {
		return nil, err
	}

	devices := make([]*config.Rule, 0, len(specconv.AllowedDevices))
	for _, device := range specconv.AllowedDevices {
		devices = append(devices, &device.Rule)
	}

	libcontainerConfig.Devices = specconv.AllowedDevices
	libcontainerConfig.Cgroups.Devices = devices

	return libcontainerConfig, nil
}

func getCapabilities() *configs.Capabilities {
	capList := []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_FSETID",
		"CAP_KILL",
		"CAP_NET_BIND_SERVICE",
		"CAP_SETFCAP",
		"CAP_SETGID",
		"CAP_SETPCAP",
		"CAP_SETUID",
		"CAP_SYS_CHROOT",
	}

	return &configs.Capabilities{
		Bounding:  capList,
		Effective: capList,
		Permitted: capList,
	}
}

func startInitProcess(container *libcontainer.Container) error {
	initProcess := &libcontainer.Process{
		Args:         []string{"/bin/sh", "-c", "sleep infinity"},
		Init:         true,
		Capabilities: getCapabilities(),
	}

	if err := container.Run(initProcess); err != nil {
		return fmt.Errorf("failed to start init process: %w", err)
	}

	return nil
}

func createProcess(path string, args []string, hc interp.HandlerContext) *libcontainer.Process {
	return &libcontainer.Process{
		Args:         append([]string{path}, args[1:]...),
		Env:          execEnv(hc.Env),
		Cwd:          hc.Dir,
		Stdin:        hc.Stdin,
		Stdout:       hc.Stdout,
		Stderr:       hc.Stderr,
		Capabilities: getCapabilities(),
	}
}

func createHandler(container *libcontainer.Container, killTimeout time.Duration) interp.ExecHandlerFunc {
	return func(ctx context.Context, args []string) error {
		hc := interp.HandlerCtx(ctx)

		if len(args) == 0 {
			fmt.Fprintln(hc.Stderr, "no command provided")
			return interp.ExitStatus(127)
		}

		path, err := interp.LookPathDir(hc.Dir, hc.Env, args[0])
		if err != nil {
			fmt.Fprintln(hc.Stderr, err)
			return interp.ExitStatus(127)
		}

		process := createProcess(path, args, hc)

		if err := container.Run(process); err != nil {
			fmt.Fprintf(hc.Stderr, "run failed: %v\n", err)
			return interp.ExitStatus(1)
		}

		return waitForProcess(ctx, process, killTimeout)
	}
}

func waitForProcess(ctx context.Context, process *libcontainer.Process, killTimeout time.Duration) error {
	done := make(chan error, 1)
	go waitProcess(process, done)

	select {
	case <-ctx.Done():
		return handleProcessTermination(process, killTimeout, done, ctx)
	case err := <-done:
		return handleProcessError(err)
	}
}

func waitProcess(process *libcontainer.Process, done chan<- error) {
	state, err := process.Wait()

	switch {
	case err != nil:
		done <- err
	case state != nil && state.ExitCode() != 0:
		done <- fmt.Errorf("exit code %d", state.ExitCode())
	default:
		done <- nil
	}
}

func handleProcessTermination(process *libcontainer.Process, killTimeout time.Duration, done <-chan error, ctx context.Context) error {
	_ = process.Signal(syscall.SIGTERM)

	select {
	case <-time.After(killTimeout):
		_ = process.Signal(syscall.SIGKILL)
	case <-done: //nolint:errcheck // the process is being terminated, its exit error is irrelevant
	}

	return ctx.Err()
}

func handleProcessError(err error) error {
	if err == nil {
		return nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			// Clamp to 0–255 to avoid integer overflow (POSIX exit codes are 1 byte)
			//gosec:disable G115
			return interp.ExitStatus(uint8((128 + int(status.ExitStatus())) & 0xFF))
		}
	}

	return interp.ExitStatus(1)
}

// execEnv was extracted from github.com/mvdan/sh/interp/vars.go
func execEnv(env expand.Environ) []string {
	list := make([]string, 0, 64)
	env.Each(func(name string, vr expand.Variable) bool {
		if !vr.IsSet() {
			// If a variable is set globally but unset in the
			// runner, we need to ensure it's not part of the final
			// list. Seems like zeroing the element is enough.
			// This is a linear search, but this scenario should be
			// rare, and the number of variables shouldn't be large.
			for i, kv := range list {
				if strings.HasPrefix(kv, name+"=") {
					list[i] = ""
				}
			}
		}
		if vr.Exported && vr.Kind == expand.String {
			list = append(list, name+"="+vr.String())
		}
		return true
	})
	return list
}

func generateMappings(out output.Output) ([]specs.LinuxIDMapping, []specs.LinuxIDMapping, error) {
	u, err := user.Current()
	if err != nil {
		return nil, nil, fmt.Errorf("getting current user: %w", err)
	}

	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing uid %q: %w", u.Uid, err)
	}

	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing gid %q: %w", u.Gid, err)
	}

	subUIDs, err := ownedSubIDs(subid.UIDFile, u)
	if err != nil {
		return nil, nil, err
	}

	// Owners in /etc/subgid are users too, so match them by uid.
	subGIDs, err := ownedSubIDs(subid.GIDFile, u)
	if err != nil {
		return nil, nil, err
	}

	if len(subUIDs) == 0 || len(subGIDs) == 0 {
		warnOnce(out, gotext.Get("You have no subordinate ids, so the build sandbox can't change file ownership. See https://stplr.dev/docs/guide#build-without-root"))
		subUIDs, subGIDs = nil, nil
	} else if err := checkIDMapHelpers(); err != nil {
		warnOnce(out, gotext.Get("Can't use newuidmap and newgidmap (%s), so the build sandbox can't change file ownership. See https://stplr.dev/docs/guide#build-without-root", err))
		subUIDs, subGIDs = nil, nil
	}

	return toSpecMappings(subid.Mappings(uint32(uid), subUIDs)),
		toSpecMappings(subid.Mappings(uint32(gid), subGIDs)),
		nil
}

func checkIDMapHelpers() error {
	if v, err := unix.PrctlRetInt(unix.PR_GET_NO_NEW_PRIVS, 0, 0, 0, 0); err == nil && v == 1 {
		return errors.New("no_new_privs is set, newuidmap and newgidmap can't gain privileges")
	}
	for _, name := range []string{"newuidmap", "newgidmap"} {
		if err := checkIDMapHelper(name); err != nil {
			return err
		}
	}
	return nil
}

// checkIDMapHelper reports whether the current user can run the helper and it
// can gain the privileges to write a mapping with several ranges, either
// through the setuid bit or file capabilities.
func checkIDMapHelper(name string) error {
	path, err := exec.LookPath(name)
	if err != nil {
		return err
	}
	if err := unix.Access(path, unix.X_OK); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	var st unix.Stat_t
	if err := unix.Stat(path, &st); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if st.Mode&unix.S_ISUID != 0 && st.Uid == 0 {
		return nil
	}
	if _, err := unix.Getxattr(path, "security.capability", nil); err == nil {
		return nil
	}
	return fmt.Errorf("%s has neither the setuid bit nor file capabilities", path)
}

func ownedSubIDs(path string, u *user.User) ([]subid.Range, error) {
	ranges, err := subid.Owned(path, u.Username, u.Uid)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return ranges, nil
}

// The sandbox is created for every script run, so the warning is shown once.
var mappingWarning sync.Once

func warnOnce(out output.Output, msg string) {
	mappingWarning.Do(func() { out.Warn("%s", msg) })
}

func toSpecMappings(maps []subid.IDMap) []specs.LinuxIDMapping {
	out := make([]specs.LinuxIDMapping, 0, len(maps))
	for _, m := range maps {
		out = append(out, specs.LinuxIDMapping{ContainerID: m.ContainerID, HostID: m.HostID, Size: m.Size})
	}
	return out
}

// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
// Copyright (C) 2026 The Stapler Authors
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

// Package buildns prepares the user namespace for the build sandbox when
// stplr runs as root.
//
// The sandbox runs as the builder user with no_new_privs set, so it can't use
// the setuid newuidmap/newgidmap to map its subordinate ids. Instead, root
// creates a namespace owned by the builder, writes the mappings itself and
// hands the namespace path over; the sandbox then joins it.
package buildns

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/subid"
)

// HolderSubcommand must be handled before the CLI starts, see [RunHolder].
const HolderSubcommand = "_builder_userns_holder"

const (
	holderWaitArg = "wait"
	envPath       = "STPLR_BUILDER_USERNS"
)

// Namespace is a user namespace kept alive by a holder process.
type Namespace struct {
	// Path is /proc/<pid>/ns/user of the holder. runc reads the mappings
	// from /proc/<pid>/*id_map only for paths of exactly this form.
	Path string

	stdin io.Closer
	cmd   *exec.Cmd
}

// Close stops the holder; the namespace is gone once nothing uses it.
func (n *Namespace) Close() error {
	_ = n.stdin.Close()
	return n.cmd.Wait()
}

// Create returns a user namespace owned by the builder, with the builder
// mapped to root and its subordinate ranges mapped after it.
func Create(username string, out output.Output) (*Namespace, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return nil, err
	}
	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parsing uid %q: %w", u.Uid, err)
	}
	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parsing gid %q: %w", u.Gid, err)
	}

	// Owners in /etc/subgid are users too, so both files are matched by uid.
	subUIDs, err := subid.Owned(subid.UIDFile, u.Username, u.Uid)
	if err != nil {
		return nil, err
	}
	subGIDs, err := subid.Owned(subid.GIDFile, u.Username, u.Uid)
	if err != nil {
		return nil, err
	}
	if len(subUIDs) == 0 || len(subGIDs) == 0 {
		out.Warn("%s", gotext.Get("The build user has no subordinate ids, so the build sandbox can't change file ownership. Run as root: stplr setup"))
	}
	uidMap := subid.Format(subid.Mappings(uint32(uid), subUIDs))
	gidMap := subid.Format(subid.Mappings(uint32(gid), subGIDs))

	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(exe, HolderSubcommand)
	cmd.Env = []string{}
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid), Groups: []uint32{}},
		Pdeathsig:  syscall.SIGKILL,
	}

	// The holders live until stdin is closed.
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting userns holder: %w", err)
	}
	ns := &Namespace{stdin: stdin, cmd: cmd}

	pid, err := readPid(stdout)
	if err != nil {
		_ = ns.Close()
		return nil, err
	}
	ns.Path = fmt.Sprintf("/proc/%d/ns/user", pid)

	// The kernel accepts only one write per map file.
	if err := os.WriteFile(fmt.Sprintf("/proc/%d/uid_map", pid), []byte(uidMap), 0); err != nil {
		_ = ns.Close()
		return nil, fmt.Errorf("writing uid_map: %w", err)
	}
	if err := os.WriteFile(fmt.Sprintf("/proc/%d/gid_map", pid), []byte(gidMap), 0); err != nil {
		_ = ns.Close()
		return nil, fmt.Errorf("writing gid_map: %w", err)
	}

	return ns, nil
}

func readPid(r io.Reader) (int, error) {
	line, err := bufio.NewReader(r).ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("reading userns holder pid: %w", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return 0, fmt.Errorf("parsing userns holder pid %q: %w", line, err)
	}
	return pid, nil
}

// RunHolder implements [HolderSubcommand]. Running as the builder, it starts
// a child in a new user namespace (the namespace has to be created by the
// builder to be owned by it) and prints the child pid. The child only waits
// for stdin to close.
func RunHolder(args []string) {
	if len(args) > 0 && args[0] == holderWaitArg {
		_, _ = io.Copy(io.Discard, os.Stdin)
		os.Exit(0)
	}

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	child := exec.Command(exe, HolderSubcommand, holderWaitArg)
	child.Env = []string{}
	child.Stdin = os.Stdin
	child.Stderr = os.Stderr
	child.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER,
		Pdeathsig:  syscall.SIGKILL,
	}
	if err := child.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "creating user namespace:", err)
		os.Exit(1)
	}

	fmt.Println(child.Process.Pid)
	_ = child.Wait()
	os.Exit(0)
}

// Pass tells cmd to run its sandboxes in ns.
func Pass(cmd *exec.Cmd, ns *Namespace) {
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", envPath, ns.Path))
}

// InheritedPath returns the namespace path passed with [Pass], if any.
func InheritedPath() (string, bool) {
	path := os.Getenv(envPath)
	return path, path != ""
}

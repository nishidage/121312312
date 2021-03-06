//go:build darwin || dragonfly || linux || openbsd || freebsd || solaris
// +build darwin dragonfly linux openbsd freebsd solaris

/*
Copyright 2020 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package exechelper

import (
	"syscall"
)

func getSysProcAttr(tty bool, origin *syscall.SysProcAttr) *syscall.SysProcAttr {
	if tty {
		// if using tty in unix, github.com/creack/pty will Setsid, and if we
		// Setpgid, will fail the process creation
		//
		// https://github.com/creack/pty/issues/35#issuecomment-147947212
		// do not Setpgid if already Setsid
		if origin != nil {
			origin.Setpgid = false
			origin.Pgid = 0
		}
	}

	return origin
}

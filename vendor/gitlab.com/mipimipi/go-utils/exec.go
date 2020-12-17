// SPDX-FileCopyrightText: 2018-2020 Michael Picht <mipi@fsfe.org>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

// ExecQuiet executes a shell command. It receives the environment, the name of
// the command and its arguments.
// It returns stdout and stderr as byte array
func ExecQuiet(env []string, name string, args ...string) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer

	// create the command
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// run the command
	err := cmd.Run()

	return stdout.Bytes(), stderr.Bytes(), err
}

// ExecVerbose executes a shell command and prints its stdout and stderr
// simultaneously.
// It receives the environment, the command name and its arguments.
func ExecVerbose(env []string, name string, args ...string) (err error) {
	// create command
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	// start command
	err = cmd.Start()
	if err != nil {
		return
	}

	var wg sync.WaitGroup

	// helper function to print output asynchronously
	out := func(pipe io.ReadCloser) {
		sc := bufio.NewScanner(pipe)
		sc.Split(bufio.ScanLines)
		for sc.Scan() {
			fmt.Printf("%s", sc.Bytes())
		}
		wg.Done()
	}
	// continuously print output ...
	wg.Add(2)
	go out(stdoutPipe)
	go out(stderrPipe)

	// ... and wait
	wg.Wait()

	// wait for command to be done
	return cmd.Wait()
}

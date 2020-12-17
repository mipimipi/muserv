// SPDX-FileCopyrightText: 2018-2020 Michael Picht <mipi@fsfe.org>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package utils

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"reflect"
	"time"
	"unsafe"
)

// Contains checks if the array a contains the element e.
// inspired by: https://stackoverflow.com/questions/10485743/contains-method-for-a-slice
func Contains(a interface{}, e interface{}) bool {
	arr := reflect.ValueOf(a)

	if arr.Kind() == reflect.Slice {
		for i := 0; i < arr.Len(); i++ {
			// XXX - panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if arr.Index(i).Interface() == e {
				return true
			}
		}
	}

	return false
}

// HashUint64 calculates the FNV hash of the input as uint64
func HashUint64(format string, a ...interface{}) uint64 {
	h := fnv.New64a()
	fmt.Fprintf(h, format, a...)
	return uint64(h.Sum64())
}

// ProgressStr shows and moves a bar '...' on the command line. It can be used
// to show that an activity is ongoing. The parameter 'interval' steers the
// refresh rate (in milli seconds). The text in 'msg' is displayed in form of
// '...'. The progress bar is stopped by sending an empty struct to the
// returned channel:
//	 chan <- struct{}{}
//	 close(chan)
func ProgressStr(msg string, interval time.Duration) (chan<- struct{}, <-chan struct{}) {
	// create channel to receive stop signal
	stop := make(chan struct{})

	// create channel to send stop confirmation
	confirm := make(chan struct{})

	go func() {
		var (
			ticker  = time.NewTicker(interval * time.Millisecond)
			bar     = "   ...  "
			i       = 5
			isFirst = true
			ticked  = false
		)

		for {
			select {
			case <-ticker.C:
				// at the very first tick, the output switches to the next row.
				// At all subsequent ticks, the output is printed into that
				// same row.
				if isFirst {
					fmt.Println()
					isFirst = false
				}
				// print message and progress indicator
				fmt.Printf("\r%s %s ", msg, bar[i:i+3])
				// increase progress indicator counter for next tick
				if i--; i < 0 {
					i = 5
				}
				// ticker has ticked: set flag accordingly
				ticked = true
			case <-stop:
				// stop ticker ...
				ticker.Stop()
				// if the ticker had displayed at least once, move to next row
				if ticked {
					fmt.Println()
				}
				// send stop confirmation
				confirm <- struct{}{}
				close(confirm)
				// and return
				return
			}
		}
	}()

	return stop, confirm
}

// random string generation.  Adopted from:
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

// RandomString generates a random string og length n consisting of characters
// a-z, A-Z and 0-9.
func RandomString(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// UserOK print the message s followed by " (Y/n)?" on stdout and askes the
// user to press either Y (to continue) or n (to stop). Y is treated as
// default. I.e. if the user only presses return, that's interpreted as if
// he has pressed Y.
func UserOK(s string) bool {
	var input string

	for {
		fmt.Printf("\r%s (Y/n)? ", s)
		if _, err := fmt.Scanln(&input); err != nil {
			if err.Error() != "unexpected newline" {
				return false
			}
			input = "Y"
		}
		switch {
		case input == "Y":
			return true
		case input == "n":
			return false
		}
		fmt.Println()
	}
}

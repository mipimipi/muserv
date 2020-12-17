// SPDX-FileCopyrightText: 2018-2020 Michael Picht <mipi@fsfe.org>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package utils

import (
	"math/rand"
	"time"
)

// RandomNap executes a sleep for a random number of 0 to dur milliseconds
func RandomNap(dur time.Duration) {
	time.Sleep(dur * time.Millisecond * time.Duration(rand.Float64()))
}

// SplitDuration disaggregates a duration and returns it splitted into hours,
// minutes, seconds and nanoseconds
func SplitDuration(d time.Duration) map[time.Duration]time.Duration {
	var (
		out  = make(map[time.Duration]time.Duration)
		cmps = []time.Duration{time.Hour, time.Minute, time.Second, time.Nanosecond}
	)

	for _, cmp := range cmps {
		out[cmp] = d / cmp
		d -= out[cmp] * cmp
	}

	return out
}

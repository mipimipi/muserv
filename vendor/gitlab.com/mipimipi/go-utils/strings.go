// SPDX-FileCopyrightText: 2018-2020 Michael Picht <mipi@fsfe.org>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package utils

import "strings"

// SplitMulti slices s into all substrings separated by any character of sep
// and returns a slice of the substrings between those separators.
// If s does not contain any character of sep and sep is not empty, SplitMulti
// returns a slice of length 1 whose only element is s.
// If sep is empty, SplitMulti splits after each UTF-8 sequence. If both s and
// sep are empty, SplitMulti returns an empty slice.
func SplitMulti(s string, sep string) []string {
	var a []string

	// handle special cases: if sep is empty ...
	if len(sep) == 0 {
		//... and if s is empty: return an empty slice
		if len(s) == 0 {
			return a
		}
		// ... else split after each character
		return strings.Split(s, "")
	}

	// split s by the characters of sep
	for i, j := -1, 0; j <= len(s); j++ {
		if j == len(s) || strings.Contains(sep, string(s[j])) {
			if i+1 > j-1 {
				a = append(a, "")
			} else {
				a = append(a, s[i+1:j])
			}
			i = j
		}
	}

	// if s does not contain any character of sep: return a slice that only
	// contains s
	if len(a) == 0 {
		a = append(a, s)
	}

	return a
}

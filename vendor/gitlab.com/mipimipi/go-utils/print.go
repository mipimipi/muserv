// SPDX-FileCopyrightText: 2018-2020 Michael Picht <mipi@fsfe.org>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package utils

import (
	"fmt"

	"github.com/logrusorgru/aurora"
)

// PrintPlainln prints a plain text
func PrintPlainln(format string, a ...interface{}) {
	_, _ = fmt.Printf("    %s\n", aurora.Bold(fmt.Sprintf(format, a...)))
}

// PrintInfoln prints an info text
func PrintInfoln(format string, a ...interface{}) {
	_, _ = fmt.Printf("%s %s\n", aurora.Bold(aurora.BrightGreen("==>")), aurora.Bold(fmt.Sprintf(format, a...)))
}

// PrintMsgln prints a message
func PrintMsgln(format string, a ...interface{}) {
	_, _ = fmt.Printf("%s %s\n", aurora.Bold(aurora.BrightCyan("==>")), aurora.Bold(fmt.Sprintf(format, a...)))
}

// PrintWarnln prints a warning
func PrintWarnln(format string, a ...interface{}) {
	_, _ = fmt.Printf("%s %s\n", aurora.Bold(aurora.BrightYellow("==> WARNING:")), aurora.Bold(fmt.Sprintf(format, a...)))
}

// PrintErrorln prints an error
func PrintErrorln(format string, a ...interface{}) {
	_, _ = fmt.Printf("%s %s\n", aurora.Bold(aurora.BrightRed("==> ERROR:")), aurora.Bold(fmt.Sprintf(format, a...)))
}

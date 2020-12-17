// SPDX-FileCopyrightText: 2018-2020 Michael Picht <mipi@fsfe.org>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package utils

import (
	"strings"
	"syscall"
)

// SysInfo contains all fields of syscall.Utsname but as type string instead
// of [65]int8
type SysInfo struct {
	Nodename   string
	Release    string
	Sysname    string
	Version    string
	Machine    string
	Domainname string
}

// Uname retrieves system information via syscall.Uname and converts the
// resulting struct of [65]int8 values in a struct of strings
func Uname() (SysInfo, error) {
	var sysinfo syscall.Utsname
	if err := syscall.Uname(&sysinfo); err != nil {
		return SysInfo{}, err
	}
	return SysInfo{
		Nodename:   arrayToString(sysinfo.Nodename),
		Release:    arrayToString(sysinfo.Release),
		Sysname:    arrayToString(sysinfo.Sysname),
		Version:    arrayToString(sysinfo.Version),
		Machine:    arrayToString(sysinfo.Machine),
		Domainname: arrayToString(sysinfo.Domainname),
	}, nil
}

// arrayToString is a helper function to convert an [65]int8 value into a
// string
func arrayToString(x [65]int8) string {
	var buf [65]byte
	for i, b := range x {
		buf[i] = byte(b)
	}
	str := string(buf[:])
	if i := strings.Index(str, "\x00"); i != -1 {
		str = str[:i]
	}
	return str
}

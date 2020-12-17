// SPDX-FileCopyrightText: 2018-2020 Michael Picht <mipi@fsfe.org>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package file

// Package file contains practical and handy functions for dealing with files
// and directories.

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// Info extends the standard interface os.FileInfo
type Info interface {
	os.FileInfo
	Path() string // get complete file name
}

// info is an internal helper structire that implements FileInfo
type info struct {
	os.FileInfo
	path string // complete name of the file
}

// Path implements the Path() method, so that fileinfo implements FileInfo
func (i info) Path() string { return i.path }

// create FileInfo from os.FileInfo and a path
func newInfo(fi os.FileInfo, p string) Info { return info{fi, p} }

// ValidPropagate defines if validity shall be propagated to sub
// directories. Needed for Find()
type ValidPropagate int

// constants for Find(): propagation of validity to sub directories
const (
	NoneFromSuper    ValidPropagate = iota // no propagation to sub directories
	ValidFromSuper                         // propagate valid=true to sub directories
	InvalidFromSuper                       // propagate valid=false to sub directories
)

// CheckMkdir checks if a directory exists. If it doesn't exist, it's
// being created
func CheckMkdir(path string, perm os.FileMode) error {
	var (
		b   bool
		err error
	)
	if b, err = Exists(path); err != nil {
		return err
	}
	if b {
		return nil
	}
	return os.MkdirAll(path, perm)
}

// Copy copies srcFn to dstFn. Prequisite is, that srcFn and (if existing)
// dstFn are regular files (i.e. no devices etc.). In case both files are the
// same, nothing is done. In case, dstFn is already existing it is overwritten.
func Copy(srcFn, dstFn string) error {
	var (
		err    error
		exists bool
		srcFi  os.FileInfo
		dstFi  os.FileInfo
		src    *os.File
		dst    *os.File
	)

	// check if source file exists
	exists, err = Exists(srcFn)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("CopyFile: Source file %s doesn't exist", srcFn)
	}

	// get fileinfo for source file
	srcFi, err = os.Stat(srcFn)
	if err != nil {
		return err
	}

	// make sure that source file is a regular file
	if !srcFi.Mode().IsRegular() {
		return fmt.Errorf("CopyFile: Non-regular source file %s (%q)", srcFi.Name(), srcFi.Mode().String())
	}

	// determine existence of dest file
	exists, err = Exists(dstFn)
	if err != nil {
		return err
	}
	if exists {
		// get fileinfo for dest file
		dstFi, err = os.Stat(dstFn)
		if err != nil {
			return err
		}
		// make sure that dest file is a regular file
		if !(dstFi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: Non-regular destination file %s (%q)", dstFi.Name(), dstFi.Mode().String())
		}
		// if source and dest file are the same: do nothing
		if os.SameFile(srcFi, dstFi) {
			return nil
		}
	}

	// open source file and defer closing
	src, err = os.Open(srcFn)
	if err != nil {
		return err
	}
	defer func() { err = src.Close() }()

	// create/open dest file and defer closing
	dst, err = os.OpenFile(dstFn, os.O_WRONLY|os.O_CREATE, srcFi.Mode())
	if err != nil {
		return err
	}
	defer func() { err = dst.Close() }()

	// copy content of source file to dest file
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	// flush dest file
	return dst.Sync()
}

// EscapePattern escapes special characters in pattern strings for usage in
// filepath.Glob() or filepath.Match()
// See: https://godoc.org/path/filepath#Match
func EscapePattern(s string) string {
	special := [...]string{"[", "?", "*"}
	for _, sp := range special {
		s = strings.Replace(s, sp, "\\"+sp, -1)
	}
	return s
}

// IsDir returns true is path exists and is a directory
func IsDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fi.Mode().IsDir(), nil
}

// Exists returns true if path exists, otherwise false
func Exists(path string) (bool, error) {
	exists, _, err := ExistsInfo(path)
	return exists, err
}

// ExistsInfo returns true if path exists, otherwise false. In addition to
// Exists it also return file.Info
func ExistsInfo(path string) (bool, Info, error) {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, fmt.Errorf("Existence of file '%s' couldn't be determined: %v", path, err)
	}
	return true, newInfo(fi, path), nil
}

// Find traverses directory trees to find files that fulfill a certain filter
// condition. It starts at a list of root directories. The condition must be
// implemented in a function, which is passed to Find as a parameter. This
// condition returns two boolean value. The first one determines if a certain
// entry is valid (i.e. fulfills the actual filter condition), the second
// determines (only in case of a directory) if the filter result (i.e. the first
// boolean) shall be propagated to the entries of the directory.
// numWorkers is the number of concurrent Go routines that Find uses.
// Find returns the absolute paths of the files and directories that fulfill the
// filter condition as string array.
// This function is inspired by the Concurrent Directory Traversal from the
// book "The Go Programming Language" by Alan A. A. Donovan & Brian W.
// Kernighan.
// See: https://github.com/adonovan/gopl.io/blob/master/ch8/du4/main.go
func Find(roots []Info, filter func(Info, ValidPropagate) (bool, ValidPropagate), numWorkers int) (files *[]*Info) {
	var (
		filterDescendDir func(Info, ValidPropagate) // func needs to be declared here since it calls itself recursively
		wg               sync.WaitGroup             // waiting group for the concurrent traversal
		addfiles         = make(chan *Info)         // channel to collect results
	)

	// allocate result array
	files = new([]*Info)

	// create buffered channel, used as semaphore to restrict the number of Go routines
	sema := make(chan struct{}, numWorkers)
	defer close(sema)

	// function to retrieve the entries of a directory
	entries := func(dir Info) []os.FileInfo {
		// send to limited buffer and retrieve from buffer at the end
		sema <- struct{}{}
		defer func() { <-sema }()

		// retrieve directory entries
		entrs, err := ioutil.ReadDir(dir.Path())
		if err != nil {
			return nil
		}
		return entrs
	}

	// function to check if a directory is relevant. If yes, descend into that
	// directory
	filterDescendDir = func(info Info, vp ValidPropagate) {
		defer wg.Done()

		valid, vpSub := filter(info, vp)
		if valid {
			addfiles <- &info
		}
		// descend into directory
		if vpSub != InvalidFromSuper {
			// loop at the entries of dir
			for _, entr := range entries(info) {
				// distinguish between dirs and files. Both need to fulfill the
				// filter condition. If they do, the entry is send to the
				// results channel
				if entr.IsDir() {
					// filter entr and descend
					wg.Add(1)
					go filterDescendDir(newInfo(entr, filepath.Join(info.Path(), entr.Name())), vpSub)
				} else {
					// only regular files are relevant
					if !entr.Mode().IsRegular() {
						continue
					}
					// filter and add entry to files
					infoSub := newInfo(entr, filepath.Join(info.Path(), entr.Name()))
					if valid, _ := filter(infoSub, vpSub); valid {
						addfiles <- &infoSub
					}
				}
			}
		}
	}

	// start traversal for the root directories
	for _, root := range roots {
		// verify that root is a directory
		if !root.IsDir() {
			continue
		}
		// filter root and descend
		wg.Add(1)
		go filterDescendDir(root, NoneFromSuper)
	}

	// wait for traversals to be finalized
	go func() {
		wg.Wait()
		close(addfiles)
	}()

	// collect results into results array
	for f := range addfiles {
		*files = append(*files, f)
	}

	return files
}

// GlobOr execute filepath.Glob on a list of patterns. The result is a list of
// files that match at least one of the patterns in the list. GlobOr returns an
// error if at least one of the calls of filepath.Glob returned an error
func GlobOr(patterns []string) (matches []string, err error) {
	for _, pattern := range patterns {
		m, err := filepath.Glob(pattern)
		if err != nil {
			return []string{}, err
		}
		matches = append(matches, m...)
	}
	return matches, nil
}

// IsEmpty returns true if file or directory is empty, otherwise false
func IsEmpty(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fi.IsDir() {
		entries, err := ioutil.ReadDir(path)
		if err != nil {
			return false, err
		}
		return (len(entries) == 0), nil
	}
	if fi.Mode().IsRegular() {
		return (fi.Size() == 0), nil
	}
	return false, nil
}

// MD5 calculates the md5 hash of a file
func MD5(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// MkdirAll creates a directory named path, along with any necessary parents.
// In this regard, it behaves like the standard os.MkdirAll. In contrast to
// this, it doesn't complain if path already exists.
func MkdirAll(path string, perm os.FileMode) error {
	var err error

	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, perm)
	}
	return err
}

// PathRelCopy determines first a relative path that is lexically equivalent to
// path when joined to srcBase with an intervening separator. If this is
// successful, it returns a path joined from dstBase, a separator and the
// relative path from the previous step.
func PathRelCopy(srcBase, path, dstBase string) (string, error) {
	var (
		rel string
		err error
	)

	// determine the relative path using filepath.Rel()
	// See: https://godoc.org/path/filepath#Rel
	if rel, err = filepath.Rel(srcBase, path); err != nil {
		return "", fmt.Errorf("PathRelCopy: %v", err)
	}

	// if the relative path is empty: Just return dstBase
	if rel == "" {
		return dstBase, nil
	}

	// other wise join dstBase and the relative path
	return filepath.Join(dstBase, rel), nil
}

// PathTrunk returns the file path without the file extension.
// E.g. Trunk("/home/test/abc.mp3") return "/home/test/abc"
func PathTrunk(p string) string {
	return p[0 : len(p)-len(path.Ext(p))]
}

// RenameAll moves all files that match pattern to path
func RenameAll(pattern string, path string) error {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	for _, m := range matches {
		_, f := filepath.Split(m)
		err := os.Rename(m, filepath.Join(path, f))
		if err != nil {
			continue
		}
	}
	return nil
}

// RemoveEmpty removes a file or directory if it is empty. If it is not empty,
// RemoveEmpty returns and error
func RemoveEmpty(f string) error {

	var (
		empty bool
		err   error
	)

	if empty, err = IsEmpty(f); err != nil {
		return err
	}

	if empty {
		if err = os.Remove(f); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("file or directory %s is not empty: cannot be removed", f)
	}

	return nil
}

// SplitPath splits the path of a file into the different folders and files.
func SplitPath(path string) []string {
	if len(path) == 0 {
		return []string{}
	}
	b := (path[0] == os.PathSeparator)
	res := strings.Split(path, string(os.PathSeparator))
	if b {
		if len(res) == 1 {
			return []string{string(os.PathSeparator)}
		}
		res[1] = string(os.PathSeparator) + res[1]
		return res[1:]
	}
	return res
}

// Stat returns info about the file whose path is passed as
// parameter. In this regard, it is simlar to the standard function os.Stat.
// Different from it, FileStat return file info of type FileInfo, i.e. extended
// by Path(), which return the path of the file.
func Stat(path string) (Info, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return newInfo(fi, path), nil
}

// Suffix return the suffix of a file without the dot. If the file name
// contains no dot, an empty string is returned
func Suffix(f string) string {
	if len(path.Ext(f)) == 0 {
		return ""
	}
	return path.Ext(f)[1:]
}

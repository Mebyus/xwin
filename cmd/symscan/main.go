package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func fatal(v any) {
	fmt.Fprintf(os.Stderr, "%v\n", v)
	os.Exit(1)
}

var Target string

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fatal("start directory and target symbol must be specified")
	}

	Target = args[0]
	dir := args[1]
	err := walk(dir)
	if err != nil {
		fatal(err)
	}
}

func walk(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			// TODO: optionally report this error, but continue scan
			return nil
		}
		return err
	}
	for _, e := range entries {
		t := e.Type()
		if t.IsDir() && t&fs.ModeSymlink == 0 {
			err = walk(filepath.Join(dir, e.Name()))
			if err != nil {
				return err
			}
		}

		if t.IsRegular() {
			name := e.Name()
			if isDynamicLibraryName(name) {
				path := filepath.Join(dir, name)
				list, err := listDynamicSymbols(path)
				if err != nil {
					return fmt.Errorf("list symbols in %s: %w", path, err)
				}
				for _, s := range list {
					if s == Target {
						fmt.Printf("%s: %s\n", path, s)
					}
				}
			}
		}
	}
	return nil
}

func isDynamicLibraryName(name string) bool {
	split := strings.Split(name, ".")
	if len(split) < 2 {
		return false
	}
	if len(split) == 3 {
		if split[2] == "cache" { // exclude /etc/ld.so.cache
			return false
		}
		if split[2] == "conf" { // exclude /etc/ld.so.conf
			return false
		}
		if split[2] == "sig" {
			return false
		}
	}
	if split[1] == "so" {
		return true
	}
	// TODO: improve this detection algorithm
	return false
}

func listDynamicSymbols(path string) ([]string, error) {
	var buf bytes.Buffer

	c := exec.Command("nm", "-D", path)
	c.Stdout = &buf

	err := c.Run()
	if err != nil {
		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			if exitError.ExitCode() == 1 {
				return nil, nil
			}
			if bytes.Contains(exitError.Stderr, []byte("file format not recognized")) {
				return nil, nil
			}
		}
		return nil, err
	}

	var list []string
	s := bufio.NewScanner(&buf)
	for s.Scan() {
		line := s.Text()
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}

		if fields[1] != "T" {
			continue
		}

		list = append(list, fields[2])
	}
	err = s.Err()
	if err != nil {
		return nil, err
	}

	return list, nil
}

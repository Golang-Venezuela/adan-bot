package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strings"
)

//nolint:cyclop,gocognit,nestif
func Profile(fn func() error) (err error) {
	if os.Getenv("TRACE") == "1" {
		var fname string

		if v := os.Getenv("TRACE_FILE"); v != "" {
			fname = v
		} else {
			fname = "trace.out"
		}

		f, err := os.Create(fname)
		if err != nil {
			return fmt.Errorf("cannot create trace execution file: %w", err)
		}

		defer func() {
			if errC := f.Close(); errC != nil {
				errC = fmt.Errorf("cannot close trace execution file: %w", errC)
				err = errors.Join(err, errC)
			}
		}()

		if err := trace.Start(f); err != nil {
			return fmt.Errorf("cannot start execution tracing: %w", err)
		}

		defer trace.Stop()
	}

	if os.Getenv("PROFILE_CPU") == "1" {
		var fname string

		if v := os.Getenv("PROFILE_CPU_FILE"); v != "" {
			fname = v
		} else {
			fname = "cpu.pprof"
		}

		f, err := os.Create(fname)
		if err != nil {
			return fmt.Errorf("cannot create cpu profile file: %w", err)
		}

		defer func() {
			if errC := f.Close(); errC != nil {
				errC = fmt.Errorf("cannot close cpu profile file: %w", errC)
				err = errors.Join(err, errC)
			}
		}()

		if err := pprof.StartCPUProfile(f); err != nil {
			return fmt.Errorf("cannot profile cpu usage: %w", err)
		}

		defer pprof.StopCPUProfile()
	}

	if err := fn(); err != nil {
		return err
	}

	for _, prof := range pprof.Profiles() {
		name := prof.Name()
		ev := "PROFILE_" + strings.ToUpper(name)

		if os.Getenv(ev) != "1" {
			continue
		}

		var fname string

		if v := os.Getenv(ev + "_FILE"); v != "" {
			fname = v
		} else {
			fname = name + ".pprof"
		}

		if err := writeProfileToFile(fname, name); err != nil {
			return fmt.Errorf("cannot write %s profile: %w", name, err)
		}
	}

	return nil
}

func writeProfile(w io.Writer, name string) error {
	prof := pprof.Lookup(name)
	if prof == nil {
		return errors.New("invalid profile given")
	}

	if name == "allocs" || name == "heap" {
		runtime.GC()
	}

	if err := prof.WriteTo(w, 0); err != nil {
		return fmt.Errorf("cannot write: %w", err)
	}

	return nil
}

func writeProfileToFile(fname, name string) error {
	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}

	defer f.Close()

	if err := writeProfile(f, name); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("cannot close file: %w", err)
	}

	return nil
}

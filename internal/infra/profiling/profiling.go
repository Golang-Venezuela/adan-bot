// Package profiling offers native runtime instrumentation utilities.
// It vastly simplifies the integration of CPU profiling and execution tracing structurally
// by intercepting predefined environment flags seamlessly.
package profiling

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

// Profile acts as a foundational middleware or wrapper designated for the main entry point of the application.
// It actively evaluates discrete environment variables (e.g., TRACE=1, PROFILE_CPU=1) and structurally
// generates tracing execution files and pprof profile artifacts suited for deep performance analysis.
//
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

// writeProfile resolves a pprof entity by its corresponding name and securely writes its
// serialized contents inherently to the provided io.Writer parameter.
// When dealing with memory-based profiles ("allocs" or "heap"), it forces an instantaneous
// Garbage Collection cycle to proactively assure output accuracy.
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

// writeProfileToFile physically coordinates the creation of a system file on disk
// and subsequently delegating the raw telemetry parsing and payload writing logic downward to writeProfile.
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

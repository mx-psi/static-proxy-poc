// Command mkproxy builds the static GOPROXY artifacts for one module version from a
// source directory: <version>.zip (via x/mod/zip), <version>.mod (a byte-identical copy
// of the go.mod), and <version>.info plus latest.info (JSON). It prints the h1: hash for
// cross-checking against go.sum. The source dir must already contain the generated files
// to embed.
//
// Usage: mkproxy <module-path> <version> <src-dir> <out-dir> <rfc3339-time>
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/mod/module"
	"golang.org/x/mod/sumdb/dirhash"
	"golang.org/x/mod/zip"
)

func main() {
	if len(os.Args) != 6 {
		fmt.Fprintln(os.Stderr, "usage: mkproxy <module-path> <version> <src-dir> <out-dir> <rfc3339-time>")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2], os.Args[3], os.Args[4], os.Args[5]); err != nil {
		fmt.Fprintln(os.Stderr, "mkproxy:", err)
		os.Exit(1)
	}
}

func run(modpath, version, src, out, stamp string) error {
	if _, err := time.Parse(time.RFC3339, stamp); err != nil {
		return fmt.Errorf("bad time %q: %w", stamp, err)
	}

	mv := module.Version{Path: modpath, Version: version}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}

	// Copy the repo LICENSE into the zip
	if err := ensureLicense(src); err != nil {
		return err
	}

	// <version>.zip: entries are prefixed with "<modpath>@<version>/".
	zipPath := filepath.Join(out, version+".zip")
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	if err := zip.CreateFromDir(f, mv, src); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	// <version>.mod: byte-identical to the go.mod inside the zip.
	gomod, err := os.ReadFile(filepath.Join(src, "go.mod"))
	if err != nil {
		return err
	}
	if err := write(filepath.Join(out, version+".mod"), gomod); err != nil {
		return err
	}

	// <version>.info and latest.info: JSON {Version,Time}. latest.info backs @latest
	// via the "latest release" redirect.
	info := []byte(fmt.Sprintf("{%q:%q,%q:%q}\n", "Version", version, "Time", stamp))
	if err := write(filepath.Join(out, version+".info"), info); err != nil {
		return err
	}
	if err := write(filepath.Join(out, "latest.info"), info); err != nil {
		return err
	}

	h, err := dirhash.HashZip(zipPath, dirhash.DefaultHash)
	if err != nil {
		return err
	}
	fmt.Printf("%s %s\n", modpath+"@"+version, h)
	return nil
}

func write(path string, b []byte) error {
	return os.WriteFile(path, b, 0o644)
}

// ensureLicense copies a LICENSE from the nearest ancestor dir into src if src lacks
// its own, erroring if none is found.
func ensureLicense(src string) error {
	dst := filepath.Join(src, "LICENSE")
	if _, err := os.Stat(dst); err == nil {
		return nil // module already ships its own LICENSE.
	} else if !os.IsNotExist(err) {
		return err
	}

	for dir := filepath.Dir(src); ; {
		b, err := os.ReadFile(filepath.Join(dir, "LICENSE"))
		if err == nil {
			return write(dst, b)
		}
		if !os.IsNotExist(err) {
			return err
		}
		parent := filepath.Dir(dir)
		if parent == dir { // reached the filesystem root.
			return fmt.Errorf("no LICENSE found in %s or any ancestor directory", src)
		}
		dir = parent
	}
}

package main

import (
	"os"
	"path/filepath"
	"syscall"
)

func pivotRoot(newroot string) error {
	putold := filepath.Join(newroot, "/.pivot_root")

	// bind mount newroot to itself
	if err := syscall.Mount(
		newroot,
		newroot,
		"",
		syscall.MS_BIND|syscall.MS_REC,
		"",
	); err != nil {
		return err
	}

	// create putold directory
	if err := os.MkdirAll(putold, 0700); err != nil {
		return err
	}

	// call pivot_root
	if err := syscall.PivotRoot(newroot, putold); err != nil {
		return err
	}

	// ensure current working directory is set to new root
	if err := os.Chdir("/"); err != nil {
		return err
	}

	// unmount putold, which now lives at /.pivot_root
	putold = "/.pivot_root"
	if err := syscall.Unmount(
		putold,
		syscall.MNT_DETACH,
	); err != nil {
		return err
	}

	// remove putold
	if err := os.RemoveAll(putold); err != nil {
		return err
	}

	return nil
}

func mountProc(newRoot string) error {
	target := filepath.Join(newRoot, "/proc")
	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	if err := syscall.Mount(
		"proc",
		target,
		"proc",
		0,
		"",
	); err != nil {
		return err
	}
	return nil
}

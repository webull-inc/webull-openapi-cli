//go:build windows

// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
)

const (
	lockfileExclusiveLock   = 0x02
	lockfileFailImmediately = 0x01
)

// writeFileWithLock writes data to a file using LockFileEx (Windows).
func writeFileWithLock(path string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	// Lock the file
	ol := new(syscall.Overlapped)
	handle := syscall.Handle(f.Fd())
	r1, _, _ := procLockFileEx.Call(
		uintptr(handle),
		uintptr(lockfileExclusiveLock),
		0,
		1, 0,
		uintptr(unsafe.Pointer(ol)),
	)
	if r1 == 0 {
		// Lock failed, write without lock
		_, err = f.Write(data)
		return err
	}

	defer procUnlockFileEx.Call(
		uintptr(handle),
		0,
		1, 0,
		uintptr(unsafe.Pointer(ol)),
	)

	_, err = f.Write(data)
	return err
}

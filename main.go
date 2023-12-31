package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	user32             = syscall.MustLoadDLL("user32.dll")
	procEnumWindows    = user32.MustFindProc("EnumWindows")
	procGetWindowTextW = user32.MustFindProc("GetWindowTextW")
	procSetWindowTextW = user32.MustFindProc("SetWindowTextW")
)

func EnumWindows(enumFunc uintptr, lparam uintptr) (err error) {
	r1, _, e1 := syscall.Syscall(procEnumWindows.Addr(), 2, uintptr(enumFunc), uintptr(lparam), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func GetWindowText(hwnd syscall.Handle, str *uint16, maxCount int32) (len int32, err error) {
	r0, _, e1 := syscall.Syscall(procGetWindowTextW.Addr(), 3, uintptr(hwnd), uintptr(unsafe.Pointer(str)), uintptr(maxCount))
	len = int32(r0)
	if len == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func FindWindow(title string) (syscall.Handle, error) {
	var hwnd syscall.Handle
	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		b := make([]uint16, 200)
		_, err := GetWindowText(h, &b[0], int32(len(b)))
		if err != nil {
			// ignore the error
			return 1 // continue enumeration
		}

		if syscall.UTF16ToString(b) == title {
			// note the window
			hwnd = h
			return 0 // stop enumeration
		}
		return 1 // continue enumeration
	})
	EnumWindows(cb, 0)
	if hwnd == 0 {
		return 0, fmt.Errorf("no window with title '%s' found", title)
	}
	return hwnd, nil
}

func SetWindowText(hwnd syscall.Handle, title string) (err error) {

	b, err := syscall.UTF16FromString(title)

	r1, _, e1 := syscall.Syscall(procSetWindowTextW.Addr(), 2, uintptr(hwnd), uintptr(unsafe.Pointer(&b[0])), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func help() {
	fmt.Println("WinChangeTitle - Tool for changing a Win32 app title at runtime, (c) 2023")
	fmt.Println("Usage:")
	fmt.Println("  winchangetitle <old-title> <new-title>")
	fmt.Println("Note: Use quotes if a title contains space characters.")
}

func main() {
	args := os.Args
	if len(args) < 3 {
		help()
		os.Exit(0)
	}

	oldTitle := args[1]
	newTitle := args[2]

	h, err := FindWindow(oldTitle)
	if err != nil {
		fmt.Println("Window not found")
		os.Exit(1)
	}
	fmt.Println("Window title replaced successfully")

	err = SetWindowText(h, newTitle)
	if err != nil {
		fmt.Println("Can't replace the window title")
	}
}

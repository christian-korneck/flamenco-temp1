//go:build windows

package find_blender

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// fileAssociation returns the full path of `blender.exe` associated with ".blend" files.
func fileAssociation() (string, error) {
	exe, err := getFileAssociation(".blend")
	if err != nil {
		return "", err
	}

	// Often the association will be with blender-launcher.exe, which is
	// unsuitable for use in Flamenco. Use its path to find its `blender.exe`.
	dir, file := filepath.Split(exe)
	if file != "blender-launcher.exe" {
		return exe, nil
	}

	blenderPath := filepath.Join(dir, "blender.exe")
	_, err = os.Stat(blenderPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("blender-launcher found at %s but not its blender.exe", exe)
		}
		return "", fmt.Errorf("investigating %s: %w", blenderPath, err)
	}

	return blenderPath, nil
}

// getFileAssociation finds the executable associated with the given extension.
// The extension must be a string like ".blend".
func getFileAssociation(extension string) (string, error) {
	// Load library.
	libname := "shlwapi.dll"
	libshlwapi, err := syscall.LoadLibrary(libname)
	if err != nil {
		return "", fmt.Errorf("loading %s: %w", libname, err)
	}
	defer func() { _ = syscall.FreeLibrary(libshlwapi) }()

	// Find function.
	funcname := "AssocQueryStringW"
	assocQueryString, err := syscall.GetProcAddress(libshlwapi, funcname)
	if err != nil {
		return "", fmt.Errorf("finding function %s in %s: %w", funcname, libname, err)
	}

	// https://docs.microsoft.com/en-gb/windows/win32/api/shlwapi/nf-shlwapi-assocquerystringw
	pszAssoc, err := syscall.UTF16PtrFromString(extension)
	if err != nil {
		return "", fmt.Errorf("converting string to UTF16: %w", err)
	}

	pszExtra, err := syscall.UTF16PtrFromString("open")
	if err != nil {
		return "", fmt.Errorf("converting string to UTF16: %w", err)
	}

	var cchOut uint32 = 65535
	buf := make([]uint16, cchOut)
	pszOut := unsafe.Pointer(&buf[0])

	result1, _, errno := syscall.SyscallN(
		assocQueryString,
		uintptr(ASSOCF_INIT_DEFAULTTOSTAR), // [in]            ASSOCF   flags
		uintptr(ASSOCSTR_EXECUTABLE),       // [in]            ASSOCSTR str
		uintptr(unsafe.Pointer(pszAssoc)),  // [in]            LPCWSTR  pszAssoc
		uintptr(unsafe.Pointer(pszExtra)),  // [in, optional]  LPCWSTR  pszExtra
		uintptr(pszOut),                    // [out, optional] LPWSTR   pszOut
		uintptr(unsafe.Pointer(&cchOut)),   // [in, out]       DWORD    *pcchOut
	)
	if errno != 0 {
		return "", fmt.Errorf("error calling AssocQueryStringW from shlwapi.dll: %w", errno)
	}
	if result1 != 0 {
		return "", fmt.Errorf("unknown result %d calling AssocQueryStringW from shlwapi.dll: %w", result1, err)
	}

	exe := syscall.UTF16ToString(buf)
	return exe, nil
}

// Source: https://docs.microsoft.com/en-us/windows/win32/shell/assocf_str
const (
	ASSOCF_NONE                 = ASSOCF(0x00000000)
	ASSOCF_INIT_NOREMAPCLSID    = ASSOCF(0x00000001)
	ASSOCF_INIT_BYEXENAME       = ASSOCF(0x00000002)
	ASSOCF_OPEN_BYEXENAME       = ASSOCF(0x00000002)
	ASSOCF_INIT_DEFAULTTOSTAR   = ASSOCF(0x00000004)
	ASSOCF_INIT_DEFAULTTOFOLDER = ASSOCF(0x00000008)
	ASSOCF_NOUSERSETTINGS       = ASSOCF(0x00000010)
	ASSOCF_NOTRUNCATE           = ASSOCF(0x00000020)
	ASSOCF_VERIFY               = ASSOCF(0x00000040)
	ASSOCF_REMAPRUNDLL          = ASSOCF(0x00000080)
	ASSOCF_NOFIXUPS             = ASSOCF(0x00000100)
	ASSOCF_IGNOREBASECLASS      = ASSOCF(0x00000200)
	ASSOCF_INIT_IGNOREUNKNOWN   = ASSOCF(0x00000400)
	ASSOCF_INIT_FIXED_PROGID    = ASSOCF(0x00000800)
	ASSOCF_IS_PROTOCOL          = ASSOCF(0x00001000)
	ASSOCF_INIT_FOR_FILE        = ASSOCF(0x00002000)
)

type ASSOCF uint32

// Source: https://docs.microsoft.com/en-us/windows/win32/api/shlwapi/ne-shlwapi-assocstr
const (
	ASSOCSTR_COMMAND ASSOCSTR = iota + 1
	ASSOCSTR_EXECUTABLE
	ASSOCSTR_FRIENDLYDOCNAME
	ASSOCSTR_FRIENDLYAPPNAME
	ASSOCSTR_NOOPEN
	ASSOCSTR_SHELLNEWVALUE
	ASSOCSTR_DDECOMMAND
	ASSOCSTR_DDEIFEXEC
	ASSOCSTR_DDEAPPLICATION
	ASSOCSTR_DDETOPIC
	ASSOCSTR_INFOTIP
	ASSOCSTR_QUICKTIP
	ASSOCSTR_TILEINFO
	ASSOCSTR_CONTENTTYPE
	ASSOCSTR_DEFAULTICON
	ASSOCSTR_SHELLEXTENSION
	ASSOCSTR_DROPTARGET
	ASSOCSTR_DELEGATEEXECUTE
	ASSOCSTR_SUPPORTED_URI_PROTOCOLS
	ASSOCSTR_PROGID
	ASSOCSTR_APPID
	ASSOCSTR_APPPUBLISHER
	ASSOCSTR_APPICONREFERENCE
	ASSOCSTR_MAX
)

type ASSOCSTR uint32

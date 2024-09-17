package util

import (
	"errors"
	"net"
	"syscall"
)

func IsNetworkError(err error) bool {
	if err == nil {
		// "Ok"
		return false
	}
	var netError net.Error
	if errors.As(err, &netError) && netError.Timeout() {
		// "Timeout"
		return true
	}

	var netDNSError *net.DNSError
	if errors.As(err, &netDNSError) {
		if netDNSError.Err == "no such host" {
			// "Unknown host"
			return true
		}
	}
	var netOpError *net.OpError
	if errors.As(err, &netOpError) {
		if netOpError.Op == "dial" {
			// "Unknown host"
			return true
		}
		if netOpError.Op == "read" {
			// "Connection refused"
			return true
		}
	}

	var syscallError syscall.Errno
	if errors.As(err, &syscallError) {
		// numbers are the error codes for "Connection refused" for Windows systems (not included in UNIX syscall package)
		if errors.Is(syscallError, syscall.ECONNREFUSED) || syscallError == 10061 {
			// "Connection refused"
			return true
		}
		if errors.Is(syscallError, syscall.ECONNRESET) || syscallError == 10054 {
			// "Connection reset"
			return true
		}
		if errors.Is(syscallError, syscall.ECONNABORTED) || syscallError == 10053 {
			// "Connection aborted"
			return true
		}
	}
	return false
}

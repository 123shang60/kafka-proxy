package proxy

import (
	"io"
	"net"
)

// myCopyN is similar to io.CopyN, but reports whether the returned error was due
// to a bad read or write. The returned error will never be nil
func myCopyN(dst io.Writer, src io.Reader, size int64, buf []byte) (readErr bool, err error) {
	// If both src and dst are net.Conns, use io.CopyN directly.
	if dstNet, ok := dst.(net.Conn); ok {
		if srcNet, ok := src.(net.Conn); ok {
			return netConnCopy(dstNet, srcNet, size)
		}
	}
	return notNetConnCopy(dst, src, size, buf)
}

func notNetConnCopy(dst io.Writer, src io.Reader, size int64, buf []byte) (readErr bool, err error) {
	// limit reader  - EOF when finished
	src = io.LimitReader(src, size)

	var written int64
	var n int
	for {
		n, err = src.Read(buf)
		if n > 0 {
			nw, werr := dst.Write(buf[0:n])
			if nw > 0 {
				written += int64(nw)
			}
			if err != nil {
				// Read and write error; just report read error (it happened first).
				readErr = true
				break
			}
			if werr != nil {
				err = werr
				break
			}
			if n != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if err != nil {
			readErr = true
			break
		}
	}

	if written == size {
		return false, nil
	}
	if written < size && err == nil {
		// src stopped early; must have been EOF.
		readErr = true
		err = io.EOF
	}
	return
}

func netConnCopy(dstNet net.Conn, srcNet net.Conn, size int64) (readErr bool, err error) {
	readErr = true
	var written int64 = 0
	written, err = io.CopyN(dstNet, srcNet, size)
	if written == size {
		return false, nil
	}
	if written < size && err == nil {
		// src stopped early; must have been EOF.
		readErr = true
		err = io.EOF
	}
	return
}

package http_

import (
	"bufio"
	"errors"
	"github.com/searKing/golib/io_"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// The algorithm uses at most sniffLen bytes to make its decision.
const sniffLen = 512

func ContentType(content io.Reader, name string) (ctype string, bufferedContent io.Reader, err error) {
	ctype = mime.TypeByExtension(filepath.Ext(name))
	if ctype == "" && content != nil {
		// read a chunk to decide between utf-8 text and binary
		var buf [sniffLen]byte
		var n int
		if readSeeker, ok := content.(io.Seeker); ok {
			n, _ = io.ReadFull(content, buf[:])
			_, err = readSeeker.Seek(0, io.SeekStart) // rewind to output whole file
			if err != nil {
				err = errors.New("seeker can't seek")
				return "", content, err
			}
		} else {
			contentBuffer := bufio.NewReader(content)
			sniffed, err := contentBuffer.Peek(sniffLen)
			if err != nil {
				err = errors.New("reader can't read")
				return "", contentBuffer, err
			}
			n = copy(buf[:], sniffed)
			content = contentBuffer
		}
		ctype = http.DetectContentType(buf[:n])
	}
	return ctype, content, nil
}

func ServeContent(w http.ResponseWriter, r *http.Request, name string, modtime time.Time, content io.Reader, size int64) {
	readseeker, ok := content.(io.ReadSeeker)
	if !ok {
		ctype, content, err := ContentType(content, name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", ctype)

		readseeker = newServeContentSeekable(content, size)
	}

	if modtime.IsZero() {
		modtime = time.Now()
	}
	if stater, ok := content.(io_.Stater); ok {
		if fi, err := stater.Stat(); err == nil {
			modtime = fi.ModTime()
		}
	}
	http.ServeContent(w, r, name, modtime, readseeker)
	return
}

// can only be used for ServeContent
type serveContentSeekable struct {
	io.Reader
	size int64
}

func newServeContentSeekable(r io.Reader, size int64) *serveContentSeekable {
	return &serveContentSeekable{
		Reader: r,
		size:   size,
	}
}

func (s *serveContentSeekable) Seek(offset int64, whence int) (int64, error) {
	if offset != 0 {
		return 0, os.ErrInvalid
	}
	if whence == io.SeekStart {
		return 0, nil
	}
	if whence == io.SeekEnd {
		return s.size, nil
	}
	return 0, os.ErrInvalid
}

package core

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/k0kubun/pp"

	pool "github.com/libp2p/go-buffer-pool"
)

var cpool = pool.BufferPool{}

func Impls(name string, src any) {
	pp.Println(name)

	if _, ok := src.(io.WriterTo); ok {
		pp.Println("@implements WriterTo")

	}

	if _, ok := src.(io.ReaderFrom); ok {
		pp.Println("@implements WriterTo")

	}

	if _, ok := src.(*io.LimitedReader); ok {
		pp.Println("@is *LimitedReader")
	}

	fmt.Print("\n")

}

// this is almost exact copy from io packages but ability to throttle
func Copy(dst io.Writer, src io.Reader, throttle bool) (int64, error) {
	size := 1024 * 32
	if throttle {
		size = 1024 * 8
	}

	buf := cpool.Get(size)

	n, err := copyBuffer(dst, src, buf, throttle)
	defer cpool.Put(buf)
	return n, err
}

var errInvalidWrite = errors.New("invalid write result")

func copyBuffer(dst io.Writer, src io.Reader, buf []byte, throttle bool) (written int64, err error) {

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}

		if throttle {
			pp.Println("@throttling")
			time.Sleep(time.Millisecond * 250)
		}

	}
	return written, err
}

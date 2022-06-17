package file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/helpers"
)

const (
	bufSize       = 0x10000 // 64 KiB
	flushInterval = 100 * time.Millisecond
)

var (
	files      = map[string]*syncWriter{}
	filesMutex = sync.Mutex{}
)

type syncWriter struct {
	w     *bufio.Writer
	mutex sync.Mutex
}

func openFile(filename string, perm os.FileMode) (io.Writer, error) {
	filename = config.AbsPath(filename)

	filesMutex.Lock()
	defer filesMutex.Unlock()

	file, ok := files[filename]
	if ok {
		return file, nil
	}

	f, err := os.OpenFile(config.AbsPath(filename), os.O_CREATE|os.O_APPEND|os.O_WRONLY, perm)
	if err != nil {
		return nil, err
	}

	file = new(syncWriter)
	file.w = bufio.NewWriterSize(f, bufSize)
	files[filename] = file

	helpers.NewInterval(flushInterval, func() {
		err := file.Flush()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}).Start()

	return file, nil
}

func (f *syncWriter) Write(b []byte) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return f.w.Write(b)
}

func (f *syncWriter) Flush() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return f.w.Flush()
}

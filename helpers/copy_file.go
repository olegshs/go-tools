package helpers

import (
	"io"
	"os"
	"time"
)

func CopyFile(srcFilename string, dstFilename string) error {
	stat, err := os.Stat(srcFilename)
	if err != nil {
		return err
	}

	src, err := os.Open(srcFilename)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, stat.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	err = os.Chtimes(dstFilename, time.Now(), stat.ModTime())
	if err != nil {
		return err
	}

	return nil
}

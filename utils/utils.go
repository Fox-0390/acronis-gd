package utils

import (
	"os"
	"io"
	bufio "bufio"
)

func CreateDirectory(path string) error {
	return os.MkdirAll(path, 0777)
}

func CreateFile(path string, content []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(content)

	return err
}

func CreateFileWithReader(path string, closer io.Reader) error {
	fo, err := os.Create(path)
	if err != nil {
		return err
	}

	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	// make a write buffer
	w := bufio.NewWriter(fo)

	// make a buffer to keep chunks that are read
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := closer.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			panic(err)
		}
	}

	if err = w.Flush(); err != nil {
		return err
	}

	return nil
}

func ReadFile(path string) (io.Reader, error) {
	return os.Open(path)
}

func IsFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

func RemoveFile(path string) error {
	return os.Remove(path)
}
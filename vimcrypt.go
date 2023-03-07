package vimcrypt

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

const cryptPrefix = "VimCrypt~"

var errNotVimEncrypted = errors.New("data is not vim encrypted")
var errUnsupportedCryptMethod = errors.New("unsupported crypto method")

type Reader struct {
	originReader io.Reader
	decompressor func(io.Reader) ([]byte, error)
}

func NewReader(r io.Reader, key []byte) (*Reader, error) {
	z := new(Reader)
	z.originReader = bufio.NewReader(r)
	err := z.readHeader(key)
	return z, err
}

func (z *Reader) Read(b []byte) (n int, err error) {
	buf, err := z.decompressor(z.originReader)
	if err != nil {
		return
	}
	n = len(buf)
	copy(b, buf)
	return
}

func (z *Reader) readHeader(key []byte) (err error) {
	headerSize := len(cryptPrefix)
	buf := make([]byte, 512)
	if _, err = io.ReadFull(z.originReader, buf[0:headerSize]); err != nil {
		return fmt.Errorf("read failed: %w", err)
	}
	if string(buf[0:headerSize]) != cryptPrefix {
		return fmt.Errorf("%w: invalid header, data does not start with %s", errNotVimEncrypted, cryptPrefix)
	}
	if _, err = io.ReadFull(z.originReader, buf[0:3]); err != nil {
		return fmt.Errorf("invalid header, could not read crypt method: %w", err)
	}

	switch string(buf[0:3]) {
	case "01!":
		// zip
		return fmt.Errorf("%w: 'zip' is currently not supported", errUnsupportedCryptMethod)
	case "02!":
		// blowfish
		return fmt.Errorf("%w: 'blowfish' is currently not supported", errUnsupportedCryptMethod)
	case "03!":
		// blowfish2
		// details in https://github.com/vim/vim/blob/master/src/crypt.c
		// header starts with 8 bytes of salt
		if _, err = io.ReadFull(z.originReader, buf[0:8]); err != nil {
			return fmt.Errorf("invalid header, could not read salt: %w", err)
		}
		salt := buf[0:8]
		z.decompressor, err = NewBlowfish2(key, salt, bfDecrypt)
		if err != nil {
			return fmt.Errorf("failed to initialize blowfish2 decrypt: %w", err)
		}
		return
	}
	return fmt.Errorf("%w: '%s' does not map any crypt method", errUnsupportedCryptMethod, buf[0:3])
}

type Writer struct {
	originWriter io.Writer
	decompressor func(io.Reader) ([]byte, error)
	salt         []byte
	seed         []byte
}

func NewWriter(w io.Writer, key []byte) (*Writer, error) {
	z := new(Writer)
	z.originWriter = w
	var err error
	z.salt, err = z.buildSalt()
	if err != nil {
		return nil, fmt.Errorf("getting salt failed: %w", err)
	}
	z.seed, err = z.buildSalt()
	if err != nil {
		return nil, fmt.Errorf("getting seed failed: %w", err)
	}
	z.decompressor, err = NewBlowfish2(key, z.salt, bfEncrypt)
	if err != nil {
		return nil, fmt.Errorf("blowfish2 init failed: %w", err)
	}
	return z, nil
}

func (z *Writer) Write(b []byte) (n int, err error) {
	_, err = fmt.Fprintf(z.originWriter, "%s03!%s%s", cryptPrefix, z.salt, z.seed)
	if err != nil {
		return 0, fmt.Errorf("write failed: %w", err)
	}

	data := make([]byte, len(z.seed)+len(b))
	copy(data, z.seed)
	copy(data[len(z.seed):], b)
	encrypted, err := z.decompressor(bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("encrypt failed: %w", err)
	}
	n, err = z.originWriter.Write(encrypted)
	if err != nil {
		return 0, fmt.Errorf("write failed: %w", err)
	}
	return
}

func (z *Writer) buildSalt() (salt []byte, err error) {
	salt = make([]byte, 8)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("salt initialize failed: %w", err)
	}
	return
}

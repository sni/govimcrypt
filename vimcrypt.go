package vimcrypt

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

const cryptPrefix = "VimCrypt~"

var errNotVimEncrypted = errors.New("data is not vim encrypted")
var errUnsupportedCryptMethod = errors.New("unsupported crypto method")

func NewReader(reader io.Reader, key []byte) (io.Reader, error) {
	decryptingReader, err := readHeader(reader, key)
	if err != nil {
		return nil, fmt.Errorf("vimcrypt.NewReader: %w", err)
	}
	return decryptingReader, nil
}

func readHeader(originReader io.Reader, key []byte) (io.Reader, error) {
	headerSize := len(cryptPrefix)
	buf := make([]byte, 512)
	if _, err := io.ReadFull(originReader, buf[0:headerSize]); err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}
	if string(buf[0:headerSize]) != cryptPrefix {
		return nil, fmt.Errorf("%w: invalid header, data does not start with %s", errNotVimEncrypted, cryptPrefix)
	}
	if _, err := io.ReadFull(originReader, buf[0:3]); err != nil {
		return nil, fmt.Errorf("invalid header, could not read crypt method: %w", err)
	}

	switch string(buf[0:3]) {
	case "01!":
		// zip
		return nil, fmt.Errorf("%w: 'zip' is currently not supported", errUnsupportedCryptMethod)
	case "02!":
		// blowfish
		return nil, fmt.Errorf("%w: 'blowfish' is currently not supported", errUnsupportedCryptMethod)
	case "03!":
		// blowfish2
		// details in https://github.com/vim/vim/blob/master/src/crypt.c
		// header starts with 8 bytes of salt, 8 bytes of seed (iv)
		if _, err := io.ReadFull(originReader, buf[0:16]); err != nil {
			return nil, fmt.Errorf("invalid header, could not read salt+seed: %w", err)
		}
		salt := buf[0:8]
		seed := buf[8:16]
		decryptingReader, err := newBlowfish2Reader(key, salt, seed, originReader)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize blowfish2 decrypt: %w", err)
		}
		return decryptingReader, nil
	}
	return nil, fmt.Errorf("%w: '%s' does not map any crypt method", errUnsupportedCryptMethod, buf[0:3])
}

func NewWriter(writer io.Writer, key []byte) (io.Writer, error) {
	salt, err := buildSalt()
	if err != nil {
		return nil, fmt.Errorf("generating salt failed: %w", err)
	}
	seed, err := buildSalt()
	if err != nil {
		return nil, fmt.Errorf("generating seed failed: %w", err)
	}
	_, err = fmt.Fprintf(writer, "%s03!%s%s", cryptPrefix, salt, seed)
	if err != nil {
		return nil, fmt.Errorf("writing header failed: %w", err)
	}
	encryptingWriter, err := newBlowfish2Writer(key, salt, seed, writer)
	if err != nil {
		return nil, fmt.Errorf("blowfish2 init failed: %w", err)
	}
	return encryptingWriter, nil
}

func buildSalt() (salt []byte, err error) {
	salt = make([]byte, 8)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("salt initialize failed: %w", err)
	}
	return
}

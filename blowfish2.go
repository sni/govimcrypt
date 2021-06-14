package vimcrypt

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/crypto/blowfish"
)

type cryptDirection int

const (
	bfEncrypt cryptDirection = iota
	bfDecrypt
)

func NewBlowfish2(key, salt []byte, mode cryptDirection) (func(io.Reader) ([]byte, error), error) {
	cipher, err := buildBlowfish2Cipher(key, salt)
	if err != nil {
		return nil, err
	}
	blocksize := blowfish.BlockSize
	decrypt := func(reader io.Reader) (decryped []byte, err error) {
		block0 := make([]byte, blocksize)
		n, err := io.ReadFull(reader, block0)
		if err != nil && n == 0 {
			return
		}
		block1 := make([]byte, blocksize)
		n, err = io.ReadFull(reader, block1)
		if err != nil && n == 0 {
			return
		}
		err = nil

		for {
			buf := append([]byte(nil), block0...)
			buf, _ = convertEndian(buf)
			cipher.Encrypt(buf, buf)
			buf, _ = convertEndian(buf)
			xorBytes(buf, buf, block1, blocksize)
			decryped = append(decryped, buf[:n]...)

			switch mode {
			case bfEncrypt:
				block0 = buf[:n]
			case bfDecrypt:
				block0 = block1
			}
			block1 = make([]byte, blocksize)
			n, _ = io.ReadFull(reader, block1)
			if n == 0 {
				return
			}
		}
	}
	return decrypt, nil
}

func buildBlowfish2Cipher(key, salt []byte) (cipher *blowfish.Cipher, err error) {
	pw := append([]byte(nil), key...)
	for i := 1; i <= 1000; i++ {
		hash := sha256.Sum256(append(pw, salt...))
		pw = []byte(fmt.Sprintf("%x", hash))
	}
	hash := sha256.Sum256(append(pw, salt...))
	cipher, err = blowfish.NewCipher(hash[:])
	return
}

func convertEndian(in []byte) ([]byte, error) {
	// Read byte array as uint32 (little-endian)
	var v1, v2 uint32
	buf := bytes.NewReader(in)
	if err := binary.Read(buf, binary.LittleEndian, &v1); err != nil {
		return nil, fmt.Errorf("convertEndian failed: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &v2); err != nil {
		return nil, fmt.Errorf("convertEndian failed: %w", err)
	}

	// convert uint32 to byte array
	out := make([]byte, 8)
	binary.BigEndian.PutUint32(out, v1)
	binary.BigEndian.PutUint32(out[4:], v2)

	return out, nil
}

func xorBytes(dst, a, b []byte, n int) {
	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
}

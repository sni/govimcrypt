package vimcrypt

import (
	"bytes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/crypto/blowfish"
)

func NewBlowfish2Reader(key, salt, seed []byte, reader io.Reader) (io.Reader, error) {
	bfCipher, err := buildBlowfish2Cipher(key, salt)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCFBDecrypter(bfCipher, seed)
	return &cipher.StreamReader{S: stream, R: reader}, nil
}

func NewBlowfish2Writer(key, salt, seed []byte, writer io.Writer) (io.Writer, error) {
	bfCipher, err := buildBlowfish2Cipher(key, salt)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(bfCipher, seed)
	return &cipher.StreamWriter{S: stream, W: writer, Err: nil}, nil
}

func buildBlowfish2Cipher(key, salt []byte) (cipher.Block, error) {
	pw := append([]byte(nil), key...)
	for i := 1; i <= 1000; i++ {
		hash := sha256.Sum256(append(pw, salt...))
		pw = []byte(fmt.Sprintf("%x", hash))
	}
	hash := sha256.Sum256(append(pw, salt...))
	bfCipher, err := blowfish.NewCipher(hash[:])
	return &VimBlowfish{bfCipher}, err //nolint:wrapcheck // can't happen
}

// VimBlowfish is the blowfish cipher, with an endianness conversion.
type VimBlowfish struct {
	*blowfish.Cipher
}

func (vbf *VimBlowfish) Encrypt(dst, src []byte) {
	convertEndian(src, src)
	vbf.Cipher.Encrypt(dst, src)
	convertEndian(dst, dst)
}

func (vbf *VimBlowfish) Decrypt(dst, src []byte) {
	// We provide Decrypt but note crypto/cipher.cfb only uses Encrypt.
	convertEndian(src, src)
	vbf.Cipher.Decrypt(dst, src)
	convertEndian(dst, dst)
}

func convertEndian(out, in []byte) {
	// Read byte array as uint32 (little-endian)
	var v1, v2 uint32
	buf := bytes.NewReader(in)
	if err := binary.Read(buf, binary.LittleEndian, &v1); err != nil {
		// crypto/cipher.Block interface assumes the byte arrays are the correct
		// size, the code later would panic anyway if there isn't enough to read.
		panic(err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &v2); err != nil {
		panic(err)
	}

	// convert uint32 to byte array
	binary.BigEndian.PutUint32(out, v1)
	binary.BigEndian.PutUint32(out[4:], v2)
}

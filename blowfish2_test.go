package vimcrypt_test

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sni/vimcrypt"
)

func TestBlowfish2ReadFile1(t *testing.T) {
	t.Parallel()

	file := "test/data/blowfish2/test"
	cipher := []byte("test")
	expect := "test\n"

	testFile(t, file, cipher, expect)
}

func TestBlowfish2ReadFile2(t *testing.T) {
	t.Parallel()

	file := "test/data/blowfish2/testtesttest"
	cipher := []byte("testtesttest")
	expect := "test file with longer key\n"

	testFile(t, file, cipher, expect)
}

func TestBlowfish2ReadFile_Large(t *testing.T) {
	t.Parallel()

	file := "test/data/blowfish2/40k"
	cipher := []byte("test")
	expect := strings.Repeat("A", 40000) + "\n"

	testFile(t, file, cipher, expect)
}

func TestBlowfish2Write(t *testing.T) {
	t.Parallel()

	cipher := []byte("testtesttest")
	content := "test file with longer key 12345678901234567890123456789012345678901234567890\n"

	buf := bytes.NewBuffer(nil)
	enc, err := vimcrypt.NewWriter(buf, cipher)
	if err != nil {
		t.Fatalf("creating writer failed: %s", err)
	}
	_, err = enc.Write([]byte(content))
	if err != nil {
		t.Fatalf("writer failed: %s", err)
	}

	dec, err := vimcrypt.NewReader(buf, cipher)
	if err != nil {
		t.Fatalf("creating reader failed: %s", err)
	}
	res := new(strings.Builder)
	n, err := io.Copy(res, dec)
	if err != nil {
		t.Fatalf("cannot read reader: %s", err)
	}
	if int64(len(content)) != n {
		t.Errorf("expected %d bytes, but got %d", len(content), n)
	}
	if res.String() != content {
		t.Errorf("expected '%s', but got '%s'", content, res.String())
	}
}

func testFile(t *testing.T, file string, cipher []byte, expect string) {
	t.Helper()

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("cannot open test file: %s", err)
	}

	r := bufio.NewReader(f)
	v, err := vimcrypt.NewReader(r, cipher)
	if err != nil {
		t.Fatalf("cannot create reader: %s", err)
	}

	buf := new(strings.Builder)
	n, err := io.Copy(buf, v)
	if err != nil {
		t.Fatalf("cannot read reader: %s", err)
	}
	if int64(len(expect)) != n {
		t.Errorf("expected %d bytes, but got %d", len(expect), n)
	}
	if buf.String() != expect {
		t.Errorf("expected '%s', but got '%s'", expect, buf.String())
	}
}

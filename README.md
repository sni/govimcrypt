Go VimCrypt
===========

Package vimcrypt implements reading and writing of vim crypted files with golang.

Supports blowfish2 only, since this is the recommended vim cryptmethod.

Installation
============

    %> go get github.com/sni/govimcrypt

Usage
=====

Reading
-------

```golang
password := []byte("secret")
file := "/tmp/testfile"
fh, err := os.Open(file)
r := bufio.NewReader(fh)
vc, err := vimcrypt.NewReader(r, password)
decrypted := new(strings.Builder)
io.Copy(decrypted, vc)
fmt.Printf("decrypted content is: %s", decrypted.String())
```


Writing
-------

```golang
password := []byte("secret")
file := "/tmp/testfile"
encrypted := bytes.NewBuffer(nil)
vc, err := vimcrypt.NewWriter(encrypted, password)
vc.Write([]byte("...clear text"))
ioutil.WriteFile(file, encrypted.Bytes(), 0644)
```
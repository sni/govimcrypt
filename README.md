# Go VimCrypt

Package vimcrypt implements reading and writing of vim encrypted files with golang.

Supports blowfish2 only, since this is the recommended vim cryptmethod.

## Using VIM

### Create

To create a new encrypted file, run:

    %> vim -x /tmp/testfile.txt

Enter the new password twice.

In vim, set the cryptmethod with:

    :set cm=blowfish2

Enter some text (the file must not be empty) and save the file:

    :x!

Verify the file is encrypted properly:

    %> file /tmp/testfile.txt
    /tmp/testfile.txt: Vim encrypted file data with blowfish2 cryptmethod

### Read

To read the file with vim again, you don't need to specify -x again:

    %> vim /tmp/testfile.txt

## Usage

### Reading

```golang
import (
    "github.com/sni/govimcrypt"
)

password := []byte("secret")
file := "/tmp/testfile.txt"
fh, err := os.Open(file)
r := bufio.NewReader(fh)
vc, err := vimcrypt.NewReader(r, password)
decrypted := new(strings.Builder)
io.Copy(decrypted, vc)
fmt.Printf("decrypted content is: %s", decrypted.String())
```

### Writing

```golang
import (
    "github.com/sni/govimcrypt"
)

password := []byte("secret")
file := "/tmp/testfile.txt"
encrypted := bytes.NewBuffer(nil)
vc, err := vimcrypt.NewWriter(encrypted, password)
vc.Write([]byte("...clear text"))
ioutil.WriteFile(file, encrypted.Bytes(), 0644)
```

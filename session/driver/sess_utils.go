package driver

import (
	"bytes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	r "math/rand"
	"strconv"
	"time"
)

func init() {
	gob.Register([]interface{}{})
	gob.Register(map[int]interface{}{})
	gob.Register(map[string]interface{}{})
	gob.Register(map[interface{}]interface{}{})
	gob.Register(map[string]string{})
	gob.Register(map[int]string{})
	gob.Register(map[int]int{})
	gob.Register(map[int]int64{})
}

func EncodeGob(obj map[interface{}]interface{}) ([]byte, error) {
	for _, v := range obj {
		gob.Register(v)
	}
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(obj)
	if err != nil {
		return []byte(""), err
	}
	return buf.Bytes(), nil
}

func DecodeGob(encoded []byte) (map[interface{}]interface{}, error) {
	buf := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(buf)
	var out map[interface{}]interface{}
	err := dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func generateRandomKey(strength int) []byte {
	k := make([]byte, strength)
	if n, err := io.ReadFull(rand.Reader, k); n != strength || err != nil {
		return RandomCreateBytes(strength)
	}
	return k
}

func encrypt(block cipher.Block, value []byte) ([]byte, error) {
	iv := generateRandomKey(block.BlockSize())
	if iv == nil {
		return nil, errors.New("encrypt: failed to generate random iv")
	}
	// Encrypt it.
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(value, value)
	// Return iv + ciphertext.
	return append(iv, value...), nil
}

func decrypt(block cipher.Block, value []byte) ([]byte, error) {
	size := block.BlockSize()
	if len(value) > size {
		iv := value[:size]
		value = value[size:]
		stream := cipher.NewCTR(block, iv)
		stream.XORKeyStream(value, value)
		return value, nil
	}
	return nil, errors.New("decrypt: the value could not be decrypted")
}

func encodeCookie(block cipher.Block, hashKey, name string, value map[interface{}]interface{}) (string, error) {
	var err error
	var b []byte
	if b, err = EncodeGob(value); err != nil {
		return "", err
	}
	if b, err = encrypt(block, b); err != nil {
		return "", err
	}
	b = encode(b)
	b = []byte(fmt.Sprintf("%s|%d|%s|", name, time.Now().UTC().Unix(), b))
	h := hmac.New(sha1.New, []byte(hashKey))
	h.Write(b)
	sig := h.Sum(nil)
	b = append(b, sig...)[len(name)+1:]
	b = encode(b)
	return string(b), nil
}

func decodeCookie(block cipher.Block, hashKey, name, value string, gcmaxlifetime int64) (map[interface{}]interface{}, error) {
	b, err := decode([]byte(value))
	if err != nil {
		return nil, err
	}
	parts := bytes.SplitN(b, []byte("|"), 3)
	if len(parts) != 3 {
		return nil, errors.New("decode: invalid value format")
	}

	b = append([]byte(name+"|"), b[:len(b)-len(parts[2])]...)
	h := hmac.New(sha1.New, []byte(hashKey))
	h.Write(b)
	sig := h.Sum(nil)
	if len(sig) != len(parts[2]) || subtle.ConstantTimeCompare(sig, parts[2]) != 1 {
		return nil, errors.New("decode: the value is not valid")
	}
	var t1 int64
	if t1, err = strconv.ParseInt(string(parts[0]), 10, 64); err != nil {
		return nil, errors.New("decode: invalid timestamp")
	}
	t2 := time.Now().UTC().Unix()
	if t1 > t2 {
		return nil, errors.New("decode: timestamp is too new")
	}
	if t1 < t2-gcmaxlifetime {
		return nil, errors.New("decode: expired timestamp")
	}
	b, err = decode(parts[1])
	if err != nil {
		return nil, err
	}
	if b, err = decrypt(block, b); err != nil {
		return nil, err
	}
	dst, err := DecodeGob(b)
	if err != nil {
		return nil, err
	}
	return dst, nil
}

func encode(value []byte) []byte {
	encoded := make([]byte, base64.URLEncoding.EncodedLen(len(value)))
	base64.URLEncoding.Encode(encoded, value)
	return encoded
}

func decode(value []byte) ([]byte, error) {
	decoded := make([]byte, base64.URLEncoding.DecodedLen(len(value)))
	b, err := base64.URLEncoding.Decode(decoded, value)
	if err != nil {
		return nil, err
	}
	return decoded[:b], nil
}

var alphaNum = []byte(`0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`)

func RandomCreateBytes(n int, alphabets ...byte) []byte {
	if len(alphabets) == 0 {
		alphabets = alphaNum
	}
	var bytesA = make([]byte, n)
	var randBy bool
	if num, err := rand.Read(bytesA); num != n || err != nil {
		r.Seed(time.Now().UnixNano())
		randBy = true
	}
	for i, b := range bytesA {
		if randBy {
			bytesA[i] = alphabets[r.Intn(len(alphabets))]
		} else {
			bytesA[i] = alphabets[b%byte(len(alphabets))]
		}
	}
	return bytesA
}

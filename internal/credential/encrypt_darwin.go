//go:build darwin

package credential

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// Chrome macOS encryption constants
	chromeSalt       = "saltysalt"
	chromeIterations = 1003
	chromeKeyLen     = 16 // AES-128
	chromeV10Prefix  = "v10"
)

// chromeIV is a fixed 16-byte IV of space characters (0x20).
var chromeIV = bytes.Repeat([]byte{0x20}, 16)

// getChromeKeychainPassword retrieves the Chrome Safe Storage password from macOS Keychain.
func getChromeKeychainPassword() (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", "Chrome Safe Storage",
		"-a", "Chrome",
		"-w",
	)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("cannot read Chrome Safe Storage from Keychain: %w (you may need to allow access)", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// deriveKey derives the AES-128 key from the Keychain password using PBKDF2.
func deriveKey(keychainPassword string) []byte {
	return pbkdf2.Key([]byte(keychainPassword), []byte(chromeSalt), chromeIterations, chromeKeyLen, sha1.New)
}

// EncryptPassword encrypts a plaintext password into Chrome's v10 format.
// Returns the encrypted blob: "v10" + AES-128-CBC(plaintext with PKCS7 padding).
func EncryptPassword(plaintext string, aesKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create AES cipher: %w", err)
	}

	// PKCS7 padding
	padded := pkcs7Pad([]byte(plaintext), aes.BlockSize)

	// Encrypt with AES-128-CBC
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, chromeIV)
	mode.CryptBlocks(ciphertext, padded)

	// Prepend "v10" prefix
	result := make([]byte, 0, len(chromeV10Prefix)+len(ciphertext))
	result = append(result, []byte(chromeV10Prefix)...)
	result = append(result, ciphertext...)
	return result, nil
}

// DecryptPassword decrypts a Chrome v10-encrypted password blob.
func DecryptPassword(encrypted []byte, aesKey []byte) (string, error) {
	if len(encrypted) < 3 || string(encrypted[:3]) != chromeV10Prefix {
		// Not v10 encrypted — return as-is (may be plaintext)
		return string(encrypted), nil
	}

	ciphertext := encrypted[3:] // Strip "v10" prefix
	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("invalid ciphertext length: %d", len(ciphertext))
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("cannot create AES cipher: %w", err)
	}

	decrypted := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, chromeIV)
	mode.CryptBlocks(decrypted, ciphertext)

	// Remove PKCS7 padding
	unpadded, err := pkcs7Unpad(decrypted)
	if err != nil {
		return "", fmt.Errorf("cannot unpad decrypted data: %w", err)
	}
	return string(unpadded), nil
}

// GetChromeEncryptionKey retrieves and derives the AES key from macOS Keychain.
func GetChromeEncryptionKey() ([]byte, error) {
	password, err := getChromeKeychainPassword()
	if err != nil {
		return nil, err
	}
	return deriveKey(password), nil
}

// IsV10Encrypted checks whether a password blob is Chrome v10 encrypted.
func IsV10Encrypted(data []byte) bool {
	return len(data) >= 3 && string(data[:3]) == chromeV10Prefix
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padBytes := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padBytes...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	return data[:len(data)-padding], nil
}

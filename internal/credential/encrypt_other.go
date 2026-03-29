//go:build !darwin

package credential

import "fmt"

// GetChromeEncryptionKey is not yet implemented for this platform.
func GetChromeEncryptionKey() ([]byte, error) {
	return nil, fmt.Errorf("Chrome password encryption is not yet supported on this platform")
}

// EncryptPassword is not yet implemented for this platform.
func EncryptPassword(plaintext string, aesKey []byte) ([]byte, error) {
	return nil, fmt.Errorf("Chrome password encryption is not yet supported on this platform")
}

// DecryptPassword is not yet implemented for this platform.
func DecryptPassword(encrypted []byte, aesKey []byte) (string, error) {
	return "", fmt.Errorf("Chrome password decryption is not yet supported on this platform")
}

// IsV10Encrypted checks whether a password blob is Chrome v10 encrypted.
func IsV10Encrypted(data []byte) bool {
	return len(data) >= 3 && string(data[:3]) == "v10"
}

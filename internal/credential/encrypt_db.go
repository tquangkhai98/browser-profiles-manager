package credential

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

// EncryptLoginDB reads Login Data at dbPath, encrypts any plaintext passwords
// using Chrome's v10 format, and writes the encrypted values back.
// Returns the number of passwords that were encrypted.
func EncryptLoginDB(dbPath string) (int, error) {
	// Verify file exists
	if _, err := os.Stat(dbPath); err != nil {
		return 0, fmt.Errorf("Login Data not found at %s: %w", dbPath, err)
	}

	// Get the Chrome encryption key from Keychain
	aesKey, err := GetChromeEncryptionKey()
	if err != nil {
		return 0, fmt.Errorf("cannot get encryption key: %w", err)
	}

	// Open database in read-write mode
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, fmt.Errorf("cannot open Login Data: %w", err)
	}
	defer db.Close()

	// Read all logins with non-empty password_value
	rows, err := db.Query("SELECT id, password_value FROM logins WHERE length(password_value) > 0")
	if err != nil {
		return 0, fmt.Errorf("cannot query logins: %w", err)
	}

	type loginRow struct {
		id            int
		passwordValue []byte
	}

	var toEncrypt []loginRow
	for rows.Next() {
		var row loginRow
		if err := rows.Scan(&row.id, &row.passwordValue); err != nil {
			continue
		}
		// Skip already encrypted passwords (v10 prefix)
		if IsV10Encrypted(row.passwordValue) {
			continue
		}
		toEncrypt = append(toEncrypt, row)
	}
	rows.Close()

	if len(toEncrypt) == 0 {
		return 0, nil
	}

	// Encrypt and update each plaintext password
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("cannot begin transaction: %w", err)
	}

	stmt, err := tx.Prepare("UPDATE logins SET password_value = ? WHERE id = ?")
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("cannot prepare update statement: %w", err)
	}
	defer stmt.Close()

	encrypted := 0
	for _, row := range toEncrypt {
		plaintext := string(row.passwordValue)
		encryptedValue, err := EncryptPassword(plaintext, aesKey)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("cannot encrypt password for id %d: %w", row.id, err)
		}
		if _, err := stmt.Exec(encryptedValue, row.id); err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("cannot update password for id %d: %w", row.id, err)
		}
		encrypted++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("cannot commit transaction: %w", err)
	}

	return encrypted, nil
}

// EncryptProfileLogins encrypts all plaintext passwords in a profile's Login Data.
// It searches both Default/ and root locations.
func EncryptProfileLogins(profileDir string) (int, error) {
	total := 0
	for _, dbPath := range findAllDBPaths(profileDir, "Login Data") {
		n, err := EncryptLoginDB(dbPath)
		if err != nil {
			return total, fmt.Errorf("encrypt %s: %w", dbPath, err)
		}
		total += n
	}
	return total, nil
}

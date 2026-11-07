package credential

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

// SiteCredential represents credentials found for a single domain.
type SiteCredential struct {
	Domain      string `json:"domain"`
	CookieCount int    `json:"cookie_count"`
	LoginCount  int    `json:"login_count"`
}

// InspectResult holds the credential inspection results for a profile.
type InspectResult struct {
	ProfileName  string           `json:"profile_name"`
	Sites        []SiteCredential `json:"sites"`
	TotalCookies int              `json:"total_cookies"`
	TotalLogins  int              `json:"total_logins"`
}

// Inspect reads the Chromium cookie and login databases from a profile directory.
// It only reads domain names and counts — never decrypts values.
//
// Chromium uses --user-data-dir=<profileDir> and stores data in <profileDir>/Default/.
// However, synced credential files may sit at the profile root. This function checks
// ALL candidate locations and merges the results.
func Inspect(profileDir, profileName string) (*InspectResult, error) {
	result := &InspectResult{ProfileName: profileName}
	siteMap := make(map[string]*SiteCredential)
	attempted, failed := 0, 0
	var lastErr error

	// Read cookies from all candidate locations
	for _, cookieDB := range findAllDBPaths(profileDir, "Cookies") {
		attempted++
		cookies, err := readCookieDomains(cookieDB)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot read cookies DB %s: %v\n", cookieDB, err)
			failed++
			lastErr = fmt.Errorf("%s: %w", cookieDB, err)
			continue
		}
		for domain, count := range cookies {
			if s, ok := siteMap[domain]; ok {
				s.CookieCount += count
			} else {
				siteMap[domain] = &SiteCredential{Domain: domain, CookieCount: count}
			}
			result.TotalCookies += count
		}
	}

	// Read logins from all candidate locations
	for _, loginDB := range findAllDBPaths(profileDir, "Login Data") {
		attempted++
		logins, err := readLoginDomains(loginDB)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot read logins DB %s: %v\n", loginDB, err)
			failed++
			lastErr = fmt.Errorf("%s: %w", loginDB, err)
			continue
		}
		for domain, count := range logins {
			if s, ok := siteMap[domain]; ok {
				s.LoginCount += count
			} else {
				siteMap[domain] = &SiteCredential{Domain: domain, LoginCount: count}
			}
			result.TotalLogins += count
		}
	}

	// Don't report "0 credentials" when nothing could actually be read.
	if attempted > 0 && failed == attempted {
		return nil, fmt.Errorf("cannot read any credential database in %s (is the browser running?): %w", profileDir, lastErr)
	}

	// Convert map to sorted slice
	for _, s := range siteMap {
		result.Sites = append(result.Sites, *s)
	}
	sort.Slice(result.Sites, func(i, j int) bool {
		return result.Sites[i].Domain < result.Sites[j].Domain
	})

	return result, nil
}

// findDBPath looks for a Chromium DB file, preferring Default/ subdirectory.
func findDBPath(profileDir, dbName string) string {
	candidates := []string{
		filepath.Join(profileDir, "Default", dbName),
		filepath.Join(profileDir, dbName),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// findAllDBPaths returns ALL existing paths for a Chromium DB file.
// Chromium stores active data in Default/ but synced files may be at root.
// This returns both so callers can merge results.
func findAllDBPaths(profileDir, dbName string) []string {
	candidates := []string{
		filepath.Join(profileDir, "Default", dbName),
		filepath.Join(profileDir, dbName),
	}
	var found []string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			found = append(found, p)
		}
	}
	return found
}

// readOnlyDSN builds a sqlite URI that modernc.org/sqlite honours: query
// parameters are only parsed for "file:" URIs, so a bare "path?mode=ro" would
// be opened read-write. immutable=1 also stops SQLite from taking locks or
// replaying a hot journal into the live browser database.
func readOnlyDSN(dbPath string) string {
	p := filepath.ToSlash(dbPath)
	if filepath.VolumeName(dbPath) != "" && !strings.HasPrefix(p, "/") {
		p = "/" + p // Windows drive path: file:/C:/... form
	}
	return "file:" + (&url.URL{Path: p}).String() + "?mode=ro&immutable=1"
}

func readCookieDomains(dbPath string) (map[string]int, error) {
	db, err := sql.Open("sqlite", readOnlyDSN(dbPath))
	if err != nil {
		return nil, fmt.Errorf("cannot open cookies DB: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT host_key, COUNT(*) FROM cookies GROUP BY host_key")
	if err != nil {
		return nil, fmt.Errorf("cannot query cookies: %w", err)
	}
	defer rows.Close()

	domains := make(map[string]int)
	for rows.Next() {
		var domain string
		var count int
		if err := rows.Scan(&domain, &count); err != nil {
			return nil, fmt.Errorf("cannot read row: %w", err)
		}
		domains[domain] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot read rows: %w", err)
	}
	return domains, nil
}

func readLoginDomains(dbPath string) (map[string]int, error) {
	db, err := sql.Open("sqlite", readOnlyDSN(dbPath))
	if err != nil {
		return nil, fmt.Errorf("cannot open login DB: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT signon_realm, COUNT(*) FROM logins GROUP BY signon_realm")
	if err != nil {
		return nil, fmt.Errorf("cannot query logins: %w", err)
	}
	defer rows.Close()

	domains := make(map[string]int)
	for rows.Next() {
		var domain string
		var count int
		if err := rows.Scan(&domain, &count); err != nil {
			return nil, fmt.Errorf("cannot read row: %w", err)
		}
		domains[domain] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot read rows: %w", err)
	}
	return domains, nil
}

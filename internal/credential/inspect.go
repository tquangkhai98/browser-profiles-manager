package credential

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"

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
	ProfileName string           `json:"profile_name"`
	Sites       []SiteCredential `json:"sites"`
	TotalCookies int             `json:"total_cookies"`
	TotalLogins  int             `json:"total_logins"`
}

// Inspect reads the Chromium cookie and login databases from a profile directory.
// It only reads domain names and counts — never decrypts values.
func Inspect(profileDir, profileName string) (*InspectResult, error) {
	result := &InspectResult{ProfileName: profileName}
	siteMap := make(map[string]*SiteCredential)

	// Read cookies database
	cookieDB := findDBPath(profileDir, "Cookies")
	if cookieDB != "" {
		cookies, err := readCookieDomains(cookieDB)
		if err != nil {
			// Non-fatal: log warning but continue
			fmt.Fprintf(os.Stderr, "Warning: cannot read cookies DB: %v\n", err)
		} else {
			for domain, count := range cookies {
				if s, ok := siteMap[domain]; ok {
					s.CookieCount = count
				} else {
					siteMap[domain] = &SiteCredential{Domain: domain, CookieCount: count}
				}
				result.TotalCookies += count
			}
		}
	}

	// Read logins database
	loginDB := findDBPath(profileDir, "Login Data")
	if loginDB != "" {
		logins, err := readLoginDomains(loginDB)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot read logins DB: %v\n", err)
		} else {
			for domain, count := range logins {
				if s, ok := siteMap[domain]; ok {
					s.LoginCount = count
				} else {
					siteMap[domain] = &SiteCredential{Domain: domain, LoginCount: count}
				}
				result.TotalLogins += count
			}
		}
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

// findDBPath looks for a Chromium DB file in common subdirectory patterns.
func findDBPath(profileDir, dbName string) string {
	// Chromium stores DBs in Default/ or the profile root
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

func readCookieDomains(dbPath string) (map[string]int, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
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
			continue
		}
		domains[domain] = count
	}
	return domains, nil
}

func readLoginDomains(dbPath string) (map[string]int, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
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
			continue
		}
		domains[domain] = count
	}
	return domains, nil
}

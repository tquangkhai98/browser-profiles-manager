package profile

import "time"

// Profile represents a single isolated browser profile.
type Profile struct {
	Name      string     `json:"name"`
	Browser   string     `json:"browser"`
	DataDir   string     `json:"data_dir"`
	CreatedAt time.Time  `json:"created_at"`
	LastUsed  *time.Time `json:"last_used_at,omitempty"`
}

// ProfileStatus extends Profile with runtime lock information.
type ProfileStatus struct {
	Profile
	Locked   bool      `json:"locked"`
	LockInfo *LockInfo `json:"lock_info,omitempty"`
}

package sessionstore

import (
	"time"
)

type Token struct {
	Access          string
	Refresh         string
	AccessDeadline  int64
	RefreshDeadline int64
	Deadline        int64
}

func NewToken(access, refresh string,
	accessDeadline, refreshDeadline int64,
	deadline int64) *Token {
	return &Token{
		Access:          access,
		Refresh:         refresh,
		AccessDeadline:  accessDeadline,
		RefreshDeadline: refreshDeadline,
		Deadline:        deadline,
	}
}

func (t *Token) IsExpired() bool {
	return time.Now().Unix() > t.AccessDeadline
}
func (t *Token) IsDeleted() bool {
	return time.Now().Unix() > t.RefreshDeadline
}
func (t *Token) CanRefresh() bool {
	if t.Deadline == 0 {
		return true
	}
	return time.Now().Unix() <= t.Deadline
}

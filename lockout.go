package lockout

import (
	"sync"
	"time"
)

type Attempts struct {
	lastAttemptTime time.Time
	count           int
}

type lockoutStruct struct {
	mutex          sync.Mutex
	attempts       map[string]Attempts
	maxRetries     int
	retryTime      time.Duration
	expirationTime time.Duration
}

// NewLockout creates a lockout instance.
func NewLockout(maxRetries int, retryTime time.Duration,
	expirationTime time.Duration) *lockoutStruct {
	return &lockoutStruct{
		attempts:       make(map[string]Attempts),
		maxRetries:     maxRetries,
		retryTime:      retryTime,
		expirationTime: expirationTime,
	}
}

func (pl *lockoutStruct) IsLockedOut(identifer string) bool {
	pl.mutex.Lock()
	defer pl.mutex.Unlock()

	ua, ok := pl.attempts[identifer]
	if !ok {
		return false
	}

	if time.Since(ua.lastAttemptTime) < pl.retryTime && ua.count >= pl.maxRetries {
		return true
	}

	return false
}

func (pl *lockoutStruct) RecordAttempt(identifer string) {
	pl.mutex.Lock()
	defer pl.mutex.Unlock()

	ua, ok := pl.attempts[identifer]
	if !ok {
		ua = Attempts{}
	}

	if time.Since(ua.lastAttemptTime) > pl.retryTime {
		ua.count = 1
	} else {
		ua.count++
	}

	ua.lastAttemptTime = time.Now()

	pl.attempts[identifer] = ua

	// Check for and remove expired attempts
	for user, ua := range pl.attempts {
		if time.Since(ua.lastAttemptTime) > pl.expirationTime {
			delete(pl.attempts, user)
		}
	}
}

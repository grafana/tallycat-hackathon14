package schema

import (
	"sync"
	"sync/atomic"
	"time"
)

// StringInterner provides thread-safe string interning to reduce memory usage
type StringInterner struct {
	mu          sync.RWMutex
	values      map[string]string
	hits        int64
	misses      int64
	size        int64
	maxSize     int64
	lastCleanup time.Time
}

// NewStringInterner creates a new string interner with the specified size limit
func NewStringInterner(maxSize int64) *StringInterner {
	return &StringInterner{
		values:      make(map[string]string),
		maxSize:     maxSize,
		lastCleanup: time.Now(),
	}
}

// Intern interns a string to reduce memory usage
func (i *StringInterner) Intern(s string) string {
	// Fast path: check if we need cleanup
	if i.size >= i.maxSize && time.Since(i.lastCleanup) > time.Hour {
		i.cleanup()
	}

	i.mu.RLock()
	if interned, ok := i.values[s]; ok {
		atomic.AddInt64(&i.hits, 1)
		i.mu.RUnlock()
		return interned
	}
	i.mu.RUnlock()

	i.mu.Lock()
	defer i.mu.Unlock()

	// Double-check after acquiring write lock
	if interned, ok := i.values[s]; ok {
		atomic.AddInt64(&i.hits, 1)
		return interned
	}

	// Check if we have space
	if i.size >= i.maxSize {
		i.cleanup()
	}

	i.values[s] = s
	atomic.AddInt64(&i.size, 1)
	atomic.AddInt64(&i.misses, 1)
	return s
}

// cleanup removes old entries to free up space
func (i *StringInterner) cleanup() {
	// Remove 20% of entries, starting with the oldest
	targetSize := int64(float64(i.maxSize) * 0.8)
	if i.size <= targetSize {
		return
	}

	// Simple cleanup: remove all entries
	// In a real implementation, we would use a more sophisticated strategy
	// like LRU or time-based eviction
	i.values = make(map[string]string)
	i.size = 0
	i.lastCleanup = time.Now()
}

// Stats returns statistics about the interner
func (i *StringInterner) Stats() map[string]int64 {
	return map[string]int64{
		"hits":   atomic.LoadInt64(&i.hits),
		"misses": atomic.LoadInt64(&i.misses),
		"size":   atomic.LoadInt64(&i.size),
	}
}

// Reset clears all interned values
func (i *StringInterner) Reset() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.values = make(map[string]string)
	i.size = 0
	i.hits = 0
	i.misses = 0
	i.lastCleanup = time.Now()
}

// PreIntern interns a list of common strings
func (i *StringInterner) PreIntern(strings []string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	for _, s := range strings {
		if _, ok := i.values[s]; !ok {
			i.values[s] = s
			atomic.AddInt64(&i.size, 1)
		}
	}
}

// Global field name interner
var fieldInterner = NewStringInterner(10000)

// InternField is a convenience wrapper around the global field interner
func InternField(name string) string {
	return fieldInterner.Intern(name)
}

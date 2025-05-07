package schema

import (
	"sync"
	"time"
)

// CardinalityTracker tracks the cardinality of field values
type CardinalityTracker struct {
	values     map[string]map[string]struct{}
	mu         sync.RWMutex
	threshold  int
	windowSize time.Duration
	lastReset  time.Time
}

// NewCardinalityTracker creates a new cardinality tracker with the given configuration
func NewCardinalityTracker(threshold int, windowSize time.Duration) *CardinalityTracker {
	return &CardinalityTracker{
		values:     make(map[string]map[string]struct{}),
		threshold:  threshold,
		windowSize: windowSize,
		lastReset:  time.Now(),
	}
}

// TrackValue adds a value to the tracker for a given field
func (t *CardinalityTracker) TrackValue(fieldName, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Reset if window has expired
	if time.Since(t.lastReset) > t.windowSize {
		t.reset()
	}

	if t.values[fieldName] == nil {
		t.values[fieldName] = make(map[string]struct{})
	}
	t.values[fieldName][value] = struct{}{}
}

// GetCardinality returns the number of unique values for a field
func (t *CardinalityTracker) GetCardinality(fieldName string) int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.values[fieldName])
}

// IsHighCardinality checks if a field has high cardinality based on the threshold
func (t *CardinalityTracker) IsHighCardinality(fieldName string) bool {
	return t.GetCardinality(fieldName) > t.threshold
}

// GetHighCardinalityFields returns all fields that have high cardinality
func (t *CardinalityTracker) GetHighCardinalityFields() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var highCardinalityFields []string
	for fieldName := range t.values {
		if t.IsHighCardinality(fieldName) {
			highCardinalityFields = append(highCardinalityFields, fieldName)
		}
	}
	return highCardinalityFields
}

// Reset clears all tracked values
func (t *CardinalityTracker) reset() {
	t.values = make(map[string]map[string]struct{})
	t.lastReset = time.Now()
}

// GetFieldStats returns statistics about a field's cardinality
func (t *CardinalityTracker) GetFieldStats(fieldName string) FieldStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	cardinality := len(t.values[fieldName])
	return FieldStats{
		FieldName:         fieldName,
		Cardinality:       cardinality,
		IsHighCardinality: cardinality > t.threshold,
		LastUpdated:       t.lastReset,
	}
}

// FieldStats contains statistics about a field's cardinality
type FieldStats struct {
	FieldName         string
	Cardinality       int
	IsHighCardinality bool
	LastUpdated       time.Time
}

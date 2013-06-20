package credentials

import (
	"testing"
	"time"
)

func TestInvalidExpiryTime(t *testing.T) {
	duration := calculateRefreshDuration("invalid")

	if duration != defaultFetchDuration {
		t.Fatalf("Expected duration to be %d, got %d", defaultFetchDuration, duration)
	}
}

func TestPastExpiryTime(t *testing.T) {
	past := time.Date(2000, time.November, 10, 23, 0, 0, 0, time.UTC)
	duration := calculateRefreshDuration(past.Format(time.RFC3339))

	if duration != 1*time.Second {
		t.Fatalf("Expected duration for time in the past to be 1, instead got %s", duration)
	}
}

func TestFutureExpiryTime(t *testing.T) {
	future := time.Now().Add(10 * time.Second)
	duration := calculateRefreshDuration(future.Format(time.RFC3339))

	// Check duration within 1s tolerance
	if !(duration > 9*time.Second && duration < 11*time.Second) {
		t.Fatalf("Expected duration with 1s of 10s, got %s", duration)
	}
}

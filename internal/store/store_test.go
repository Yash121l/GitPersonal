package store

import "testing"

func TestRepositoryLeaseKeyNormalizesOwnerAndName(t *testing.T) {
	t.Parallel()

	first := RepositoryLeaseKey(" Yash ", " Forge ")
	second := RepositoryLeaseKey("yash", "forge")
	other := RepositoryLeaseKey("yash", "other")

	if first != second {
		t.Fatalf("expected normalized repository lease keys to match: %d != %d", first, second)
	}
	if first == other {
		t.Fatalf("expected different repositories to produce different lease keys: %d", first)
	}
}

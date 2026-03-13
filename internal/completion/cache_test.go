package completion

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := &Cache{
		entries: make(map[string]cacheEntry),
	}

	// Test setting and getting a value
	testData := []string{"item1", "item2", "item3"}
	cache.Set("test_key", testData, 1*time.Second)

	result, found := cache.Get("test_key")
	if !found {
		t.Error("Expected to find cached value")
	}
	if len(result) != len(testData) {
		t.Errorf("Expected %d items, got %d", len(testData), len(result))
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := &Cache{
		entries: make(map[string]cacheEntry),
	}

	// Test that expired entries are not returned
	testData := []string{"item1"}
	cache.Set("expired_key", testData, 50*time.Millisecond)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	_, found := cache.Get("expired_key")
	if found {
		t.Error("Expected expired entry to not be found")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := &Cache{
		entries: make(map[string]cacheEntry),
	}

	cache.Set("key1", []string{"a"}, 1*time.Hour)
	cache.Set("key2", []string{"b"}, 1*time.Hour)

	cache.Clear()

	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")

	if found1 || found2 {
		t.Error("Expected cache to be empty after Clear()")
	}
}

func TestCache_Cleanup(t *testing.T) {
	cache := &Cache{
		entries: make(map[string]cacheEntry),
	}

	cache.Set("expired", []string{"a"}, 50*time.Millisecond)
	cache.Set("valid", []string{"b"}, 1*time.Hour)

	// Wait for first entry to expire
	time.Sleep(100 * time.Millisecond)

	cache.Cleanup()

	_, foundExpired := cache.Get("expired")
	_, foundValid := cache.Get("valid")

	if foundExpired {
		t.Error("Expected expired entry to be cleaned up")
	}
	if !foundValid {
		t.Error("Expected valid entry to still exist")
	}
}

func TestFilterByPrefix(t *testing.T) {
	items := []string{"apple", "apricot", "banana", "cherry"}

	// Test with prefix
	result := filterByPrefix(items, "ap")
	if len(result) != 2 {
		t.Errorf("Expected 2 items starting with 'ap', got %d", len(result))
	}

	// Test with empty prefix
	result = filterByPrefix(items, "")
	if len(result) != 4 {
		t.Errorf("Expected all 4 items with empty prefix, got %d", len(result))
	}

	// Test with no matches
	result = filterByPrefix(items, "xyz")
	if len(result) != 0 {
		t.Errorf("Expected 0 items with 'xyz' prefix, got %d", len(result))
	}
}

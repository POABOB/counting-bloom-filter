package counting_bloom_filter

import (
	"fmt"
	"testing"
	"time"
)

var (
	testDuration = 300 * time.Millisecond
)

func TestAddAndCheckWithDefault(t *testing.T) {
	cbf := NewDefaultCountingBloomFilter()
	// Test adding an item
	cbf.Add("item1")
	if !cbf.Check("item1") {
		t.Errorf("Expected item1 to be found in the Bloom Filter")
	}

	// Test checking an item that has not been added
	if cbf.Check("item2") {
		t.Errorf("Expected item2 not to be found in the Bloom Filter")
	}
}

func TestAddAndCheck(t *testing.T) {
	cbf := NewCountingBloomFilter(DefaultBitSize, WithExpiryDuration(EXPIRY_DURATION, testDuration))

	// Test adding an item
	cbf.Add("item1")
	if !cbf.Check("item1") {
		t.Errorf("Expected item1 to be found in the Bloom Filter")
	}

	// Test checking an item that has not been added
	if cbf.Check("item2") {
		t.Errorf("Expected item2 not to be found in the Bloom Filter")
	}
}

func TestRemove(t *testing.T) {
	cbf := NewCountingBloomFilter(DefaultBitSize, WithOptions(Options{
		ExpiryStrategy: LAZY_EXPIRATION,
		Duration:       testDuration,
	}))

	// Test adding an item
	cbf.Add("item1")
	if !cbf.Check("item1") {
		t.Errorf("Expected item1 to be found in the Bloom Filter")
	}

	// Test removing the item
	cbf.Remove("item1")
	if cbf.Check("item1") {
		t.Errorf("Expected item1 to be removed from the Bloom Filter")
	}
}

func TestExpiry(t *testing.T) {
	cbf := NewCountingBloomFilter(DefaultBitSize, WithOptions(Options{
		ExpiryStrategy: EXPIRY_DURATION,
		Duration:       testDuration,
	}))

	// Test adding an item
	cbf.Add("item1")
	if !cbf.Check("item1") {
		t.Errorf("Expected item1 to be found in the Bloom Filter")
	}

	// Simulate the expiration
	time.Sleep(500 * time.Millisecond)

	// Test that the item has expired and is removed
	if cbf.Check("item1") {
		t.Errorf("Expected item1 to be expired and not found in the Bloom Filter")
	}
}

func TestLazyExpiration(t *testing.T) {
	cbf := NewCountingBloomFilter(DefaultBitSize, WithOptions(Options{
		ExpiryStrategy: LAZY_EXPIRATION,
		Duration:       testDuration,
	}))

	// Test adding an item
	cbf.Add("item1")
	if !cbf.Check("item1") {
		t.Errorf("Expected item1 to be found in the Bloom Filter")
	}

	// Simulate some time passing
	time.Sleep(testDuration)

	// Test that the item is still found even if lazy expiration is used
	if !cbf.Check("item1") {
		t.Errorf("Expected item1 to still be found in the Bloom Filter after lazy expiration")
	}
}

func TestResetExpiration(t *testing.T) {
	cbf := NewCountingBloomFilter(DefaultBitSize, WithOptions(Options{
		ExpiryStrategy: RESET_EVERY_PERIOD,
		Duration:       testDuration,
	}))

	// Test adding an item
	cbf.Add("item1")
	if !cbf.Check("item1") {
		t.Errorf("Expected item1 to be found in the Bloom Filter")
	}

	// Simulate some time passing
	time.Sleep(1 * time.Second)

	// Test that the item has been reset
	if cbf.Check("item1") {
		t.Errorf("Expected item1 to be reset and not found in the Bloom Filter after reset expiration")
	}
}

func TestCleanupWithLazy(t *testing.T) {
	cbf := NewCountingBloomFilter(DefaultBitSize, WithOptions(Options{
		ExpiryStrategy: LAZY_EXPIRATION,
		Duration:       testDuration,
	}))

	// Add several items to the Bloom Filter
	for i := 0; i < 1000; i++ {
		cbf.Add(fmt.Sprintf("item%v", i))
	}
	time.Sleep(2 * time.Millisecond)

	// Check if the lazy expiration did not affect the items immediately
	if !cbf.Check("item1") {
		t.Errorf("Expected item1 to be found in the Bloom Filter after cleanup with lazy expiration")
	}
}

func TestCleanupWithReset(t *testing.T) {
	cbf := NewCountingBloomFilter(DefaultBitSize, WithOptions(Options{
		ExpiryStrategy: RESET_EVERY_PERIOD,
		Duration:       testDuration,
	}))

	// Add several items to the Bloom Filter
	for i := 0; i < 1000; i++ {
		cbf.Add(fmt.Sprintf("item%v", i))
	}

	time.Sleep(500 * time.Millisecond)

	// Check if the reset expiration strategy resets all counters
	if cbf.Check("item1") {
		t.Errorf("Expected item1 to be reset and not found in the Bloom Filter after cleanup with reset")
	}
}

func TestCleanupWithExpiry(t *testing.T) {
	cbf := NewCountingBloomFilter(DefaultBitSize, WithOptions(Options{
		ExpiryStrategy: EXPIRY_DURATION,
		Duration:       testDuration,
	}))

	// Add several items to the Bloom Filter
	for i := 0; i < 1000; i++ {
		cbf.Add(fmt.Sprintf("item%v", i))
	}
	time.Sleep(1 * time.Second)
	// Check if expired items are removed
	if cbf.Check("item1") {
		t.Errorf("Expected item1 to be removed after expiration")
	}
}

package counting_bloom_filter

import (
	_ "embed"
	"hash/fnv"
	"math"
	"sync"
	"time"
)

const (
	// K for detailed error rate table, see http://pages.cs.wisc.edu/~cao/papers/summary-cache/node8.html
	// maps as K in the error rate table, time complexity is O(K)
	K = 12

	// DefaultBitSize is the default size of the bloom filter
	// default value is 1MB
	// if the size of element is 33,333, the false positive rate is 0.00000165
	DefaultBitSize = 1 * 1024 * 1024

	// DefaultCleanIntervalTime is the interval time to clean up the bloom filter.
	DefaultCleanIntervalTime = 60 * time.Second

	// DefaultBatchRate is the rate of batch processing
	DefaultBatchRate = 0.1
)

type CountingBloomFilter struct {
	bitmap []uint8
	size   int

	mu sync.RWMutex // Read/Write lock to protect bitmap and timestampMap

	options      *Options
	expiryTicker *time.Ticker
	currentIndex int                  // Track the current index for batch processing
	timestampMap map[string]time.Time // store the timestamp of the element
}

// NewCountingBloomFilter create a Filter, bits is how many bits will be used, maps is how many hashes for each addition.
// best practices:
// elements - means how many actual elements
// when k = 12, formula: 0.7*(bits/k), bits = 30*elements, the error rate is 0.00000165 = 1.65e-06
func NewCountingBloomFilter(size int, options ...Option) *CountingBloomFilter {
	opts := loadOptions(options...)
	if opts.Duration <= 0 {
		opts.Duration = DefaultCleanIntervalTime
	}
	return newCountingBloomFilter(size, opts)
}

// NewDefaultCountingBloomFilter create a default Counting Bloom Filter
func NewDefaultCountingBloomFilter() *CountingBloomFilter {
	return newCountingBloomFilter(DefaultBitSize, nil)
}

func newCountingBloomFilter(size int, opts *Options) *CountingBloomFilter {
	if opts == nil {
		opts = &Options{
			ExpiryStrategy: NO_EXPIRATION,
			Duration:       DefaultCleanIntervalTime,
		}
	}

	cbf := &CountingBloomFilter{
		bitmap:       make([]uint8, size),
		size:         size,
		options:      opts,
		timestampMap: make(map[string]time.Time),
	}

	if cbf.options.ExpiryStrategy != NO_EXPIRATION {
		ticker := time.NewTicker(cbf.options.Duration)
		cbf.expiryTicker = ticker
		if cbf.options.ExpiryStrategy == LAZY_EXPIRATION { // Start ticker for lazy expiration
			go cbf.cleanupWithLazy()
		} else if cbf.options.ExpiryStrategy == RESET_EVERY_PERIOD { // Start ticker for resetting counters periodically
			go cbf.cleanupWithReset()
		} else { // Start ticker for cleaning up expired items
			go cbf.cleanupWithExpiry()
		}
	}
	return cbf
}

// Add add element into Counting Bloom Filter
func (cbf *CountingBloomFilter) Add(item string) {
	cbf.mu.Lock()
	for i := 0; i < K; i++ {
		hashValue := hash(item, i) % uint64(cbf.size)
		cbf.bitmap[hashValue]++
	}

	// add element to timestamp map
	if cbf.options.ExpiryStrategy == EXPIRY_DURATION {
		cbf.timestampMap[item] = time.Now()
	}
	cbf.mu.Unlock()
}

// Check check element in Counting Bloom Filter
func (cbf *CountingBloomFilter) Check(item string) bool {
	cbf.mu.RLock()
	for i := 0; i < K; i++ {
		hashValue := hash(item, i) % uint64(cbf.size)
		if cbf.bitmap[hashValue] == 0 {
			cbf.mu.RUnlock()
			return false
		}
	}

	// check element in timestamp map
	if cbf.options.ExpiryStrategy == EXPIRY_DURATION {
		if timestamp, exists := cbf.timestampMap[item]; exists {
			if time.Since(timestamp) > cbf.options.Duration {
				cbf.mu.RUnlock()
				cbf.Remove(item)
				return false
			}
		}
	}
	cbf.mu.RUnlock()
	return true
}

// Remove element from Counting Bloom Filter
func (cbf *CountingBloomFilter) Remove(item string) {
	cbf.mu.Lock()
	for i := 0; i < K; i++ {
		hashValue := hash(item, i) % uint64(cbf.size)
		if cbf.bitmap[hashValue] > 0 {
			cbf.bitmap[hashValue]--
		}
	}
	delete(cbf.timestampMap, item)
	cbf.mu.Unlock()
}

// RemoveAll remove all elements from Counting Bloom Filter
func (cbf *CountingBloomFilter) RemoveAll() {
	cbf.mu.Lock()
	for i := 0; i < cbf.size; i++ {
		cbf.bitmap[i] = 0 // Reset the counter for the selected element
	}
	cbf.timestampMap = make(map[string]time.Time)
	cbf.mu.Unlock()
}

// Hash function
func hash(input string, seed int) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(input))
	return h.Sum64() + uint64(seed)
}

// cleanupExpiredItems periodically decrease the counters for all items to simulate lazy expiration
func (cbf *CountingBloomFilter) cleanupWithExpiry() {
	for {
		select {
		case <-cbf.expiryTicker.C:
			evictedCount := 0
			totalCount := 20
			i := 0
			for key := range cbf.timestampMap {
				if key != "" && cbf.Check(key) {
					if cbf.bitmap[hash(key, 0)%uint64(cbf.size)] > 0 {
						cbf.Remove(key) // Remove expired item
						evictedCount++
					}
				}

				// Break the loop after 20 iterations
				i++
				if i >= totalCount {
					break
				}
			}

			// If evicted more than 25%, repeat the process
			if float64(evictedCount)/float64(totalCount) > 0.25 {
				continue // Repeat eviction process
			}
		}
	}
}

// cleanupWithLazy periodically decreases counters in batches to simulate lazy expiration
func (cbf *CountingBloomFilter) cleanupWithLazy() {
	for {
		select {
		case <-cbf.expiryTicker.C:
			cbf.mu.Lock() // Acquire lock before starting the batch cleanup
			batchSize := int(math.Ceil(float64(cbf.size) * DefaultBatchRate))
			if batchSize < 1 {
				batchSize = 1_000
			}
			startIndex := cbf.currentIndex // Start from the current index

			// Process a batch of elements starting from currentIndex
			for i := 0; i < batchSize && startIndex < cbf.size; i++ {
				if cbf.bitmap[startIndex] > 0 {
					cbf.bitmap[startIndex]-- // Gradually decrease the counter to simulate expiration
				}
				startIndex++ // Move to the next index
			}

			// Update currentIndex to continue from the new position for next cleanup
			cbf.currentIndex = startIndex

			// If we've processed all the elements, reset the currentIndex for the next round
			if cbf.currentIndex >= cbf.size {
				cbf.currentIndex = 0 // Reset to start from the beginning for the next round
			}
			cbf.mu.Unlock() // Release the lock after processing the batch
		}
	}
}

// cleanupWithReset periodically resets counters
func (cbf *CountingBloomFilter) cleanupWithReset() {
	for {
		select {
		case <-cbf.expiryTicker.C:
			cbf.RemoveAll()
		}
	}
}

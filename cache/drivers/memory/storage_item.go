package memory

import (
	"time"
)

// storageItem is a cached entry. created and expire are Unix timestamps;
// expire is -1 for entries that never expire.
type storageItem struct {
	created int64
	expire  int64
	hits    int64
	data    []byte
}

// isExpired reports whether the entry has expired.
func (item *storageItem) isExpired() bool {
	return item.expire > 0 && item.expire <= time.Now().Unix()
}

// size returns the size of the entry's data in bytes.
func (item *storageItem) size() int64 {
	return int64(len(item.data))
}

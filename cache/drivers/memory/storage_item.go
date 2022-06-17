package memory

import (
	"time"
)

type storageItem struct {
	created int64
	expire  int64
	hits    int64
	data    []byte
}

func (item *storageItem) isExpired() bool {
	return item.expire > 0 && item.expire <= time.Now().Unix()
}

func (item *storageItem) size() int64 {
	return int64(len(item.data))
}

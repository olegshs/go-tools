package memory

import (
	"time"
)

type storageIndex []storageIndexItem

type storageIndexItem struct {
	key   string
	value int64
}

func newIndex(items map[string]*storageItem) storageIndex {
	now := time.Now().Unix()

	index := make(storageIndex, len(items))
	i := 0

	for key, item := range items {
		indexItem := &index[i]
		indexItem.key = key

		if item.isExpired() {
			indexItem.value = 0
		} else {
			d := now - item.created
			if d <= 0 {
				d = 1
			}

			// hits per hour
			indexItem.value = (item.hits + 1) * 3600 / d
		}

		i++
	}

	return index
}

func (index storageIndex) Len() int {
	return len(index)
}

func (index storageIndex) Swap(i, j int) {
	index[i], index[j] = index[j], index[i]
}

func (index storageIndex) Less(i, j int) bool {
	return index[i].value < index[j].value
}

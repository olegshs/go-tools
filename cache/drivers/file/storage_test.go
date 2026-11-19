package file

import (
	"fmt"
	"sync"
	"testing"
)

func TestStorage_Concurrent(t *testing.T) {
	conf := DefaultConfig()
	conf.Dir = t.TempDir()
	stor := NewStorage(conf)

	const n = 20

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			key := fmt.Sprintf("key%d", i)

			err := stor.Set(key, []byte(key), 0)
			if err != nil {
				t.Error(err)
				return
			}

			data, err := stor.Get(key)
			if err != nil {
				t.Error(err)
				return
			}

			if string(data) != key {
				t.Errorf("Get(%q) = %q", key, data)
			}
		}(i)
	}
	wg.Wait()

	if hits := stor.Hits(); hits != n {
		t.Errorf("Hits() = %d, expected %d", hits, n)
	}
}

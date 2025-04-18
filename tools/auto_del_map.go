package tools

import (
	"sync"
	"time"
)

type AutoDeleteMap struct {
	sync.Map
}

func (m *AutoDeleteMap) Set(key, value interface{}, t time.Duration) {
	m.Store(key, value)
	GoSafe(func() {
		time.Sleep(t)
		m.Delete(key)
	})
}

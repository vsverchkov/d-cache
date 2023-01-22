package cache

import (
    "sync"
    "time"
)

type Cache struct {
    lock sync.RWMutex
    data map[string][]byte

}

func New() *Cache {
    return &Cache{
        data: make(map[string][]byte),
    }
}

func (c *Cache) Set(key []byte, value []byte, ttl time.Duration) error {
    c.lock.Lock()
    defer c.lock.Unlock()

    c.data[string(key)] = value

    if ttl > 0 {
        go func() {
            <-time.After(ttl)
            c.lock.Lock()
            defer c.lock.Unlock()
            delete(c.data, string(key))
        }()
    }

    return nil
}

func (c *Cache) Has(key []byte) (bool, error) {
    c.lock.RLock()
    defer c.lock.RUnlock()

    if _, ok := c.data[string(key)]; ok {
        return true, nil
    }

    return false, nil
}

func (c *Cache) Get(key []byte) ([]byte, error)  {
    c.lock.RLock()
    defer c.lock.RUnlock()

    keyStr := string(key)

    val, ok := c.data[keyStr]
    if !ok {
        return []byte{}, nil
    }
    return val, nil
}

func (c *Cache) Delete(key []byte) (bool, error) {
    c.lock.Lock()
    defer c.lock.Unlock()

    delete(c.data, string(key))

    return true, nil
}
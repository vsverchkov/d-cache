package cache

import "time"

type Cacher interface {
    Set([]byte, []byte, time.Duration) error
    Has([]byte) (bool, error)
    Get([]byte) ([]byte, error)
    Delete([]byte) (bool, error)
}
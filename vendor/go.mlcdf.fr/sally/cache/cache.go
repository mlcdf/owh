package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"go.mlcdf.fr/sally/logging"
	"golang.org/x/xerrors"
)

var DefaultCache *Cache

type expirableValue struct {
	Value      string    `json:"value"`
	Expiration time.Time `json:"expiration"`
}

type Cache struct {
	name            string
	location        string
	kv              map[string]expirableValue
	encryptionKey   []byte
	createCacheOnce sync.Once
}

// New instanciate a new cache object.
// The encryptionKey parameter is here for obscurity purpose.
func New(applicationName, name string, encryptionKey []byte) (*Cache, error) {
	if len(encryptionKey) == 0 {
		encryptionKey = []byte("Dy8ANW03Zz9e=9Ra")
	}

	c := &Cache{
		name:          name,
		encryptionKey: encryptionKey,
	}

	var err error

	c.location, err = xdg.CacheFile(filepath.Join(applicationName, name))
	if err != nil {
		return nil, err
	}

	c.kv, err = c.read()

	return c, err
}

func (c *Cache) Set(key string, value string, expiration time.Duration) error {
	c.kv[key] = expirableValue{value, time.Now().Add(expiration)}

	return c.write(c.kv)
}

func (c *Cache) Get(key string) string {
	ev, ok := c.kv[key]
	if !ok {
		return ""
	}

	if time.Now().After(ev.Expiration) {
		logging.Debugf("expired key %s", key)

		delete(c.kv, key)
		return ""
	}
	return ev.Value
}

func (c *Cache) read() (map[string]expirableValue, error) {
	logging.Debugf("reading cache %s", c.name)

	if c.location == "" {
		return make(map[string]expirableValue), nil
	}

	ciphertext, err := os.ReadFile(c.location)
	if err != nil {
		return nil, xerrors.Errorf("failed to read file %s: %w", c.location, err)
	}

	plaintext, err := decrypt(ciphertext, c.encryptionKey)
	if err != nil {
		return nil, xerrors.Errorf("failed to decrypt: %w", err)
	}

	data := make(map[string]expirableValue)

	json.Unmarshal(plaintext, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Cache) write(data map[string]expirableValue) error {
	var err error

	c.createCacheOnce.Do(func() {
		_, err = os.Open(c.location)
		if errors.Is(err, os.ErrNotExist) {
			var fh *os.File

			fh, err = os.Create(c.location)
			if err != nil {
				return
			}

			defer fh.Close()
		}
	})

	if err != nil {
		return xerrors.Errorf("error creating cache file: %w", err)
	}

	logging.Debugf("write cache %s", c.name)

	plaintext, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	ciphertext, err := encrypt(plaintext, c.encryptionKey)
	if err != nil {
		return xerrors.Errorf("failed to encrypt: %w", err)
	}

	return os.WriteFile(c.location, ciphertext, 0600)
}

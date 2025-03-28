package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/coocood/freecache"
	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
)

type Cache struct {
	rds        *redis.Client
	localcache *freecache.Cache
	sf         *singleflight.Group
}

func NewCache(rds *redis.Client, localcache *freecache.Cache) *Cache {
	return &Cache{
		rds:        rds,
		localcache: localcache,
		sf:         &singleflight.Group{},
	}
}

func (c *Cache) Get(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) ([]byte, error) {
	var err error
	var bytes []byte
	// load from cache
	if bytes, err = c.doGet(ctx, key, ttl); err != nil {
		return nil, err
	}
	if bytes != nil {
		return bytes, nil
	}
	// load from db
	value, err := fn()
	if err != nil {
		return bytes, nil
	}
	bytes, err = json.Marshal(value)
	if err != nil {
		return nil, err
	}
	// set to cache
	if err = c.doSet(ctx, key, bytes, ttl, ttl); err != nil {
		_ = fmt.Errorf("doSet error %s", err.Error())
	}

	return bytes, nil
}

func (c *Cache) singleDoGet(ctx context.Context, key string, ttl time.Duration) ([]byte, error) {
	ch := c.sf.DoChan(key, func() (interface{}, error) {
		bytes, err := c.doGet(ctx, key, ttl)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	})

	select {
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.([]byte), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Cache) doGet(ctx context.Context, key string, ttl time.Duration) ([]byte, error) {
	var valueBytes []byte
	var err error
	valueBytes, err = c.localcache.Get([]byte(key))
	if err != nil && !errors.Is(err, freecache.ErrNotFound) {
		return nil, err
	}
	// local miss
	if valueBytes == nil {
		valueBytes, err = c.rds.Get(ctx, key).Bytes()
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, err
		}
		// redis miss
		if valueBytes == nil {
			return nil, nil
		}
		err = c.localcache.Set([]byte(key), valueBytes, int(ttl))
		if err != nil {
			_ = fmt.Errorf("localcache set error %v", err)
		}
		return valueBytes, nil
	}
	valueBytes, err = c.rds.Get(ctx, key).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	return valueBytes, nil
}

func (c *Cache) doSet(ctx context.Context, key string, bytes []byte, localTTL, remoteTTL time.Duration) error {
	var err error
	if err = c.rds.Set(ctx, key, bytes, remoteTTL).Err(); err != nil {
		return err
	}
	if err = c.localcache.Set([]byte(key), bytes, int(localTTL)); err != nil {
		return err
	}
	return nil
}

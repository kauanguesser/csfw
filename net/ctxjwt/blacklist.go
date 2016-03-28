// Copyright 2015-2016, Cyrill @ Schumacher.fm and the CoreStore contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ctxjwt

import (
	"sync"
	"time"

	"github.com/coocood/freecache"
)

// Blacklister a backend storage to handle blocked tokens.
// Default black hole storage. Must be thread safe.
type Blacklister interface {
	Set(token string, expires time.Duration) error
	Has(token string) bool
}

// nullBL is the black hole black list
type nullBL struct{}

func (b nullBL) Set(_ string, _ time.Duration) error { return nil }
func (b nullBL) Has(_ string) bool                   { return false }

// BlackListSimpleMap creates an in-memory map which holds as a key the
// tokens and as value the token expiration duration. Once a Set() operation
// will be called the tokens list get purged. Don't use this feature in
// production as the underlying mutex will become a bottleneck with higher
// throughput, but still faster as a connection to Redis ;-)
type BlackListSimpleMap struct {
	mu     sync.Mutex
	tokens map[uint64]time.Time
}

// NewBlackListSimpleMap creates a new blacklist map.
func NewBlackListSimpleMap() *BlackListSimpleMap {
	return &BlackListSimpleMap{
		tokens: make(map[uint64]time.Time),
	}
}

// Has checks if token is within the blacklist.
func (bl *BlackListSimpleMap) Has(token string) bool {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	h := hash(token)
	d, ok := bl.tokens[h]
	if !ok {
		return false
	}
	isValid := time.Since(d) < 0

	if false == isValid {
		delete(bl.tokens, h)
	}
	return isValid
}

// Set adds a token to the map and performs a purge operation.
func (bl *BlackListSimpleMap) Set(token string, expires time.Duration) error {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	// this operation is not performance friendly
	for k, v := range bl.tokens {
		if time.Since(v) > 0 {
			delete(bl.tokens, k)
		}
	}
	bl.tokens[hash(token)] = time.Now().Add(expires)
	return nil
}

// Len returns the number of entries in the blacklist
func (bl *BlackListSimpleMap) Len() int {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	return len(bl.tokens)
}

// BlackListFreeCache high performance cache for concurrent/parallel use cases
// like in net/http needed.
type BlackListFreeCache struct {
	*freecache.Cache
	emptyVal []byte
}

// NewBlackListFreeCache creates a new cache instance with a minimum size to be
// set to 512KB.
// If the size is set relatively large, you should call `debug.SetGCPercent()`,
// set it to a much smaller value to limit the memory consumption and GC pause time.
func NewBlackListFreeCache(size int) *BlackListFreeCache {
	return &BlackListFreeCache{
		Cache:    freecache.NewCache(size),
		emptyVal: []byte(`1`),
	}
}

// Set sets a token. If expires <=0 the cached item will not expire.
func (fc *BlackListFreeCache) Set(token string, expires time.Duration) error {
	return fc.Cache.Set([]byte(token), fc.emptyVal, int(expires.Seconds()))
}

// Has checks if the cache contains the token.
func (fc *BlackListFreeCache) Has(token string) bool {
	val, err := fc.Cache.Get([]byte(token))
	if err == freecache.ErrNotFound {
		return false
	}
	if err != nil {
		return false
	}
	return val != nil
}

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fnv implements FNV-1 and FNV-1a, non-cryptographic hash functions
// created by Glenn Fowler, Landon Curt Noll, and Phong Vo.
// See
// https://en.wikipedia.org/wiki/Fowler-Noll-Vo_hash_function.

const (
	offset64 uint64 = 14695981039346656037
	prime64  uint64 = 1099511628211
)

// fnv64a
func hash(data string) uint64 {
	hash := offset64
	for _, c := range []byte(data) {
		hash ^= uint64(c)
		hash *= prime64
	}
	return hash
}

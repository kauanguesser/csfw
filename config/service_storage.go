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

package config

import (
	"sync"

	"github.com/corestoreio/csfw/config/path"
)

// Storager is the underlying data storage for holding the keys and its values.
// Implementations can be spf13/viper or MySQL backed. Default Storager
// is a simple mutex protected map[string]interface{}.
// ProTip: If you use MySQL as Storager don't execute function
// ApplyCoreConfigData()
type Storager interface {
	// Set sets a key with a value and returns on success nil or ErrKeyOverwritten,
	// on failure any other error
	Set(key path.Path, value interface{}) error
	// Get may return a ErrKeyNotFound error
	Get(key path.Path) (interface{}, error)
	// AllKeys returns the fully qualified keys
	AllKeys() (path.PathSlice, error)
}

var _ Storager = (*simpleStorage)(nil)

type keyVal struct {
	k path.Path
	v interface{}
}

type simpleStorage struct {
	sync.Mutex
	kv map[uint32]keyVal
}

func newSimpleStorage() *simpleStorage {
	return &simpleStorage{
		kv: make(map[uint32]keyVal),
	}
}

func (sp *simpleStorage) Set(key path.Path, value interface{}) error {
	sp.Lock()
	defer sp.Unlock()
	k, err := key.FQ()
	if err != nil {
		return err
	}
	h32 := k.Hash32()
	sp.kv[h32] = keyVal{key, value}
	return nil
}

func (sp *simpleStorage) Get(key path.Path) (interface{}, error) {
	sp.Lock()
	defer sp.Unlock()

	k, err := key.FQ()
	if err != nil {
		return nil, err
	}
	h32 := k.Hash32()
	if data, ok := sp.kv[h32]; ok {
		return data.v, nil
	}
	return nil, ErrKeyNotFound
}

func (sp *simpleStorage) AllKeys() (path.PathSlice, error) {
	sp.Lock()
	defer sp.Unlock()

	var ret = make(path.PathSlice, len(sp.kv), len(sp.kv))
	i := 0
	for _, kv := range sp.kv {
		ret[i] = kv.k
		i++
	}
	return ret, nil
}

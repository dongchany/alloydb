//
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package goleveldb

import (
	"sync"

	"github.com/juju/errors"
	"github.com/Dong-Chan/alloydb/store/localstore/engine"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	_ engine.DB       = (*db)(nil)
	_ engine.Batch    = (*leveldb.Batch)(nil)
	_ engine.Snapshot = (*snapshot)(nil)
)

var (
	p = sync.Pool{
		New: func() interface{} {
			return &leveldb.Batch{}
		},
	}
)

type db struct {
	*leveldb.DB
}

func (d *db) Get(key []byte) ([]byte, error) {
	v, err := d.DB.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}

	return v, err
}

func (d *db) GetSnapshot() (engine.Snapshot, error) {
	s, err := d.DB.GetSnapshot()
	if err != nil {
		return nil, err
	}
	return &snapshot{s}, nil
}

func (d *db) NewBatch() engine.Batch {
	b := p.Get().(*leveldb.Batch)
	return b
}

func (d *db) Commit(b engine.Batch) error {
	batch, ok := b.(*leveldb.Batch)
	if !ok {
		return errors.Errorf("invalid batch type %T", b)
	}
	err := d.DB.Write(batch, nil)
	batch.Reset()
	p.Put(batch)
	return err
}

func (d *db) Close() error {
	return d.DB.Close()
}

type snapshot struct {
	*leveldb.Snapshot
}

func (s *snapshot) Get(key []byte) ([]byte, error) {
	v, err := s.Snapshot.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}

	return v, err
}

func (s *snapshot) NewIterator(startKey []byte) engine.Iterator {
	it := s.Snapshot.NewIterator(&util.Range{Start: startKey}, nil)
	return it
}

func (s *snapshot) Release() {
	s.Snapshot.Release()
}

// Driver implements engine Driver.
type Driver struct {
}

// Open opens or creates a local storage database for the given path.
func (driver Driver) Open(path string) (engine.DB, error) {
	d, err := leveldb.OpenFile(path, &opt.Options{BlockCacheCapacity: 600 * 1024 * 1024})

	return &db{d}, err
}

// MemoryDriver implements engine Driver
type MemoryDriver struct {
}

// Open opens a memory storage database.
func (driver MemoryDriver) Open(path string) (engine.DB, error) {
	d, err := leveldb.Open(storage.NewMemStorage(), nil)
	return &db{d}, err
}

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

package localstore

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/Dong-Chan/alloydb/kv"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

var (
	_ kv.Transaction = (*dbTxn)(nil)

	// ErrInvalidTxn is the error when commits or rollbacks in an invalid transaction.
	ErrInvalidTxn = errors.New("invalid transaction")
	// ErrCannotSetNilValue is the error when sets an empty value
	ErrCannotSetNilValue = errors.New("can not set nil value")
)

// dbTxn is not thread safe
type dbTxn struct {
	kv.UnionStore
	store        *dbStore // for commit
	startTs      time.Time
	tID          int64
	opCnt        int64
	valid        bool
	snapshotVals map[string][]byte // origin version in snapshot
}

func (txn *dbTxn) markOrigin(k []byte) error {
	keystr := string(k)

	// Already exist, no nothing
	if _, ok := txn.snapshotVals[keystr]; ok {
		return nil
	}

	val, err := txn.Snapshot.Get(k)
	if err != nil && !kv.IsErrNotFound(err) {
		return err
	}

	//log.Debugf("markOrigin, key:%s, value:%s", keystr, val)
	txn.snapshotVals[keystr] = val
	return nil
}

// Implement transaction interface

func (txn *dbTxn) Inc(k []byte, step int64) (int64, error) {
	log.Debugf("Inc %s, step %d txn:%d", k, step, txn.tID)
	k = kv.EncodeKey(k)

	if err := txn.markOrigin(k); err != nil {
		return 0, err
	}
	val, err := txn.UnionStore.Get(k)
	if kv.IsErrNotFound(err) {
		err = txn.UnionStore.Set(k, []byte(strconv.FormatInt(step, 10)))
		if err != nil {
			return 0, err
		}

		return step, nil
	}

	if err != nil {
		return 0, err
	}

	intVal, err := strconv.ParseInt(string(val), 10, 0)
	if err != nil {
		return intVal, err
	}

	intVal += step
	err = txn.UnionStore.Set(k, []byte(strconv.FormatInt(intVal, 10)))
	if err != nil {
		return 0, err
	}

	return intVal, nil
}

func (txn *dbTxn) String() string {
	return fmt.Sprintf("%d", txn.tID)
}

func (txn *dbTxn) Get(k []byte) ([]byte, error) {
	log.Debugf("get key:%s, txn:%d", k, txn.tID)
	k = kv.EncodeKey(k)
	val, err := txn.UnionStore.Get(k)
	if kv.IsErrNotFound(err) {
		return nil, kv.ErrNotExist
	}
	if err != nil {
		return nil, err
	}

	if len(val) == 0 {
		return nil, kv.ErrNotExist
	}

	return val, nil
}

func (txn *dbTxn) Set(k []byte, data []byte) error {
	if len(data) == 0 {
		// Incase someone use it in the wrong way, we can figure it out immediately
		debug.PrintStack()
		return ErrCannotSetNilValue
	}

	log.Debugf("set key:%s, txn:%d", k, txn.tID)
	k = kv.EncodeKey(k)
	return txn.UnionStore.Set(k, data)
}

func (txn *dbTxn) Seek(k []byte, fnKeyCmp func([]byte) bool) (kv.Iterator, error) {
	log.Debugf("seek %s txn:%d", k, txn.tID)
	k = kv.EncodeKey(k)

	iter, err := txn.UnionStore.Seek(k, txn)
	if err != nil {
		return nil, err
	}

	if !iter.Valid() {
		return &kv.UnionIter{}, nil
	}

	if fnKeyCmp != nil {
		if fnKeyCmp([]byte(iter.Key())[:1]) {
			return &kv.UnionIter{}, nil
		}
	}

	return iter, nil
}

func (txn *dbTxn) Delete(k []byte) error {
	log.Debugf("delete %s txn:%d", k, txn.tID)
	k = kv.EncodeKey(k)
	return txn.UnionStore.Delete(k)
}

func (txn *dbTxn) each(f func(iterator.Iterator) error) error {
	iter := txn.UnionStore.Dirty.NewIterator(nil)
	defer iter.Release()
	for iter.Next() {
		if err := f(iter); err != nil {
			return err
		}
	}
	return nil
}

func (txn *dbTxn) doCommit() error {
	b := txn.store.newBatch()
	keysLocked := make([]string, 0, len(txn.snapshotVals))
	defer func() {
		for _, key := range keysLocked {
			txn.store.unLockKeys(key)
		}
	}()
	// Check locked keys
	for k, v := range txn.snapshotVals {
		err := txn.store.tryConditionLockKey(txn.tID, k, v)
		if err != nil {
			return errors.Trace(err)
		}
		keysLocked = append(keysLocked, k)
	}

	// Check dirty store
	err := txn.each(func(iter iterator.Iterator) error {
		if len(iter.Value()) == 0 { // Deleted marker
			b.Delete(iter.Key())
		} else {
			b.Put(iter.Key(), iter.Value())
		}
		return nil
	})
	if err != nil {
		return errors.Trace(err)
	}
	return txn.store.writeBatch(b)

}

func (txn *dbTxn) Commit() error {
	if !txn.valid {
		return ErrInvalidTxn
	}
	log.Infof("commit txn %d", txn.tID)
	defer func() {
		txn.close()
	}()

	return txn.doCommit()
}

func (txn *dbTxn) close() error {
	txn.UnionStore.Close()
	txn.snapshotVals = nil
	txn.valid = false
	return nil
}

func (txn *dbTxn) Rollback() error {
	if !txn.valid {
		return ErrInvalidTxn
	}
	log.Warnf("Rollback txn %d", txn.tID)
	return txn.close()
}

func (txn *dbTxn) LockKeys(keys ...[]byte) error {
	for _, key := range keys {
		key = kv.EncodeKey(key)
		if err := txn.markOrigin(key); err != nil {
			return err
		}
	}
	return nil
}

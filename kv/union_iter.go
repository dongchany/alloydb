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

package kv

import (
	"bytes"

	"github.com/ngaut/log"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

// UnionIter is the iterator on an UnionStore.
type UnionIter struct {
	dirtyIt       iterator.Iterator
	snapshotIt    Iterator
	curIsDirty    bool
	dirtyValid    bool
	snapshotValid bool

	isValid bool
}

func newUnionIter(dirtyIt iterator.Iterator, snapshotIt Iterator) *UnionIter {
	it := &UnionIter{
		dirtyIt:    dirtyIt,
		snapshotIt: snapshotIt,
		// leveldb use next for checking valid...
		dirtyValid:    dirtyIt.Next(),
		snapshotValid: snapshotIt.Valid(),
	}
	it.updateCur()
	return it
}

// Go next and update valid status.
func (iter *UnionIter) dirtyNext() {
	iter.dirtyValid = iter.dirtyIt.Next()
}

// Go next and update valid status.
func (iter *UnionIter) snapshotNext() {
	iter.snapshotIt, _ = iter.snapshotIt.Next(nil)
	iter.snapshotValid = iter.snapshotIt.Valid()
}

func (iter *UnionIter) updateCur() {
	iter.isValid = true
	for {
		if !iter.dirtyValid && !iter.snapshotValid {
			iter.isValid = false
			return
		}

		if !iter.dirtyValid {
			iter.curIsDirty = false
			return
		}

		if !iter.snapshotValid {
			iter.curIsDirty = true
			// if delete it
			if len(iter.dirtyIt.Value()) == 0 {
				iter.dirtyNext()
				continue
			}
			break
		}

		// both valid
		if iter.snapshotValid && iter.dirtyValid {
			snapshotKey := []byte(iter.snapshotIt.Key())
			dirtyKey := iter.dirtyIt.Key()
			cmp := bytes.Compare(dirtyKey, snapshotKey)
			// if equal, means both have value
			if cmp == 0 {
				if len(iter.dirtyIt.Value()) == 0 {
					// snapshot has a record, but trx says we have deleted it
					// just go next
					iter.dirtyNext()
					iter.snapshotNext()
					continue
				}
				// both go next
				iter.snapshotNext()
				iter.curIsDirty = true
				break
			} else if cmp > 0 {
				// record from snapshot comes first
				iter.curIsDirty = false
				break
			} else {
				// record from dirty comes first
				if len(iter.dirtyIt.Value()) == 0 {
					log.Warnf("delete a record not exists? k = %s", string(iter.dirtyIt.Key()))
					// jump over this deletion
					iter.dirtyNext()
					continue
				}
				iter.curIsDirty = true
				break
			}
		}
	}
}

// Next implements the Iterator Next interface.
func (iter *UnionIter) Next(f FnKeyCmp) (Iterator, error) {
	if iter.curIsDirty == false {
		iter.snapshotNext()
	} else {
		iter.dirtyNext()
	}
	iter.updateCur()
	return iter, nil
}

// Value implements the Iterator Value interface.
// Multi columns
func (iter *UnionIter) Value() []byte {
	if iter.curIsDirty == false {
		return iter.snapshotIt.Value()
	}
	return iter.dirtyIt.Value()
}

// Key implements the Iterator Key interface.
func (iter *UnionIter) Key() string {
	if iter.curIsDirty == false {
		return string(DecodeKey([]byte(iter.snapshotIt.Key())))
	}
	return string(DecodeKey(iter.dirtyIt.Key()))
}

// Valid implements the Iterator Valid interface.
func (iter *UnionIter) Valid() bool {
	return iter.isValid
}

// Close implements the Iterator Close interface.
func (iter *UnionIter) Close() {
	if iter.snapshotIt != nil {
		iter.snapshotIt.Close()
		iter.snapshotIt = nil
	}
	if iter.dirtyIt != nil {
		iter.dirtyIt.Release()
		iter.dirtyIt = nil
	}
}

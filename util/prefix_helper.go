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

package util

import (
	"bytes"
	"strings"

	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/kv"
	"github.com/Dong-Chan/alloydb/util/codec"
)

func hasPrefix(prefix []byte) kv.FnKeyCmp {
	return func(k []byte) bool {
		return bytes.HasPrefix(k, prefix)
	}
}

// ScanMetaWithPrefix scans metadata with the prefix.
func ScanMetaWithPrefix(txn kv.Transaction, prefix string, filter func([]byte, []byte) bool) error {
	iter, err := txn.Seek([]byte(prefix), hasPrefix([]byte(prefix)))
	if err != nil {
		return err
	}
	defer iter.Close()
	for {
		if err != nil {
			return err
		}

		if iter.Valid() && strings.HasPrefix(iter.Key(), prefix) {
			if !filter([]byte(iter.Key()), iter.Value()) {
				break
			}
			iter, err = iter.Next(hasPrefix([]byte(prefix)))
		} else {
			break
		}
	}

	return nil
}

// DelKeyWithPrefix deletes keys with prefix.
func DelKeyWithPrefix(ctx context.Context, prefix string) error {
	log.Debug("delKeyWithPrefix", prefix)
	txn, err := ctx.GetTxn(false)
	if err != nil {
		return err
	}

	var keys []string
	iter, err := txn.Seek([]byte(prefix), hasPrefix([]byte(prefix)))
	if err != nil {
		return err
	}
	defer iter.Close()
	for {
		if err != nil {
			return err
		}

		if iter.Valid() && strings.HasPrefix(iter.Key(), prefix) {
			keys = append(keys, iter.Key())
			iter, err = iter.Next(hasPrefix([]byte(prefix)))
		} else {
			break
		}
	}

	for _, key := range keys {
		err := txn.Delete([]byte(key))
		if err != nil {
			return err
		}
	}

	return nil
}

// EncodeRecordKey encodes the string value to a byte slice.
func EncodeRecordKey(tablePrefix string, h int64, columnID int64) []byte {
	var (
		buf []byte
		err error
	)

	if columnID == 0 { // Ignore columnID.
		buf, err = kv.EncodeValue(tablePrefix, h)
	} else {
		buf, err = kv.EncodeValue(tablePrefix, h, columnID)
	}
	if err != nil {
		log.Fatal("should never happend")
	}
	return buf
}

// DecodeHandleFromRowKey decodes the string form a row key and returns an int64.
func DecodeHandleFromRowKey(rk string) (int64, error) {
	vals, err := kv.DecodeValue([]byte(rk))
	if err != nil {
		return 0, errors.Trace(err)
	}
	return vals[1].(int64), nil
}

// RowKeyPrefixFilter returns a function which checks whether currentKey has decoded rowKeyPrefix as prefix.
func RowKeyPrefixFilter(rowKeyPrefix []byte) kv.FnKeyCmp {
	return func(currentKey []byte) bool {
		// Next until key without prefix of this record.
		raw, err := codec.StripEnd(rowKeyPrefix)
		if err != nil {
			return false
		}
		return !bytes.HasPrefix(currentKey, raw)
	}
}

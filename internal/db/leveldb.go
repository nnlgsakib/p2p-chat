package db

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// LevelDBStore represents a LevelDB key-value store.
type LevelDBStore struct {
	db *leveldb.DB
}

// NewLevelDBStore opens or creates a LevelDB store at the given path.
func NewLevelDBStore(path string) (*LevelDBStore, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open LevelDB: %w", err)
	}
	return &LevelDBStore{db: db}, nil
}

// Put writes a key-value pair to the store.
func (s *LevelDBStore) Put(key, value []byte) error {
	return s.db.Put(key, value, nil)
}

// Get retrieves the value for a given key from the store.
func (s *LevelDBStore) Get(key []byte) ([]byte, error) {
	data, err := s.db.Get(key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get key from LevelDB: %w", err)
	}
	return data, nil
}

// Delete removes a key-value pair from the store.
func (s *LevelDBStore) Delete(key []byte) error {
	return s.db.Delete(key, nil)
}

// NewIterator returns a new iterator over the store.
func (s *LevelDBStore) NewIterator(slice *util.Range) iterator.Iterator {
	return s.db.NewIterator(slice, nil)
}

// NewIteratorWithPrefix returns a new iterator over a key prefix.
func (s *LevelDBStore) NewIteratorWithPrefix(prefix []byte) iterator.Iterator {
	return s.db.NewIterator(util.BytesPrefix(prefix), nil)
}

// Close closes the LevelDB store.
func (s *LevelDBStore) Close() error {
	return s.db.Close()
}



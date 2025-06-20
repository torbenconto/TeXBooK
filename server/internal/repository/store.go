package repository

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

	"github.com/torbenconto/TeXBooK/internal/datasources"
	"go.etcd.io/bbolt"
)

type StoredDataSources map[string]datasources.DataSource

type Store[T any] struct {
	db         *bbolt.DB
	bucketName []byte
	path       string
}

func New[T any](path, bucketName string) (*Store[T], error) {
	_ = os.MkdirAll(filepath.Dir(path), 0755)

	db, err := bbolt.Open(path, 0600, bbolt.DefaultOptions)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bbolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})

	return &Store[T]{
		db:         db,
		bucketName: []byte(bucketName),
		path:       path,
	}, nil
}

func (s *Store[T]) Save(value T) error {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(value); err != nil {
		return fmt.Errorf("error encoding config struct: %v", err)
	}

	s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(s.bucketName)

		return bucket.Put([]byte("config"), buf.Bytes())
	})

	return nil
}

func (s *Store[T]) Get() (T, error) {
	var result T
	var buf bytes.Buffer

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(s.bucketName)
		if bucket == nil {
			return nil
		}

		data := bucket.Get([]byte("config"))
		if data == nil {
			return nil
		}

		_, err := buf.Write(data)
		return err
	})
	if err != nil {
		return result, err
	}

	if buf.Len() == 0 {
		return result, nil
	}

	dec := gob.NewDecoder(&buf)
	if err := dec.Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}

func (s *Store[T]) Close() error {
	return s.db.Close()
}

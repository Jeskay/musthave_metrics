package db

import (
	"bytes"
	"encoding/gob"
	"os"
	"sync"

	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
)

type FileStorage struct {
	filename string
	mu       sync.Mutex
}

func NewFileStorage(filename string) (*FileStorage, error) {
	f, err := os.OpenFile(filename, os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	err = f.Close()
	return &FileStorage{
		filename: filename,
	}, err
}

func (fs *FileStorage) Save(metrics []dto.Metrics) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(metrics)
	if err != nil {
		return err
	}
	fs.mu.Lock()
	err = os.WriteFile(fs.filename, buf.Bytes(), 0644)
	fs.mu.Unlock()
	return err
}

func (fs *FileStorage) Load() (metrics []dto.Metrics, err error) {
	var buf bytes.Buffer
	fs.mu.Lock()
	b, err := os.ReadFile(fs.filename)
	fs.mu.Unlock()
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(b)
	if err != nil {
		return nil, err
	}
	dec := gob.NewDecoder(&buf)
	err = dec.Decode(&metrics)
	return
}

func (fs *FileStorage) Health() bool { return true }

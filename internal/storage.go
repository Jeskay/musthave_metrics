package internal

import (
	"bytes"
	"encoding/gob"
	"os"
	"sync"
)

type MemStorage struct {
	data sync.Map
}

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

func (fs *FileStorage) Save(metrics []Metric) error {
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

func (fs *FileStorage) Load() (metrics []Metric, err error) {
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

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: sync.Map{},
	}
}

func (ms *MemStorage) Set(key string, value MetricValue) {
	ms.data.Store(key, value)
}

func (ms *MemStorage) Get(key string) (MetricValue, bool) {
	if m, ok := ms.data.Load(key); ok {
		return m.(MetricValue), ok
	}
	return MetricValue{}, false
}

func (ms *MemStorage) GetAll() []*Metric {
	m := make([]*Metric, 0)
	ms.data.Range(func(key, value any) bool {
		m = append(m, &Metric{Name: key.(string), Value: value.(MetricValue)})
		return true
	})
	return m
}

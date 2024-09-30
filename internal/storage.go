package internal

type MemStorage struct {
	data map[string]Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: make(map[string]Metric),
	}
}

func (ms *MemStorage) Set(key string, value Metric) {
	ms.data[key] = value
}

func (ms *MemStorage) Get(key string) Metric {
	return ms.data[key]
}

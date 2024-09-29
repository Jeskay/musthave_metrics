package internal

type MemStorage struct {
	data map[string]interface{}
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: make(map[string]interface{}),
	}
}

func (ms *MemStorage) Add(key string, value interface{}) {
	ms.data[key] = value
}

func (ms *MemStorage) Get(key string) interface{} {
	return ms.data[key]
}

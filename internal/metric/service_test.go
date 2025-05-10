package metric

import dto "github.com/Jeskay/musthave_metrics/internal/Dto"

type mockedFileStorage interface {
	Save([]dto.Metrics) error
	Load() (metrics []dto.Metrics, err error)
	Health() bool
}

type spyFileStorage struct{}

type spyRepository struct{}

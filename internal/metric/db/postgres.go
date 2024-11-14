package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStorage struct {
	logger *slog.Logger
	db     *sql.DB
}

func NewPostgresStorage(db *sql.DB, logger slog.Handler) (*PostgresStorage, error) {
	ps := &PostgresStorage{
		db:     db,
		logger: slog.New(logger),
	}
	err := ps.init()
	if err != nil {
		return &PostgresStorage{db: nil}, err
	}
	return ps, nil
}

func (ps *PostgresStorage) init() error {
	_, err := ps.db.Exec(`
		CREATE TABLE IF NOT EXISTS metric (
			name varchar(500) PRIMARY KEY UNIQUE,
			counterValue bigint,
			gaugeValue  double precision
		);
	`)
	return err
}

func (ps *PostgresStorage) Get(key string) (dto.Metrics, bool) {
	var (
		name    string
		gauge   sql.NullFloat64
		counter sql.NullInt64
	)
	row := ps.db.QueryRow(`SELECT * FROM metric WHERE name = $1;`, key)
	if row.Err() != nil {
		slog.Error(row.Err().Error())
		return dto.Metrics{}, false
	}
	if err := row.Scan(&name, &counter, &gauge); err != nil {
		ps.logger.Error(err.Error())
		return dto.Metrics{}, false
	}
	if counter.Valid {
		return dto.NewCounterMetrics(name, counter.Int64), true
	}
	return dto.NewGaugeMetrics(name, gauge.Float64), true
}

func (ps *PostgresStorage) Set(value dto.Metrics) error {
	counter, gauge := value.QueryValues()
	_, err := ps.db.Exec(`
		INSERT INTO metric (name, countervalue, gaugevalue)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE SET 
			gaugevalue = excluded.gaugevalue,
			countervalue = metric.countervalue + excluded.countervalue;
		`, value.ID, counter, gauge)
	if err != nil {
		ps.logger.Error(err.Error())
		return err
	}
	return nil
}

func (ps *PostgresStorage) SetMany(values []dto.Metrics) error {
	var query strings.Builder
	args := make([]any, 0)
	query.WriteString("INSERT INTO metric (name, countervalue, gaugevalue) VALUES")
	for i, v := range values {
		if i != 0 {
			query.WriteString(", ")
		}
		query.WriteString(fmt.Sprintf("($%d::varchar(500), $%d::bigint, $%d::double precision)", 3*i+1, 3*i+2, 3*i+3))
		counter, gauge := v.QueryValues()
		args = append(args, v.ID, counter, gauge)
	}
	query.WriteString(`
		ON CONFLICT (name) DO UPDATE SET 
			gaugevalue = excluded.gaugevalue,
			countervalue = metric.countervalue + excluded.countervalue;
	`)
	qstr := query.String()
	_, err := ps.db.Exec(qstr, args...)
	if err != nil {
		ps.logger.Error(err.Error())
		return err
	}
	return nil
}

func (ps *PostgresStorage) GetMany(keys []string) ([]dto.Metrics, error) {
	var (
		name    string
		gauge   sql.NullFloat64
		counter sql.NullInt64
	)
	m := make([]dto.Metrics, 0)
	args := make([]any, len(keys))
	var query strings.Builder
	query.WriteString("SELECT * FROM metric WHERE name IN (")
	for i, k := range keys {
		if i != 0 {
			query.WriteString(", ")
		}
		query.WriteString(fmt.Sprintf("$%d", i+1))
		args[i] = k
	}
	query.WriteString(");")
	qstr := query.String()
	rows, err := ps.db.Query(qstr, args...)
	if err != nil {
		ps.logger.Error(err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&name, &counter, &gauge); err != nil {
			ps.logger.Error(err.Error())
			return nil, err
		}
		var metric dto.Metrics
		if counter.Valid {
			metric = dto.NewCounterMetrics(name, counter.Int64)
		} else {
			metric = dto.NewGaugeMetrics(name, gauge.Float64)
		}
		m = append(m, metric)
	}
	return m, nil
}

func (ps *PostgresStorage) GetAll() ([]dto.Metrics, error) {
	var (
		name    string
		gauge   sql.NullFloat64
		counter sql.NullInt64
	)
	m := make([]dto.Metrics, 0)
	rows, err := ps.db.Query(`SELECT * FROM metric`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&name, &counter, &gauge); err != nil {
			return nil, err
		}
		var metric dto.Metrics
		if counter.Valid {
			metric = dto.NewCounterMetrics(name, counter.Int64)
		} else {
			metric = dto.NewGaugeMetrics(name, gauge.Float64)
		}
		m = append(m, metric)
	}
	return m, nil
}

func (ps *PostgresStorage) Health() bool {
	return ps.db != nil && ps.db.Ping() == nil
}

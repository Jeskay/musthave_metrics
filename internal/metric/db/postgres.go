package db

import (
	"database/sql"
	"fmt"

	"github.com/Jeskay/musthave_metrics/internal"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(conn string) *PostgresStorage {
	db, err := sql.Open("pgx", conn)
	if err != nil {
		return &PostgresStorage{db: nil}
	}
	ps := &PostgresStorage{db}
	err = ps.init()
	if err != nil {
		return &PostgresStorage{db: nil}
	}
	return ps
}

func (ps *PostgresStorage) init() error {
	_, err := ps.db.Exec(`
		CREATE TABLE IF NOT EXISTS metric (
			name varchar(500) PRIMARY KEY UNIQUE,
			counterValue double precision,
			gaugeValue integer
		);
	`)
	return err
}

func (ps *PostgresStorage) Get(key string) (internal.MetricValue, bool) {
	var (
		name    string
		gauge   sql.NullFloat64
		counter sql.NullInt64
	)
	row := ps.db.QueryRow(`SELECT * FROM metric WHERE name = $1;`, key)
	if row.Err() != nil {
		return internal.MetricValue{}, false
	}
	if err := row.Scan(&name, &counter, &gauge); err != nil {
		return internal.MetricValue{}, false
	}
	if counter.Valid {
		return internal.MetricValue{Value: counter.Int64, Type: internal.CounterMetric}, true
	}
	return internal.MetricValue{Value: gauge.Float64, Type: internal.GaugeMetric}, true
}

func (ps *PostgresStorage) Set(key string, value internal.MetricValue) {
	if value.Type == internal.CounterMetric {
		res, err := ps.db.Exec(`
			INSERT INTO metric (name, countervalue)
			VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE
				SET countervalue = excluded.countervalue;
		`, key, value.Value)
		if err != nil {
			fmt.Println(res)
			panic(err)
		}
	} else {
		_, err := ps.db.Exec(`
			INSERT INTO metric (name, gaugevalue)
			VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE
				SET gaugevalue = excluded.gaugevalue;
		`, key, value.Value)
		if err != nil {
			panic(err)
		}
	}
}

func (ps *PostgresStorage) GetAll() []*internal.Metric {
	var (
		name    string
		gauge   sql.NullFloat64
		counter sql.NullInt64
	)
	m := make([]*internal.Metric, 0)
	rows, err := ps.db.Query(`SELECT * FROM metric`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&name, &counter, &gauge); err != nil {
			panic(err)
		}
		metric := &internal.Metric{Name: name}
		if counter.Valid {
			metric.Value = internal.MetricValue{Type: internal.CounterMetric, Value: counter.Int64}
		} else {
			metric.Value = internal.MetricValue{Type: internal.GaugeMetric, Value: gauge.Float64}
		}
		m = append(m, metric)
	}
	return m
}

func (ps *PostgresStorage) Health() bool {
	return ps.db != nil && ps.db.Ping() == nil
}

package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	dto "github.com/Jeskay/musthave_metrics/internal/Dto"
	"github.com/Jeskay/musthave_metrics/internal/util"
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
	query := `
		CREATE TABLE IF NOT EXISTS metric (
			name varchar(500) PRIMARY KEY UNIQUE,
			counterValue bigint,
			gaugeValue  double precision
		);
	`
	err := util.TryRun(func() (err error) {
		_, err = ps.db.Exec(query)
		return
	}, util.IsPGConnectionError)

	return err
}

func (ps *PostgresStorage) Get(key string) (dto.Metrics, bool) {
	var (
		name    string
		gauge   sql.NullFloat64
		counter sql.NullInt64
	)
	var row *sql.Row
	err := util.TryRun(func() (err error) {
		row = ps.db.QueryRow(`SELECT * FROM metric WHERE name = $1;`, key)
		return row.Err()
	}, util.IsPGConnectionError)

	if err != nil {
		slog.Error(err.Error())
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
	query := `
		INSERT INTO metric (name, countervalue, gaugevalue)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE SET 
			gaugevalue = excluded.gaugevalue,
			countervalue = metric.countervalue + excluded.countervalue;`

	err := util.TryRun(func() (err error) {
		_, err = ps.db.Exec(query, value.ID, counter, gauge)
		return
	}, util.IsPGConnectionError)

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
	err := util.TryRun(func() (err error) {
		_, err = ps.db.Exec(qstr, args...)
		return
	}, util.IsPGConnectionError)

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
	var rows *sql.Rows
	err := util.TryRun(func() (err error) {
		rows, err = ps.db.Query(qstr, args...)
		return
	}, util.IsPGConnectionError)

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
	var rows *sql.Rows
	err := util.TryRun(func() (err error) {
		rows, err = ps.db.Query(`SELECT * FROM metric`)
		return
	}, util.IsPGConnectionError)

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
	err := util.TryRun(func() (err error) {
		err = ps.db.Ping()
		return
	}, util.IsPGConnectionError)
	return ps.db != nil && err == nil
}

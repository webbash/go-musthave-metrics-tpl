package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	pgerrors "github.com/webbash/go-musthave-metrics-tpl.git/internal/errors"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/retry"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	var rows *sql.Rows
	err := r.withRetry(ctx, func() error {
		var err error
		rows, err = r.db.QueryContext(ctx, "SELECT id, value FROM metrics WHERE type = $1", models.Gauge)
		if err != nil {
			return fmt.Errorf("r.GetAllGauges: %w", err)
		}
		defer rows.Close()

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("r.GetAllGauges: %w", err)
	}

	gauges := make(map[string]float64)

	for rows.Next() {
		var (
			id    string
			value float64
		)

		if err := rows.Scan(&id, &value); err != nil {
			return nil, fmt.Errorf("r.GetAllGauges: failed to scan gauge metric: %w", err)
		}

		gauges[id] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("r.GetAllGauges: failed to iterate over rows: %w", err)
	}

	return gauges, nil
}

func (r *PostgresRepository) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	var rows *sql.Rows
	err := r.withRetry(ctx, func() error {
		var err error
		rows, err = r.db.QueryContext(ctx, "SELECT id, delta FROM metrics WHERE type = $1", models.Counter)
		if err != nil {
			return err
		}
		defer rows.Close()

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("r.GetAllCounters: %w", err)
	}
	counter := make(map[string]int64)

	for rows.Next() {
		var (
			id    string
			value int64
		)

		if err := rows.Scan(&id, &value); err != nil {
			return nil, fmt.Errorf("r.GetAllCounters: failed to scan counter metric: %w", err)
		}

		counter[id] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("r.GetAllCounters: failed to iterate over rows: %w", err)
	}

	return counter, nil
}
func (r *PostgresRepository) GetCounter(ctx context.Context, metricName string) (int64, error) {
	var value int64
	err := r.withRetry(ctx, func() error {
		row := r.db.QueryRowContext(ctx, "SELECT delta FROM metrics WHERE id = $1", metricName)
		return row.Scan(&value)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("r.GetCounter: metric not found: %w", err)
	} else if err != nil {
		return 0, fmt.Errorf("r.GetCounter: failed to scan counter metric: %w", err)
	}

	return value, nil
}
func (r *PostgresRepository) GetGauge(ctx context.Context, metricName string) (float64, error) {
	var value float64
	err := r.withRetry(ctx, func() error {
		row := r.db.QueryRowContext(ctx, "SELECT value FROM metrics WHERE id = $1", metricName)
		return row.Scan(&value)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("r.GetGauge: metric not found: %w", err)
	} else if err != nil {
		return 0, fmt.Errorf("r.GetGauge: failed to get value from scan: %w", err)
	}

	return value, nil
}
func (r *PostgresRepository) IncrementCounter(ctx context.Context, metricName string, value int64) error {
	var result sql.Result
	err := r.withRetry(ctx, func() error {
		var err error
		result, err = r.db.ExecContext(ctx, "INSERT INTO metrics(id, type, delta) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET delta = metrics.delta + $3", metricName, models.Counter, value)
		if err != nil {
			return fmt.Errorf("failed to increment counter: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("r.IncrementCounter (after all retries): %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("r.IncrementCounter - failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("r.IncrementCounter - no rows affected")
	}

	return nil
}
func (r *PostgresRepository) UpdateGauge(ctx context.Context, metricName string, value float64) error {
	var result sql.Result
	err := r.withRetry(ctx, func() error {
		var err error
		result, err = r.db.ExecContext(ctx, "INSERT INTO metrics(id, type, value) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET value = $3", metricName, models.Gauge, value)
		return err
	})
	if err != nil {
		return fmt.Errorf("r.UpdateGauge: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("r.UpdateGauge: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("r.UpdateGauge: no rows affected")
	}

	return nil
}
func (r *PostgresRepository) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	var rows *sql.Rows
	err := r.withRetry(ctx, func() error {
		var err error
		rows, err = r.db.QueryContext(ctx, "SELECT id, value FROM metrics")
		if err != nil {
			return err
		}
		defer rows.Close()

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("r.GetAllMetrics: %w", err)
	}

	metrics := make([]models.Metrics, 0)

	for rows.Next() {
		var metric models.Metrics

		if err := rows.Scan(&metric); err != nil {
			return nil, fmt.Errorf("r.GetAllMetrics: failed to scan counter metric: %w", err)
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (r *PostgresRepository) UpdateMany(ctx context.Context, metrics []models.Metrics) error {
	err := r.withRetry(ctx, func() error {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		stmtCounter, err := tx.PrepareContext(ctx, `
INSERT INTO metrics (id, type, delta) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET delta = metrics.delta + EXCLUDED.delta;
`)
		if err != nil {
			return fmt.Errorf("failed to prepare statement for counter insert: %w", err)
		}
		defer stmtCounter.Close()

		stmtGauge, err := tx.PrepareContext(ctx, `
INSERT INTO metrics (id, type, value) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value;
`)
		if err != nil {
			return fmt.Errorf("failed to prepare statement for gauge insert: %w", err)
		}
		defer stmtGauge.Close()

		for _, metric := range metrics {
			switch metric.MType {
			case models.Counter:
				_, err := stmtCounter.ExecContext(ctx, metric.ID, metric.MType, *metric.Delta)
				if err != nil {
					return fmt.Errorf("failed to insert metric (counter): %w", err)
				}
			case models.Gauge:
				_, err := stmtGauge.ExecContext(ctx, metric.ID, metric.MType, *metric.Value)
				if err != nil {
					return fmt.Errorf("failed to insert metric (gauge): %w", err)
				}
			default:
				return fmt.Errorf("unknown metric type: %s", metric.MType)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("r.UpdateBatch: %w", err)
	}

	return nil
}

func (r *PostgresRepository) withRetry(ctx context.Context, fn func() error) error {
	classifier := pgerrors.NewPostgresErrorClassifier()

	return retry.Do(
		ctx,
		fn,
		func(err error) bool {
			return classifier.Classify(err) == pgerrors.Retriable
		},
		[]time.Duration{
			1 * time.Second,
			3 * time.Second,
			5 * time.Second,
		},
	)
}

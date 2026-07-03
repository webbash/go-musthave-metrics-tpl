package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
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
	rows, err := r.db.QueryContext(ctx, "SELECT id, value FROM metrics WHERE type = $1", models.Gauge)
	if err != nil {
		return nil, fmt.Errorf("failed to query gauge metrics: %w", err)
	}
	defer rows.Close()
	gauges := make(map[string]float64)

	for rows.Next() {
		var (
			id    string
			value float64
		)

		if err := rows.Scan(&id, &value); err != nil {
			return nil, fmt.Errorf("failed to scan gauge metric: %w", err)
		}

		gauges[id] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return gauges, nil
}

func (r *PostgresRepository) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, delta FROM metrics WHERE type = $1", models.Counter)
	if err != nil {
		return nil, fmt.Errorf("failed to query counter metrics: %w", err)
	}
	defer rows.Close()
	counter := make(map[string]int64)

	for rows.Next() {
		var (
			id    string
			value int64
		)

		if err := rows.Scan(&id, &value); err != nil {
			return nil, fmt.Errorf("failed to scan counter metric: %w", err)
		}

		counter[id] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return counter, nil
}
func (r *PostgresRepository) GetCounter(ctx context.Context, metricName string) (int64, error) {
	row := r.db.QueryRowContext(ctx, "SELECT delta FROM metrics WHERE id = $1", metricName)
	var value int64
	err := row.Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("metric not found")
	} else if err != nil {
		return 0, fmt.Errorf("failed to scan counter metric: %w", err)
	}

	return value, nil
}
func (r *PostgresRepository) GetGauge(ctx context.Context, metricName string) (float64, error) {
	row := r.db.QueryRowContext(ctx, "SELECT value FROM metrics WHERE id = $1", metricName)
	var value float64
	err := row.Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("metric not found")
	} else if err != nil {
		return 0, fmt.Errorf("failed to scan gauge metric: %w", err)
	}

	return value, nil
}
func (r *PostgresRepository) IncrementCounter(ctx context.Context, metricName string, value int64) error {
	result, err := r.db.ExecContext(ctx, "INSERT INTO metrics(id, type, delta) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET delta = metrics.delta + $3", metricName, models.Counter, value)
	if err != nil {
		return fmt.Errorf("failed to increment counter: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}
func (r *PostgresRepository) UpdateGauge(ctx context.Context, metricName string, value float64) error {
	result, err := r.db.ExecContext(ctx, "INSERT INTO metrics(id, type, value) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET value = $3", metricName, models.Gauge, value)
	if err != nil {
		return fmt.Errorf("failed to update gauge: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}
func (r *PostgresRepository) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM metrics")
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()
	metrics := make([]models.Metrics, 0)
	err = rows.Scan(&metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to scan metrics: %w", err)
	}

	return metrics, nil
}

func (r *PostgresRepository) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
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
			_, err := stmtCounter.ExecContext(ctx, metric.ID, metric.MType, metric.Delta)
			if err != nil {
				return fmt.Errorf("failed to insert metric (counter): %w", err)
			}
		case models.Gauge:
			_, err := stmtGauge.ExecContext(ctx, metric.ID, metric.MType, metric.Value)
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
}

//
//func (r *PostgresRepository) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
//	builder := squirrel.
//		Insert("metrics").
//		Columns("id", "type", "delta", "value")
//
//	for _, metric := range metrics {
//		builder = builder.Values(metric.ID, metric.MType, metric.Delta, metric.Value)
//	}
//
//	sqlBatchInsert, args, err := builder.
//		Suffix("ON CONFLICT (id) DO UPDATE SET delta = EXCLUDED.delta, value = EXCLUDED.value").
//		PlaceholderFormat(squirrel.Dollar).
//		ToSql()
//
//	if err != nil {
//		return fmt.Errorf("failed to build update query: %w", err)
//	}
//
//	result, err := r.db.ExecContext(ctx, sqlBatchInsert, args...)
//	if err != nil {
//		return fmt.Errorf("failed to update metrics: %w", err)
//	}
//
//	rowsAffected, err := result.RowsAffected()
//	if err != nil {
//		return fmt.Errorf("failed to get rows affected: %w", err)
//	}
//
//	if rowsAffected == 0 {
//		return fmt.Errorf("no rows affected")
//	}
//
//	return nil
//}

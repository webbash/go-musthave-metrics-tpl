-- +goose Up
CREATE TABLE metrics
(
    id    VARCHAR(200) PRIMARY KEY,
    type  TEXT NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION
);
-- goose -dir migrations postgres \
-- "postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable" up

-- +goose Down
DROP TABLE IF EXISTS metrics;

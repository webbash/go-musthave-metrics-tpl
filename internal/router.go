package internal

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/config"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value_list"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value_metric"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/ping_db"
	update_handler "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/update"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/update_batch"
	update_metric_handler "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/update_metric"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/middleware"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	"go.uber.org/zap"
)

type Router struct {
	cfg            config.Config
	logger         *zap.SugaredLogger
	router         *chi.Mux
	metricsService *service.MetricsService
	repository     service.MetricsRepository
	db             *sql.DB
}

func NewRouter(cfg config.Config, logger *zap.SugaredLogger, metricsService *service.MetricsService, repository service.MetricsRepository, db *sql.DB) *Router {
	return &Router{
		cfg:            cfg,
		logger:         logger,
		router:         chi.NewRouter(),
		metricsService: metricsService,
		repository:     repository,
		db:             db,
	}
}

func (r *Router) Init() *chi.Mux {
	pingH := ping_db.NewHandler(r.db, r.logger)
	updateH := update_handler.NewHandler(r.metricsService)
	updateMetricH := update_metric_handler.NewHandler(r.metricsService)
	updateBatchH := update_batch.NewHandler(r.metricsService, r.logger)
	getValueMetricH := get_value_metric.NewHandler(r.metricsService)
	getValueH := get_value.NewHandler(r.metricsService)
	getValueListH := get_value_list.NewHandler(r.repository)

	r.router.Use(middleware.LoggingMiddleware(r.logger))
	r.router.Use(middleware.GzipMiddleware())

	r.router.Get("/", getValueListH.ServeHTTP)
	r.router.Post("/value", getValueMetricH.ServeHTTP)
	r.router.Post("/value/", getValueMetricH.ServeHTTP)
	r.router.Post("/updates", updateBatchH.ServeHTTP)
	r.router.Post("/updates/", updateBatchH.ServeHTTP)
	r.router.Post("/update", updateMetricH.ServeHTTP)
	r.router.Post("/update/", updateMetricH.ServeHTTP)
	r.router.Get("/value/{metricType}/{metricName}", getValueH.ServeHTTP)
	r.router.Post("/update/{metricType}/{metricName}/{metricValue}", updateH.ServeHTTP)
	r.router.Get("/ping", pingH.ServeHTTP)

	return r.router
}

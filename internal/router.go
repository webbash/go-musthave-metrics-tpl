// Package internal assembles the application's HTTP routes.
package internal

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/audit"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/config"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
	getvalue "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value"
	getvaluelist "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value_list"
	getvaluemetric "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value_metric"
	pingdb "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/ping_db"
	update_handler "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/update"
	updatebatch "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/update_batch"
	updatemetric "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/update_metric"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/middleware"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
)

// generate:reset
type Router struct {
	cfg            config.Config
	logger         *zap.SugaredLogger
	router         *chi.Mux
	metricsService *service.MetricsService
	repository     service.MetricsRepository
	db             *sql.DB
	subject        *audit.Subject
	decryptor      *crypto.Decryptor
}

func NewRouter(cfg config.Config, logger *zap.SugaredLogger, metricsService *service.MetricsService, repository service.MetricsRepository, db *sql.DB, subject *audit.Subject, decryptor *crypto.Decryptor) *Router {
	return &Router{
		cfg:            cfg,
		logger:         logger,
		router:         chi.NewRouter(),
		metricsService: metricsService,
		repository:     repository,
		db:             db,
		subject:        subject,
		decryptor:      decryptor,
	}
}

func (r *Router) Init() *chi.Mux {
	pingH := pingdb.NewHandler(r.db, r.logger)
	updateH := update_handler.NewHandler(r.metricsService, r.subject, r.logger)
	updateMetricH := updatemetric.NewHandler(r.metricsService, r.subject, r.logger)
	updateBatchH := updatebatch.NewHandler(r.metricsService, r.logger, r.subject)
	getValueMetricH := getvaluemetric.NewHandler(r.metricsService)
	getValueH := getvalue.NewHandler(r.metricsService)
	getValueListH := getvaluelist.NewHandler(r.repository)

	r.router.Use(middleware.LoggingMiddleware(r.logger))
	if r.decryptor != nil {
		r.router.Use(middleware.DecryptMiddleware(r.decryptor))
	}
	r.router.Use(middleware.GzipMiddleware())
	if r.cfg.HashSecret != "" {
		signer := crypto.NewSHA256Signer(r.cfg.HashSecret)
		r.router.Use(middleware.HashCheckMiddleware(signer, r.logger))
	}

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

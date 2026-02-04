package main

import (
	"avito-courier/internal/config"
	"avito-courier/internal/gateway/order"
	"avito-courier/internal/handler"
	"avito-courier/internal/middleware"
	"avito-courier/internal/repository"
	"avito-courier/internal/router"
	"avito-courier/internal/transport/kafka"
	"avito-courier/internal/usecase"
	pkgdb "avito-courier/pkg/db"
	_ "net/http/pprof"

	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	log.Printf("Starting Courier Service")
	log.Printf("Port: %s | Metrics: %v", cfg.Port, cfg.Metrics.Enabled)
	log.Printf("Kafka: %s (topic: %s)", cfg.Kafka.Brokers, cfg.Kafka.OrderTopic)
	log.Printf("Rate Limit: %v (%.1f RPS)", cfg.RateLimit.Enabled, cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)

	if cfg.Pprof.Enabled {
		go func() {
			addr := "localhost:" + cfg.Pprof.Port
			log.Printf("PPROF server: http://%s%s", addr, cfg.Pprof.Endpoint)
			if err := http.ListenAndServe(addr, nil); err != nil {
				log.Printf("PPROF error: %v", err)
			}
		}()
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	pool := initDatabase(ctx, cfg)
	defer pool.Close()
	log.Println("Database connected")

	courierRepo := repository.NewCourierRepository(pool)
	deliveryRepo := repository.NewDeliveryRepository(pool)

	deliveryFactory := usecase.NewDeliveryTimeFactory()

	deliveryUC := usecase.NewDeliveryUsecase(pool, courierRepo, deliveryRepo, deliveryFactory)
	courierUC := usecase.NewCourierUsecase(courierRepo, deliveryUC)

	eventDeliveryUC := usecase.NewEventDeliveryUsecase(deliveryUC)
	eventFactory := usecase.NewEventHandlerFactory(eventDeliveryUC)
	log.Println("Event handler factory initialized")

	orderGateway := order.NewHTTPOrderGateway(cfg)
	log.Println("Order gateway initialized")

	courierHandler := handler.NewCourierHandler(courierUC)
	deliveryHandler := handler.NewDeliveryHandler(deliveryUC)

	var rateLimiter *middleware.RateLimiter
	if cfg.RateLimit.Enabled {
		rateLimiter = middleware.NewRateLimiter(middleware.RateLimiterConfig{
			RequestsPerSecond: cfg.RateLimit.RequestsPerSecond,
			Burst:             cfg.RateLimit.Burst,
			Enabled:           true,
		})
		log.Println("Rate limiter initialized")
	}

	log.Printf("PPROF: %v (port: %s)", cfg.Pprof.Enabled, cfg.Pprof.Port)

	if cfg.Pprof.Enabled {
		startPprofServer(cfg.Pprof.Port)
	}

	mux := router.NewRouter(courierHandler, deliveryHandler, rateLimiter)

	var httpHandler http.Handler = mux
	if cfg.Metrics.Enabled {
		registerMetrics()
		log.Printf("Prometheus metrics enabled at %s", cfg.Metrics.Path)
	}

	if cfg.ServiceOrderURL != "" {
		poller := usecase.NewOrderPoller(orderGateway, deliveryUC)
		go poller.Start(ctx)
		log.Println("Order poller started (5s ticker)")
	}

	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.OrderTopic != "" {
		consumer := kafka.NewConsumer(eventFactory, orderGateway)
		go consumer.StartConsumerGroup(ctx,
			cfg.Kafka.Brokers,
			cfg.Kafka.ConsumerGroup,
			[]string{cfg.Kafka.OrderTopic})
		log.Printf("Kafka consumer started: topic=%s", cfg.Kafka.OrderTopic)
	}

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      httpHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(100 * time.Millisecond)

		log.Printf("HTTP server listening on http://localhost:%s", cfg.Port)
		log.Println("Available endpoints:")
		log.Println("GET    /api/couriers              - List all couriers")
		log.Println("GET    /api/couriers/{id}         - Get courier by ID")
		log.Println("POST   /api/couriers              - Create new courier")
		log.Println("PUT    /api/couriers/{id}         - Update courier")
		log.Println("POST   /api/delivery/assign       - Assign courier to order")
		log.Println("POST   /api/delivery/unassign     - Unassign courier from order")
		log.Println("GET    /health                    - Health check")
		if cfg.Metrics.Enabled {
			log.Printf("  GET    %s                    - Prometheus metrics", cfg.Metrics.Path)
		}

		go func() {
			time.Sleep(1 * time.Second)
			resp, err := http.Get(fmt.Sprintf("http://localhost:%s/health", cfg.Port))
			if err != nil {
				log.Printf("Self-test failed: %v", err)
			} else {
				resp.Body.Close()
				log.Printf("Self-test passed: health check returned %d", resp.StatusCode)
			}
		}()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Println("Service started successfully. Press Ctrl+C to stop.")
	<-ctx.Done()
	log.Println("Shutdown signal received, starting graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Stopping HTTP server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Waiting for goroutines to finish...")
	wg.Wait()

	log.Println("Service stopped gracefully")
}

func initDatabase(ctx context.Context, cfg *config.Config) *pgxpool.Pool {
	dbCfg := &pkgdb.DBConfig{
		Host:            cfg.DB.Host,
		Port:            cfg.DB.Port,
		User:            cfg.DB.User,
		Password:        cfg.DB.Password,
		DBName:          cfg.DB.Name,
		MaxConns:        20,
		MinConns:        5,
		MaxConnLifetime: time.Hour,
	}

	pool, err := pkgdb.InitPool(ctx, dbCfg, 10)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	var result int
	err = pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		log.Fatalf("Database test query failed: %v", err)
	}

	log.Println("Connected to PostgreSQL")
	return pool
}
func startPprofServer(port string) {
	go func() {
		addr := "localhost:" + port
		log.Printf("PPROF server starting on http://%s/debug/pprof", addr)

		server := &http.Server{
			Addr:         addr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      http.DefaultServeMux,
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("PPROF server error: %v", err)
		}
	}()
}

func registerMetrics() {
	prometheus.Register(prometheus.NewGoCollector())
	prometheus.Register(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	prometheus.MustRegister(
		prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kafka_events_processed_total",
				Help: "Total number of Kafka events processed",
			},
			[]string{"status", "result"},
		),
	)
}

func startBackgroundJobs(ctx context.Context, deliveryUC *usecase.DeliveryUsecase) {
	intervalSec := 10
	if v := os.Getenv("DELIVERY_RELEASE_INTERVAL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			intervalSec = n
		}
	}

	go deliveryUC.AutoRelease(ctx, time.Duration(intervalSec)*time.Second)
	log.Printf("Auto-release job started (interval: %ds)", intervalSec)
}

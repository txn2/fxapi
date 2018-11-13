package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/drhodes/golorem"

	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/nu7hatch/gouuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_model/go"
	"go.uber.org/zap"
)

var Version = "0.0.0"

func main() {
	// Counter for calls to this service
	calls := promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_api_calls",
		Help: "Total number api calls.",
	})

	// UUID for this process
	callUuidV4, _ := uuid.NewV4()

	portEnv, ok := os.LookupEnv("PORT")
	if ok != true {
		portEnv = "8080"
	}

	debugEnv, ok := os.LookupEnv("DEBUG")
	if ok != true {
		debugEnv = "true"
	}

	port := flag.String("port", portEnv, "The port to listen on for HTTP requests.")
	debug := flag.String("debug", debugEnv, "debug mode true or false.")
	flag.Parse()

	if *debug == "false" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	logger, _ := zap.NewProduction()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))

	// increment on every call
	r.Use(func(c *gin.Context) {
		calls.Inc()
	})

	// Basic JSON
	r.GET("/", func(c *gin.Context) {
		instanceUuidV4, _ := uuid.NewV4()

		mmf, err := getMetrics()
		if err != nil {
			logger.Error("Error gather metrics", zap.Error(err))

		}

		r.Routes()

		c.JSON(200, gin.H{
			"message":       "ok",
			"time":          time.Now(),
			"calls":         mmf["total_api_calls"].Metric[0].Counter.Value,
			"uuid_call":     instanceUuidV4.String(),
			"uuid_instance": callUuidV4.String(),
			"version":       Version,
		})
	})

	// Epoch
	r.GET("/curve/:high/:std/:dec", func(c *gin.Context) {

		dec := c.Param("dec")

		highInt, err := strconv.Atoi(c.Param("high"))
		if err != nil {
			c.String(500, "high can not be converted to a string")
			logger.Error("high int conversion error", zap.Error(err))
		}

		std, err := strconv.Atoi(c.Param("std"))
		if err != nil {
			c.String(500, "std can not be converted to a string")
			logger.Error("std int conversion error", zap.Error(err))
		}

		num := (float64(highInt) * minuteCurve()) + (rand.Float64() * float64(std))

		// randomly subtract
		if rand.Intn(10) > 5 {
			num = (float64(highInt) * minuteCurve()) - (rand.Float64() * float64(std))
		}

		c.String(200, fmt.Sprintf("%."+dec+"f", num))
	})

	// Epoch
	r.GET("/epoch", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("%d", time.Now().Unix()))
	})

	// Epoch
	r.GET("/second", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("%d", time.Now().Second()))
	})

	// Random Text
	r.GET("/lorem", func(c *gin.Context) {
		c.String(200, lorem.Sentence(5, 20))
	})

	// Prometheus Metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	err := r.Run(":" + *port)
	if err != nil {
		logger.Error("Error starting fxApi server", zap.Error(err))
	}
}

type MappedMetricFamily map[string]*io_prometheus_client.MetricFamily

func minuteCurve() float64 {
	sec := float64(time.Now().Second() + 1)
	if sec > 30 {
		sec = 61 - sec
	}

	return sec / 30
}

func getMetrics() (MappedMetricFamily, error) {
	mmf := make(MappedMetricFamily, 0)
	mf, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return mmf, err
	}

	for _, metric := range mf {
		mmf[*metric.Name] = metric
	}

	return mmf, nil
}

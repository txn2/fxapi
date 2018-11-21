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
		Name: "fxapi_total_api_calls",
		Help: "Total number api calls.",
	})

	counter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fxapi_counter_api",
			Help: "API counter.",
		},
		[]string{"name"},
	)

	incCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fxapi_inc_api",
			Help: "API incrementer.",
		},
		[]string{"name"},
	)

	incSum := promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "fxapi_inc_api_sum",
			Help: "incrementer API summaries.",
		},
		[]string{"name"},
	)

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
	version := flag.Bool("version", false, "display version and exit.")
	flag.Parse()

	if *version == true {
		fmt.Printf("fxapi Version %s", Version)
		os.Exit(0)
	}

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
			c.JSON(500, "can not gather metrics")
			return
		}

		r.Routes()

		c.JSON(200, gin.H{
			"message":       "ok",
			"time":          time.Now(),
			"calls":         mmf["fxapi_total_api_calls"].Metric[0].Counter.Value,
			"uuid_call":     instanceUuidV4.String(),
			"uuid_instance": callUuidV4.String(),
			"version":       Version,
		})
	})

	// Incrementer returns a random number a counter for a name
	// was incremented by
	r.GET("/inc/update/:name/:min/:max", func(c *gin.Context) {
		name := c.Param("name")

		minInt, err := strconv.Atoi(c.Param("min"))
		if err != nil {
			logger.Error("min int conversion error", zap.Error(err))
			c.String(500, "min can not be converted to a number")
			return
		}

		maxInt, err := strconv.Atoi(c.Param("max"))
		if err != nil {
			logger.Error("max int conversion error", zap.Error(err))
			c.String(500, "max can not be converted to a number")
			return
		}

		rInt := rand.Intn(maxInt-minInt) + minInt

		incCounter.With(prometheus.Labels{"name": name}).Add(float64(rInt))
		incSum.With(prometheus.Labels{"name": name}).Observe(float64(rInt))

		c.String(200, fmt.Sprintf("%d", rInt))
	})

	// Inc Count
	r.GET("/inc/count/:name", func(c *gin.Context) {
		name := c.Param("name")

		mmf, err := getMetrics()
		if err != nil {
			logger.Error("Error gathering metrics", zap.Error(err))
			c.String(500, "can not gather metrics")
			return
		}

		// find the metric we just updated
		for _, m := range mmf["fxapi_inc_api"].Metric {
			for _, l := range m.Label {
				if *l.Name == "name" && *l.Value == name {
					c.String(200, fmt.Sprintf("%.0f", *m.Counter.Value))
					return
				}
			}

		}

		logger.Error("Error finding metric", zap.Error(err))
		c.String(500, "can not fine metric")
	})

	// Inc Summary
	r.GET("/inc/summary/:name", func(c *gin.Context) {
		name := c.Param("name")

		mmf, err := getMetrics()
		if err != nil {
			logger.Error("Error gathering metrics", zap.Error(err))
			c.String(500, "can not gather metrics")
			return
		}

		// find the metric we just updated
		for _, m := range mmf["fxapi_inc_api_sum"].Metric {
			for _, l := range m.Label {
				if *l.Name == "name" && *l.Value == name {
					for _, q := range m.Summary.Quantile {
						if *q.Quantile == float64(.5) {
							c.String(200, fmt.Sprintf("%.0f", *q.Value))
							return
						}
					}
				}
			}

		}

		logger.Error("Error finding metric", zap.Error(err))
		c.String(500, "can not fine metric")
	})

	// Counter
	r.GET("/counter/:name/:add", func(c *gin.Context) {
		name := c.Param("name")

		addInt, err := strconv.Atoi(c.Param("add"))
		if err != nil {
			c.String(500, "add can not be converted to an int")
			logger.Error("add int conversion error", zap.Error(err))
			return
		}

		counter.With(prometheus.Labels{"name": name}).Add(float64(addInt))

		mmf, err := getMetrics()
		if err != nil {
			logger.Error("Error gathering metrics", zap.Error(err))
			c.String(500, "can not gather metrics")
			return
		}

		// find the metric we just updated
		for _, m := range mmf["fxapi_counter_api"].Metric {
			for _, l := range m.Label {
				if *l.Name == "name" && *l.Value == name {
					c.String(200, fmt.Sprintf("%.0f", *m.Counter.Value))
					return
				}
			}

		}

		logger.Error("Error finding metric", zap.Error(err))
		c.String(500, "can not fine metric")
	})

	// Curve over a minute
	r.GET("/curve/:high/:std/:dec", func(c *gin.Context) {

		dec := c.Param("dec")

		highInt, err := strconv.Atoi(c.Param("high"))
		if err != nil {
			logger.Error("high int conversion error", zap.Error(err))
			c.String(500, "high can not be converted to a number")
			return
		}

		std, err := strconv.Atoi(c.Param("std"))
		if err != nil {
			logger.Error("std int conversion error", zap.Error(err))
			c.String(500, "std can not be converted to a number")
			return
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

	// Second of the minute
	r.GET("/second", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("%d", time.Now().Second()))
	})

	// Random Text
	r.GET("/lorem", func(c *gin.Context) {
		c.String(200, lorem.Sentence(5, 20))
	})

	// Fixed Number
	r.GET("/fixed-number/:num", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("%s", c.Param("num")))
	})

	// Fxied metric
	r.GET("/metric/:data", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("%s", c.Param("data")))
	})

	// Random Int
	r.GET("/random-int/:from/:to", func(c *gin.Context) {

		fInt, err := strconv.Atoi(c.Param("from"))
		if err != nil {
			logger.Error("from int conversion error", zap.Error(err))
			c.String(500, "from can not be converted to a number")
			return
		}

		tInt, err := strconv.Atoi(c.Param("to"))
		if err != nil {
			logger.Error("to int conversion error", zap.Error(err))
			c.String(500, "to can not be converted to a number")
			return
		}

		rInt := rand.Intn(tInt-fInt) + fInt

		c.String(200, fmt.Sprintf("%d", rInt))
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

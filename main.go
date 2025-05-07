package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	c "github.com/TOMMy-Net/go-balancer/config"
	"github.com/TOMMy-Net/go-balancer/internal/handlers/balancer"
	"github.com/TOMMy-Net/go-balancer/internal/repository/database"
	"github.com/TOMMy-Net/go-balancer/internal/service/ratelimiter"
	"github.com/sirupsen/logrus"
)

func main() {
	loadBalancerServer()
}

// The server for the load balancer operation
func loadBalancerServer() {
	var backendAddrs []string
	logger := logger("load_balancer.log")

	config, err := c.ReadConfig("config.yaml")
	if err != nil {
		logger.Fatal(err)
	}

	db, err := database.NewPostgres(&database.ConfigDB{
		Host:     config.Database.Host,
		Port:     config.Database.Port,
		User:     config.Database.User,
		Password: config.Database.Password,
		NameDB:   config.Database.NameDB,
		SSL:      config.Database.SSL,
	})
	if err != nil {
		logger.Fatal(err)
	}
	backendAddrs = append(backendAddrs, config.Backends.Endpoints...)

	var rateLimiter = ratelimiter.NewRateLimiter(ratelimiter.LimiterConfig{
		DefaultInterval:   time.Duration(config.RateLimiter.DefaultInterval),
		DefaultCapacity:   config.RateLimiter.DefaultCapacity,
		DefaultRefillRate: config.RateLimiter.DefaultRefillRate,
	})

	lb, err := balancer.NewLoadBalancerHandler(logger, backendAddrs, rateLimiter, time.Duration(config.Backends.HealthInterval)*time.Second)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"time":  time.Now().Format(time.RFC3339),
		}).Fatal(err.Error())
	}
	lb.DB = db

	lbAPI := balancer.NewApiMux(lb)

	serverAPI := &http.Server{
		Addr:    ":" + config.APIport,
		Handler: lbAPI,
	}

	server := &http.Server{
		Addr:    ":" + config.LoadBalancerPort,
		Handler: lb,
	}
	// --------- GRACEFULL SHUTDOWN
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Fatalln(err.Error())
		}
		wg.Done()
	}()
	go func() {
		if err := serverAPI.ListenAndServe(); err != nil {
			logger.Fatalln(err.Error())
		}
		wg.Done()
	}()
	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
		serverAPI.Shutdown(context.Background())
	}()
	fmt.Printf("LoadBalancer started on %s port | API started on %s port", config.LoadBalancerPort, config.APIport)
	wg.Wait()
	// ---------
}

func logger(path string) *logrus.Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error: open log file")
	}
	l.Out = file
	return l
}

package main

import (
	"fmt"
	model "mcolomerc/synth-payment-producer/pkg/avro"
	"mcolomerc/synth-payment-producer/pkg/config"
	"mcolomerc/synth-payment-producer/pkg/datagen"
	"mcolomerc/synth-payment-producer/pkg/producer"
	"runtime"
	"strings"

	"time"

	"github.com/m-mizutani/zlog"
	"github.com/m-mizutani/zlog/filter"
)

var cnf config.Config
var wkf datagen.Workflow

var numPayments int
var paymentsCh chan model.Payment
var workers int

var kProd producer.Producer
var paymentGenerator datagen.Datagen
var workflowHandler datagen.Workflow

var logger *zlog.Logger

func init() {
	logger = zlog.New(zlog.WithFilters(filter.Tag()))

	// Read config from env vars
	cnf = config.Build()
	numPayments = cnf.Datagen.Payments
	workers = cnf.Datagen.Workers

	kProd = producer.NewProducer(cnf)
	paymentGenerator = datagen.NewDatagen(cnf.Datagen.Sources, cnf.Datagen.Destinations)
	workflowHandler = datagen.NewWorkflowHandler(cnf)
}

func main() {
	logger.Info("Starting producer...")
	logger.With("bootstrap.server", cnf.Kafka.BootstrapServers).Info("Using: ")

	numPayments := cnf.Datagen.Payments
	message := fmt.Sprintf("\n Starting producer... [%v] payments", numPayments)
	defer timer(message)()

	// Create topics
	logger.With("Topics", cnf.Kafka.Topics).Info("Topics: ")
	kProd.CreateTopics() // Create topics

	// Generate payments
	paymentsCh := make(chan model.Payment, numPayments)
	for i := 0; i < numPayments; i++ {
		logger.Info(" Generating payment...%v", i)
		payment := paymentGenerator.GeneratePayment() // Generate payment
		paymentsCh <- payment
	}
	done := make(chan bool, numPayments)
	workers := cnf.Datagen.Workers
	logger.Info(" Using workers: %v", workers)
	for i := 0; i < workers; i++ { // Spawn workers
		go worker(i, paymentsCh, done)
	}
	for i := 0; i < numPayments; i++ {
		<-done
	}
	kProd.Flush()
	kProd.Close()
}

/**
 * @title worker
 * @description
 * @param records
 * @param producer
 * @param done
 * @return
 */
func worker(w int, paymentsCh <-chan model.Payment, done chan<- bool) {
	defer timer("Worker")()
	for payment := range paymentsCh {
		wk := workflowHandler.GetWorkflow()
		logger.Info(" Worker-%v : Producing payment: %v : Workflow: %v", w, payment, wk)
		// Get workflow status
		statusDone := make(chan bool, len(wk))
		for i := range wk {
			go func(i int, payment model.Payment) {
				payment.Status = wk[i].String()
				delay := cnf.Datagen.Delays[strings.ToLower(payment.Status)] // Get delay by status
				time.Sleep(time.Duration(delay) * time.Millisecond)          // Apply delay
				payment.Ts = time.Now().UTC().UnixNano() / 1000000
				payment.Date_ts = time.Now().Format(time.RFC3339)
				logger.Info("\t Worker-%v : Producing payment status update: %v ", w, payment)
				kProd.Produce(payment)
				statusDone <- true
			}(i, payment)
		}
		for i := 0; i < len(wk); i++ {
			<-statusDone
		}
		close(statusDone)
		done <- true
	}
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		logger.Info("%s took %v\n", name, time.Since(start))
		logger.Info("Number of runnable goroutines: %v", runtime.NumGoroutine())
		PrintMemUsage()
	}
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	logger.Info("Alloc = %v MiB", m.Alloc/1024/1024)
	logger.Info("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	logger.Info("\tSys = %v MiB", m.Sys/1024/1024)
	logger.Info("\tNumGC = %v\n", m.NumGC)
}

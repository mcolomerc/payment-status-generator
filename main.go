package main

import (
	"fmt"
	"math/rand"
	model "mcolomerc/synth-payment-producer/pkg/avro"
	"mcolomerc/synth-payment-producer/pkg/config"
	"mcolomerc/synth-payment-producer/pkg/datagen"
	"mcolomerc/synth-payment-producer/pkg/producer"
	"mcolomerc/synth-payment-producer/pkg/stats"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/m-mizutani/zlog"
	"github.com/m-mizutani/zlog/filter"
	"github.com/mackerelio/go-osstat/cpu"
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

var sts *stats.Stats

func init() {
	logger = zlog.New(zlog.WithFilters(filter.Tag()))

	// Read config from env vars
	cnf = config.Build()
	numPayments = cnf.Datagen.Payments
	workers = cnf.Datagen.Workers

	kProd = producer.NewProducer(cnf)
	paymentGenerator = datagen.NewDatagen(cnf.Datagen.Sources, cnf.Datagen.Destinations)
	workflowHandler = datagen.NewWorkflowHandler(cnf)
	sts = stats.NewStats()
}

func main() {
	for _, st := range datagen.GetStatusList() {
		sts.AddState(st.String())
	}
	logger.Info("Starting producer...")
	logger.With("bootstrap.server", cnf.Kafka.BootstrapServers).Info("Using: ")
	numPayments := cnf.Datagen.Payments
	message := fmt.Sprintf("Generating... [%v] payments", numPayments)
	defer timer(message)()

	// Create topics
	logger.With("Topics", cnf.Kafka.Topics).Info("Topics: ")
	kProd.CreateTopics() // Create topics

	// Generate banks
	stop := make(chan bool, 1)
	interval := time.Duration(cnf.Datagen.UpdateBanksInterval)
	go buildBanks(time.NewTicker(interval*time.Millisecond), stop)
	// Generate payments
	paymentsCh := make(chan model.Payment, numPayments)
	for i := 0; i < numPayments; i++ {
		logger.Info(" Generating payment...%v", i)
		payment := paymentGenerator.GeneratePayment() // Generate payment
		paymentsCh <- payment
	}
	done := make(chan string, numPayments)
	workers := cnf.Datagen.Workers

	logger.Info(" Using workers: %v", workers)
	for i := 0; i < workers; i++ { // Spawn workers
		go worker(i, paymentsCh, done)
	}
	for i := 0; i < numPayments; i++ {
		workflow := <-done
		sts.AddWorkflow(workflow) // Add workflow
	}
	logger.Info("## Stops the bank updater ##")
	stop <- true
	// Close Producer
	kProd.Flush()
	kProd.Close()
	// Print stats
	sts.Print()
}

func buildBanks(ticker *time.Ticker, done <-chan bool) {
	logger.Info("## BANKS ## Generating banks")
	banks := paymentGenerator.GetBanks() // Get banks
	for {
		select {
		case <-ticker.C:
			logger.Info("## BANKS ## Updating bank ...")
			source := rand.NewSource(time.Now().UnixNano())
			rng := rand.New(source)
			randIdex := rng.Intn(len(banks))
			bank := banks[randIdex]
			bank.Updated_ts = time.Now().Format(time.RFC3339)
			bank.Version += 1
			banks[randIdex] = bank
			logger.Info(" Bank: %v", bank)
			kProd.ProduceBank(bank)
			sts.AddBank(bank.Name)
		case <-done:
			logger.Info("## BANKS ## Done")
			return
		}
	}
}

/**
 * Worker
 */
func worker(w int, paymentsCh <-chan model.Payment, done chan<- string) {
	for payment := range paymentsCh {
		wk := workflowHandler.GetWorkflow()
		logger.Info(" Worker-%v : Producing payment: %v : Workflow: %v", w, payment, wk)
		// Get workflow status
		statusDone := make(chan model.Payment, len(wk))
		for i := range wk {
			go func(i int, payment model.Payment) {
				payment.Status = wk[i].String()
				delay := cnf.Datagen.Delays[strings.ToLower(payment.Status)] // Get delay by status
				time.Sleep(time.Duration(delay) * time.Millisecond)          // Apply delay
				payment.Ts = time.Now().UTC().UnixNano() / 1000000
				payment.Date_ts = time.Now().Format(time.RFC3339)
				logger.Info("\t Worker-%v : Producing payment status update: %v ", w, payment)
				kProd.Produce(payment)
				statusDone <- payment
			}(i, payment)
		}
		for i := 0; i < len(wk); i++ {
			payment := <-statusDone
			sts.IncState(payment.Status) // Increment state counter
		}
		close(statusDone)
		done <- fmt.Sprintf("%v", wk)
	}
}

func timer(name string) func() {
	start := time.Now()
	return func() {

		logger.Info("----------------------------------------")
		logger.Info("%s took %v", name, time.Since(start))
		logger.Info("----------------------------------------")
		logger.Info("-------Go Routines---------")
		logger.Info("Number of runnable goroutines: %v", runtime.NumGoroutine())
		PrintMemUsage()
		GetCPUUsage()
		logger.Info("----------------------------------------")
	}
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	logger.Info("-------Memory------")
	logger.Info("Alloc = %v MiB", m.Alloc/1024/1024)
	logger.Info("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	logger.Info("\tSys = %v MiB", m.Sys/1024/1024)
	logger.Info("\tNumGC = %v", m.NumGC)
}

func GetCPUUsage() {
	logger.Info("-------CPU------")
	logger.Info("Num CPUs..." + strconv.Itoa(runtime.NumCPU()))
	before, err := cpu.Get()
	if err != nil {
		logger.Info("%s\n", err)
		return
	}
	time.Sleep(time.Duration(1) * time.Second)
	after, err := cpu.Get()
	if err != nil {
		logger.Info("%s\n", err)
		return
	}
	total := float64(after.Total - before.Total)
	logger.Info("CPU User: %f %%", float64(after.User-before.User)/total*100)
	logger.Info("CPU System: %f %%", float64(after.System-before.System)/total*100)
	logger.Info("CPU Idle: %f %%", float64(after.Idle-before.Idle)/total*100)
}

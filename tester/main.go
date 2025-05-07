package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/dapr/go-sdk/workflow"
)

const port = 5001

func main() {
	total := 2
	success := atomic.Int32{}
	fail := atomic.Int32{}
	mu := sync.Mutex{}
	latencies := []time.Duration{}

	client, err := workflow.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	wg := sync.WaitGroup{}
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Println("Starting workflow " + strconv.Itoa(i))
			workflowId := startWorkflow(client)
			fmt.Println("Workflow started " + workflowId)
			start := time.Now()

			completed := waitForWorkflowCompleted(client, workflowId)
			if completed {
				success.Add(1)
			} else {
				fail.Add(1)
			}
			mu.Lock()
			elapsed := time.Since(start)
			latencies = append(latencies, elapsed)
			mu.Unlock()
			fmt.Println("Workflow completed " + workflowId + " in " + elapsed.String())
		}(i)
	}
	wg.Wait()
	fmt.Println("Total: " + strconv.Itoa(total) + " Success: " + strconv.Itoa(int(success.Load())) + " Fail: " + strconv.Itoa(int(fail.Load())))
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	latenciesStr := []string{}
	for _, latency := range latencies {
		latenciesStr = append(latenciesStr, latency.String())
	}
	fmt.Println("Latencies: " + strings.Join(latenciesStr, ", "))
	// latency average
	totalLatency := time.Duration(0)
	for _, latency := range latencies {
		totalLatency += latency
	}
	averageLatency := totalLatency / time.Duration(len(latencies))
	fmt.Println("Average latency: " + averageLatency.String())
	// latency median
	medianLatency := latencies[len(latencies)/2]
	fmt.Println("Median latency: " + medianLatency.String())
	// latency 90th percentile
	percentile90 := latencies[int(float64(len(latencies))*0.9)]
	fmt.Println("90th percentile latency: " + percentile90.String())
	// latency 99th percentile
	percentile99 := latencies[int(float64(len(latencies))*0.99)]
	fmt.Println("99th percentile latency: " + percentile99.String())
	// latency max
	maxLatency := latencies[len(latencies)-1]
	fmt.Println("Max latency: " + maxLatency.String())
	// latency min
	minLatency := latencies[0]
	fmt.Println("Min latency: " + minLatency.String())

}

func startWorkflow(wfClient *workflow.Client) string {
	order := map[string]interface{}{
		"Name":     "Car",
		"Quantity": 1,
	}

	// uncomment to test java app
	// wfName := "io.dapr.quickstarts.workflows.OrderProcessingWorkflow"
	// use to test golang app
	// wfName := "OrderProcessingWorkflow"

	wfName := os.Getenv("WF_NAME")
	if wfName == "" {
		// default to java
		wfName = "io.dapr.quickstarts.workflows.OrderProcessingWorkflow"
	}

	workflowId, err := wfClient.ScheduleNewWorkflow(context.Background(), wfName, workflow.WithInput(order))
	if err != nil {
		log.Fatal(err)
	}
	return workflowId
}

func waitForWorkflowCompleted(wfClient *workflow.Client, workflowId string) bool {
	var out *workflow.Metadata
	err := backoff.Retry(func() error {
		fmt.Println("Waiting for workflow to complete " + workflowId)
		var err error
		out, err = wfClient.WaitForWorkflowCompletion(context.Background(), workflowId)
		if err != nil {
			return err
		}
		return nil
	}, backoff.WithContext(backoff.NewConstantBackOff(2*time.Second), context.Background()))
	if err != nil {
		return false
	}
	return out.RuntimeStatus == workflow.StatusCompleted
}

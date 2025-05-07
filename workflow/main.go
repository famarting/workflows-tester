package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/dapr/go-sdk/workflow"
	"github.com/dapr/kit/signals"
)

func main() {
	ctx := signals.Context()

	worker, err := workflow.NewWorker()
	if err != nil {
		log.Fatal(err)
	}
	defer worker.Shutdown()

	err = errors.Join(
		worker.RegisterWorkflow(OrderProcessingWorkflow),
		worker.RegisterActivity(NotifyActivity),
		worker.RegisterActivity(ProcessPaymentActivity),
		worker.RegisterActivity(ReserveInventoryActivity),
		worker.RegisterActivity(UpdateInventoryActivity),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = worker.Start()
	if err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
	fmt.Println("Worker stopped")
}

type OrderPayload struct {
	Name     string `json:"Name"`
	Quantity int    `json:"Quantity"`
}

type OrderResult struct {
	Processed bool `json:"Processed"`
}

type Notification struct {
	Message string `json:"Message"`
}

type ActivityResult struct {
	Success bool `json:"Success"`
}

func OrderProcessingWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	// orderId := ctx.InstanceID()
	var order OrderPayload
	err := ctx.GetInput(&order)
	if err != nil {
		return nil, err
	}
	// String orderId = ctx.getInstanceId();
	//   OrderPayload order = ctx.getInput(OrderPayload.class);
	//   OrderResult orderResult = new OrderResult();
	//   orderResult.setProcessed(false);
	var orderResult OrderResult
	orderResult.Processed = false

	// Notify the user that an order has come through
	err = ctx.CallActivity(NotifyActivity, workflow.ActivityInput(Notification{Message: fmt.Sprintf("Received Order %v", order)})).Await(nil)
	if err != nil {
		return nil, err
	}

	// Determine if there is enough of the item available for purchase by checking
	// the inventory
	var inventoryResult ActivityResult
	err = ctx.CallActivity(ReserveInventoryActivity, workflow.ActivityInput(order)).Await(&inventoryResult)
	if err != nil {
		return nil, err
	}
	//   If there is insufficient inventory, fail and let the user know
	if !inventoryResult.Success {
		err = ctx.CallActivity(NotifyActivity, workflow.ActivityInput(Notification{Message: fmt.Sprintf("Insufficient inventory for order: %v", order)})).Await(nil)
		if err != nil {
			return nil, err
		}
		return orderResult, nil
	}

	// There is enough inventory available so the user can purchase the item(s).
	// Process their payment
	var paymentResult ActivityResult
	err = ctx.CallActivity(ProcessPaymentActivity, workflow.ActivityInput(order)).Await(&paymentResult)
	if err != nil {
		return nil, err
	}
	if !paymentResult.Success {
		err = ctx.CallActivity(NotifyActivity, workflow.ActivityInput(Notification{Message: fmt.Sprintf("Payment failed for order: %v", order)})).Await(nil)
		if err != nil {
			return nil, err
		}
		return orderResult, nil
	}

	// Update the inventory
	var updateInventoryResult ActivityResult
	err = ctx.CallActivity(UpdateInventoryActivity, workflow.ActivityInput(order)).Await(&updateInventoryResult)
	if err != nil {
		return nil, err
	}
	if !updateInventoryResult.Success {
		err = ctx.CallActivity(NotifyActivity, workflow.ActivityInput(Notification{Message: fmt.Sprintf("Failed to update inventory for order: %v", order)})).Await(nil)
		if err != nil {
			return nil, err
		}
		return orderResult, nil
	}

	// Let user know their order was processed
	err = ctx.CallActivity(NotifyActivity, workflow.ActivityInput(Notification{Message: fmt.Sprintf("Order completed! : %v", order)})).Await(nil)
	if err != nil {
		return nil, err
	}

	// Complete the workflow with order result is processed
	orderResult.Processed = true
	return orderResult, nil
}

func NotifyActivity(ctx workflow.ActivityContext) (any, error) {
	var notification Notification
	err := ctx.GetInput(&notification)
	if err != nil {
		return nil, err
	}
	fmt.Println("NotifyActivity", notification.Message)
	return nil, nil
}

func ProcessPaymentActivity(ctx workflow.ActivityContext) (any, error) {
	// time.Sleep(500 * time.Millisecond)
	return ActivityResult{Success: true}, nil
}

func ReserveInventoryActivity(ctx workflow.ActivityContext) (any, error) {
	return ActivityResult{Success: true}, nil
}

func UpdateInventoryActivity(ctx workflow.ActivityContext) (any, error) {
	return ActivityResult{Success: true}, nil
}

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	synchronousproxy "github.com/temporalio/samples-go/synchronous-proxy"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	fmt.Println("T-Shirt Order")

	status, err := CreateOrder(c)
	if err != nil {
		log.Fatalln("Unable to create order", err)
	}

	for {
		email := PromptAndReadInput("Please enter you email address:")
		status, err = UpdateOrder(c, status.OrderID, synchronousproxy.RegisterStage, email)
		if err != nil {
			log.Println("invalid email", err)
			continue
		}

		break
	}

	for {
		size := PromptAndReadInput("Please enter your requested size:")
		status, err = UpdateOrder(c, status.OrderID, synchronousproxy.SizeStage, size)
		if err != nil {
			log.Println("invalid size", err)
			continue
		}

		break
	}

	for {
		color := PromptAndReadInput("Please enter your required tshirt color:")
		status, err = UpdateOrder(c, status.OrderID, synchronousproxy.ColorStage, color)
		if err != nil {
			log.Println("invalid color", err)
			continue
		}

		break
	}

	fmt.Println("Thanks for your order!")
	fmt.Println("You will receive an email with shipping details shortly")
}

func PromptAndReadInput(prompt string) string {
	fmt.Println(prompt)

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func CreateOrder(c client.Client) (synchronousproxy.OrderStatus, error) {
	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "ui-driven",
	}
	ctx := context.Background()
	var status synchronousproxy.OrderStatus

	we, err := c.ExecuteWorkflow(ctx, workflowOptions, synchronousproxy.OrderWorkflow)
	if err != nil {
		return status, fmt.Errorf("unable to execute order workflow: %w", err)
	}

	status.OrderID = we.GetID()

	return status, nil
}

func UpdateOrder(c client.Client, orderID string, stage string, value string) (synchronousproxy.OrderStatus, error) {
	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "ui-driven",
	}
	ctx := context.Background()
	status := synchronousproxy.OrderStatus{OrderID: orderID}

	we, err := c.ExecuteWorkflow(ctx, workflowOptions, synchronousproxy.UpdateOrderWorkflow, orderID, stage, value)
	if err != nil {
		return status, fmt.Errorf("unable to execute workflow: %w", err)
	}
	err = we.Get(ctx, &status)
	if err != nil {
		return status, err
	}

	return status, nil
}

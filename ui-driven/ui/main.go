package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	uidriven "github.com/temporalio/samples-go/ui-driven"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	fmt.Println("T-Shirt Order")

	var status uidriven.OrderStatus

	for {
		email := PromptAndReadInput("Please enter you email address:")
		status, err = StartOrder(c, email)
		if err != nil {
			log.Println("invalid email", err)
			continue
		}

		break
	}

	for {
		size := PromptAndReadInput("Please enter your requested size:")
		status, err = RecordSize(c, status.OrderID, size)
		if err != nil {
			log.Println("invalid size", err)
			continue
		}

		break
	}

	for {
		color := PromptAndReadInput("Please enter your required tshirt color:")
		status, err = RecordColor(c, status.OrderID, color)
		if err != nil {
			log.Println("invalid color", err)
			continue
		}

		break
	}

	fmt.Println("Thanks for your order!")
}

func PromptAndReadInput(prompt string) string {
	fmt.Println(prompt)

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func StartOrder(c client.Client, email string) (uidriven.OrderStatus, error) {
	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "ui-driven",
	}
	ctx := context.Background()
	var status uidriven.OrderStatus

	we, err := c.ExecuteWorkflow(ctx, workflowOptions, uidriven.StartOrderWorkflow, email)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	err = we.Get(ctx, &status)
	if err != nil {
		return status, err
	}

	return status, nil
}

func RecordSize(c client.Client, orderID string, size string) (uidriven.OrderStatus, error) {
	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "ui-driven",
	}
	ctx := context.Background()
	status := uidriven.OrderStatus{OrderID: orderID}

	we, err := c.ExecuteWorkflow(ctx, workflowOptions, uidriven.RecordSizeWorkflow, orderID, size)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	err = we.Get(ctx, &status)
	if err != nil {
		return status, err
	}

	return status, nil
}

func RecordColor(c client.Client, orderID string, color string) (uidriven.OrderStatus, error) {
	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "ui-driven",
	}
	ctx := context.Background()
	status := uidriven.OrderStatus{OrderID: orderID}

	we, err := c.ExecuteWorkflow(ctx, workflowOptions, uidriven.RecordColorWorkflow, orderID, color)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	err = we.Get(ctx, &status)
	if err != nil {
		return status, err
	}

	return status, nil
}

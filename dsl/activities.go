package dsl

import (
	"fmt"
)

func SampleActivity1(input []string) (string, error) {
	name := "sampleActivity1"
	fmt.Printf("Run %s with input %v \n", name, input)
	return "Result_" + name, nil
}

func SampleActivity2(input []string) (string, error) {
	name := "sampleActivity2"
	fmt.Printf("Run %s with input %v \n", name, input)
	return "Result_" + name, nil
}

func SampleActivity3(input []string) (string, error) {
	name := "sampleActivity3"
	fmt.Printf("Run %s with input %v \n", name, input)
	return "Result_" + name, nil
}

func SampleActivity4(input []string) (string, error) {
	name := "sampleActivity4"
	fmt.Printf("Run %s with input %v \n", name, input)
	return "Result_" + name, nil
}

func SampleActivity5(input []string) (string, error) {
	name := "sampleActivity5"
	fmt.Printf("Run %s with input %v \n", name, input)
	return "Result_" + name, nil
}

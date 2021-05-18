// @@@SNIPSTART samples-go-branch-activity-type
package branch

import (
	"fmt"
)

// SampleActivity is a Temporal Activity Type
func SampleActivity(input string) (string, error) {
	name := "sampleActivity"
	fmt.Printf("Run %s with input %v \n", name, input)
	return "Result_" + input, nil
}
// @@SNIPEND

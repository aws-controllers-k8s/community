package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

type Event struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, event Event) (string, error) {
	env := os.Environ()
	return fmt.Sprintf(`Hello from event: %s!
	environ: %v
`, event.Name, env), nil
}

func main() {
	lambda.Start(HandleRequest)
}

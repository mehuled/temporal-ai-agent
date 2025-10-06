package main

import (
	"crypto/tls"
	"log"
	"os"
	"strconv"
	"temporal-ai-agent/activities"
	"temporal-ai-agent/workflows"

	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Get configuration from environment variables
	hostPort := getEnv("TEMPORAL_HOST_PORT", "localhost:7233")
	namespace := getEnv("TEMPORAL_NAMESPACE", "default")
	apiKey := getEnv("TEMPORAL_API_KEY", "")
	taskQueue := getEnv("TEMPORAL_TASK_QUEUE", "my-task-queue")
	tlsEnabled := getEnvBool("TEMPORAL_TLS_ENABLED", false)

	// Validate required environment variables
	if apiKey == "" {
		log.Fatal("TEMPORAL_API_KEY environment variable is required")
	}

	// Configure client options
	clientOptions := client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
	}

	// Configure TLS if enabled
	if tlsEnabled {
		clientOptions.ConnectionOptions = client.ConnectionOptions{TLS: &tls.Config{}}
	}

	// Configure credentials
	clientOptions.Credentials = client.NewAPIKeyStaticCredentials(apiKey)

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, taskQueue, worker.Options{})

	w.RegisterWorkflow(workflows.SayHelloWorkflow)
	w.RegisterActivity(activities.Greet)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}

}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets a boolean environment variable with a fallback default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

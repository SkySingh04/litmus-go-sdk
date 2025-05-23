package examples

import (
	"os"

	"github.com/litmuschaos/litmus-go-sdk/pkg/logger"
	"github.com/litmuschaos/litmus-go-sdk/pkg/sdk"
	"github.com/litmuschaos/litmus-go-sdk/pkg/types"
	models "github.com/litmuschaos/litmus/chaoscenter/graphql/server/graph/model"
)

func CompleteSDKExample() {
	// Initialize the Litmus SDK client
	client, err := sdk.NewClient(sdk.ClientOptions{
		Endpoint: getEnv("LITMUS_ENDPOINT", "http://localhost:8080"),
		Username: getEnv("LITMUS_USERNAME", "admin"),
		Password: getEnv("LITMUS_PASSWORD", "litmus"),
	})
	if err != nil {
		logger.Fatalf("Failed to initialize client: %v", err)
	}

	// ======== Project Operations ========

	// List all projects
	projects, err := client.Projects().List()
	if err != nil {
		logger.Fatalf("Failed to list projects: %v", err)
	}
	logger.InfoWithValues("Projects", map[string]interface{}{
		"projects": projects.Data.Projects,
	})

	// Create a new project
	newProject, err := client.Projects().Create("my-new-sdk-project")
	if err != nil {
		logger.Fatalf("Failed to create project: %v", err)
	}
	logger.InfoWithValues("Created project", map[string]interface{}{
		"name": newProject.Data.Name,
		"id":   newProject.Data.ID,
	})

	// ======== Environment Operations ========

	// List environments
	environments, err := client.Environments().List()
	if err != nil {
		logger.Fatalf("Failed to list environments: %v", err)
	}
	logger.InfoWithValues("Environments", map[string]interface{}{
		"environments": environments,
	})

	// Create environment
	envRequest := models.CreateEnvironmentRequest{
		Name: "production",
		Type: "kubernetes",
		Tags: []string{"production", "litmus-sdk-example"},
	}
	newEnv, err := client.Environments().Create("production", envRequest)
	if err != nil {
		logger.Fatalf("Failed to create environment: %v", err)
	}
	logger.InfoWithValues("Created environment", map[string]interface{}{
		"environment": newEnv,
	})

	// ======== Experiment Operations ========

	// List experiments
	experiments, err := client.Experiments().List(models.ListExperimentRequest{})
	if err != nil {
		logger.Fatalf("Failed to list experiments: %v", err)
	}
	logger.InfoWithValues("Experiments", map[string]interface{}{
		"experiments": experiments,
	})

	// Create experiment
	expRequest := models.SaveChaosExperimentRequest{
		Name:        "nginx-availability-test",
		Description: "Test nginx pod availability under failure conditions",
		Tags:        []string{"availability", "nginx"},
	}
	newExp, err := client.Experiments().Create("nginx-availability-test", expRequest)
	if err != nil {
		logger.Fatalf("Failed to create experiment: %v", err)
	}
	logger.InfoWithValues("Created experiment", map[string]interface{}{
		"experiment": newExp,
	})

	// Run experiment
	runResult, err := client.Experiments().Run("experiment-id-123")
	if err != nil {
		logger.Fatalf("Failed to run experiment: %v", err)
	}
	logger.InfoWithValues("Experiment run", map[string]interface{}{
		"result": runResult,
	})

	// ======== Infrastructure Operations ========

	// List infrastructure
	infraList, err := client.Infrastructure().List()
	if err != nil {
		logger.Fatalf("Failed to list infrastructure: %v", err)
	}
	logger.InfoWithValues("Infrastructure", map[string]interface{}{
		"infrastructure": infraList,
	})

	// Create infrastructure
	infraConfig := types.Infra{
		InfraName:      "gcp-cluster",
		Description:    "GCP Kubernetes cluster",
		PlatformName:   "gcp",
		Namespace:      "litmus",
		ServiceAccount: "litmus-admin",
		NsExists:       true,
		SAExists:       true,
		Mode:           "namespace",
	}
	newInfra, err := client.Infrastructure().Create("gcp-cluster", infraConfig)
	if err != nil {
		logger.Fatalf("Failed to create infrastructure: %v", err)
	}
	logger.InfoWithValues("Created infrastructure", map[string]interface{}{
		"infrastructure": newInfra,
	})

	// ======== Probe Operations ========

	// List probes
	probes, err := client.Probes().List("probe-id-123")
	if err != nil {
		logger.Fatalf("Failed to list probes: %v", err)
	}
	logger.InfoWithValues("Probes", map[string]interface{}{
		"probes": probes,
	})
	
	// Get probe YAML
	yamlRequest := models.GetProbeYAMLRequest{
		ProbeName: "my-probe-name",
		Mode: "SOT",
	}
	probeYAML, err := client.Probes().GetProbeYAML("project-id", "probe-id", yamlRequest)
	if err != nil {
		logger.Fatalf("Failed to get probe YAML: %v", err)
	}
	logger.InfoWithValues("Probe YAML", map[string]interface{}{
		"yaml": probeYAML,
	})

}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

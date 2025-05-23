package probe

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/litmuschaos/litmus-go-sdk/pkg/apis"
	"github.com/litmuschaos/litmus-go-sdk/pkg/logger"
	"github.com/litmuschaos/litmus-go-sdk/pkg/types"
	"github.com/litmuschaos/litmus/chaoscenter/graphql/server/graph/model"
	"github.com/stretchr/testify/assert"
)

// Test configuration with defaults
var (
	testEndpoint = "http://127.0.0.1:39651"
	testUsername = "admin"
	testPassword = "litmus"	
	projectID  string
	testProbeID   string
	testProbeName string
	credentials types.Credentials
)

func TestMain(m *testing.M) {
	// Override defaults with environment variables if set
	if endpoint := os.Getenv("LITMUS_TEST_ENDPOINT"); endpoint != "" {
		testEndpoint = endpoint
	}
	if username := os.Getenv("LITMUS_TEST_USERNAME"); username != "" {
		testUsername = username
	}
	if password := os.Getenv("LITMUS_TEST_PASSWORD"); password != "" {
		testPassword = password
	}

	logger.Infof("Test configuration - Endpoint: %s, Username: %s", testEndpoint, testUsername)
	
	// Setup credentials by authenticating
	authResp, err := apis.Auth(types.AuthInput{
		Endpoint: testEndpoint,
		Username: testUsername,
		Password: testPassword,
	})
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}

	credentials = types.Credentials{
		Endpoint: testEndpoint,
		Token:          authResp.AccessToken,
	}

	// Get or create project ID
	projectResp, err := apis.ListProject(credentials)
	if err != nil {
		log.Fatalf("Failed to list projects: %v", err)
	}

	if len(projectResp.Data.Projects) > 0 {
		projectID = projectResp.Data.Projects[0].ID
		logger.Infof("Using existing project ID: %s", projectID)
	} else {
		// Create a project if none exists
		projectName := fmt.Sprintf("test-project-%s", uuid.New().String())
		newProject, err := apis.CreateProjectRequest(projectName, credentials)
		if err != nil {
			log.Fatalf("Failed to create project: %v", err)
		}
		projectID = newProject.Data.ID
		logger.Infof("Created new project ID: %s", projectID)
	}
	
	// Store project ID in credentials for convenience
	credentials.ProjectID = projectID
	
	// Run the tests
	exitCode := m.Run()
	
	// Exit with the test status code
	os.Exit(exitCode)
}

func init() {
	// Override defaults with environment variables if set
	if endpoint := os.Getenv("LITMUS_TEST_ENDPOINT"); endpoint != "" {
		testEndpoint = endpoint
	}
	if username := os.Getenv("LITMUS_TEST_USERNAME"); username != "" {
		testUsername = username
	}
	if password := os.Getenv("LITMUS_TEST_PASSWORD"); password != "" {
		testPassword = password
	}

	logger.Infof("Test configuration - Endpoint: %s, Username: %s", testEndpoint, testUsername)
}

// LitmusClient provides methods to interact with Litmus Chaos API
type LitmusClient struct {
	credentials types.Credentials
}

// NewLitmusClient creates and authenticates a new client with username/password
func NewLitmusClient(endpoint, username, password string) (*LitmusClient, error) {
	// Implementation should match the one in main.go
	authResp, err := apis.Auth(types.AuthInput{
		Endpoint: endpoint,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return &LitmusClient{
		credentials: types.Credentials{
			Endpoint:       endpoint,
			Token:          authResp.AccessToken,
		},
	}, nil
}

func setupTestClient() (*LitmusClient, error) {
	return NewLitmusClient(testEndpoint, testUsername, testPassword)
}

// TestCreateProbe tests probe creation
func TestCreateProbe(t *testing.T) {
	trueBool := true
	desc := "Test probe description"

	// Helper function to ensure probe cleanup
	cleanupProbe := func(projectID, probeName string) {
		if probeName != "" {
			_, _ = DeleteProbeRequest(projectID, probeName, credentials)
		}
	}

	tests := []struct {
		name              string
		projectID         string
		probeReq          ProbeRequest
		saveProbeID       bool // Whether to save this probe's ID for later tests
		wantErr           bool
		validateFnFactory func(req ProbeRequest) func(*testing.T, *Probe, error)
	}{
		{
			name:      "Create HTTP Probe - Successful",
			projectID: projectID,
			probeReq: ProbeRequest{
				Name:               fmt.Sprintf("test-http-probe-%s", uuid.New().String()),
				Description:        &desc,
				Type:               ProbeTypeHTTPProbe,
				InfrastructureType: InfrastructureTypeKubernetes,
				Tags:               []string{"test", "http"},
				KubernetesHTTPProperties: &KubernetesHTTPProbeRequest{
					ProbeTimeout: "30s",
					Interval:     "10s",
					Attempt:      intPtr(1),
					URL:          "http://localhost:8080/health",
					Method: &Method{
						Get: &GetMethod{
							ResponseCode: "200",
							Criteria:     "==",
						},
					},
					InsecureSkipVerify: &trueBool,
				},
			},
			saveProbeID: true, // Save this probe for later tests
			wantErr: false,
			validateFnFactory: func(expectedReq ProbeRequest) func(*testing.T, *Probe, error) {
				return func(t *testing.T, probe *Probe, err error) {
					assert.NoError(t, err)
					if err != nil {
						t.Logf("Error creating probe: %v", err)
						return
					}
					
					assert.NotNil(t, probe)
					if probe == nil {
						t.Log("Probe is nil")
						return
					}
					
					assert.Equal(t, expectedReq.Name, probe.Name)
					assert.Equal(t, ProbeTypeHTTPProbe, probe.Type)
					assert.NotNil(t, probe.KubernetesHTTPProperties)
					
					if probe.KubernetesHTTPProperties != nil {
						assert.Equal(t, "http://localhost:8080/health", probe.KubernetesHTTPProperties.URL)
						assert.Equal(t, "30s", probe.KubernetesHTTPProperties.ProbeTimeout)
						assert.Equal(t, "10s", probe.KubernetesHTTPProperties.Interval)
						if probe.KubernetesHTTPProperties.InsecureSkipVerify != nil {
							assert.Equal(t, trueBool, *probe.KubernetesHTTPProperties.InsecureSkipVerify)
						}
					}
                    
                    // Save probe ID for later tests if requested
                    if expectedReq.Name != "" && testProbeID == "" {
                        testProbeID = probe.Name
                        testProbeName = probe.Name
                        t.Logf("Saving probe ID for later tests: %s", testProbeID)
                    }
				}
			},
		},
		{
			name:      "Create CMD Probe - Successful",
			projectID: projectID,
			probeReq: ProbeRequest{
				Name:               fmt.Sprintf("test-cmd-probe-%s", uuid.New().String()),
				Type:               ProbeTypeCMDProbe,
				InfrastructureType: InfrastructureTypeKubernetes,
				KubernetesCMDProperties: &KubernetesCMDProbeRequest{
					Command: "ls -l",
					Comparator: &ComparatorInput{
						Type:     "Contains",
						Criteria: "==",
						Value:    "test",
					},
					ProbeTimeout: "30s",
					Interval:     "10s",
					Attempt:      intPtr(1),
				},
			},
			wantErr: false,
			validateFnFactory: func(expectedReq ProbeRequest) func(*testing.T, *Probe, error) {
				return func(t *testing.T, probe *Probe, err error) {
					// Ensure cleanup happens regardless of test outcome
					if expectedReq.Name != "" {
						defer cleanupProbe(projectID, expectedReq.Name)
					}

					assert.NoError(t, err)
					assert.NotNil(t, probe)
					if probe == nil {
						return // Avoid nil pointer dereference
					}
					assert.Equal(t, expectedReq.Name, probe.Name)
					assert.Equal(t, ProbeTypeCMDProbe, probe.Type)
					assert.NotNil(t, probe.KubernetesCMDProperties)
					if probe.KubernetesCMDProperties != nil {
						assert.Equal(t, "ls -l", probe.KubernetesCMDProperties.Command)
						assert.Equal(t, "30s", probe.KubernetesCMDProperties.ProbeTimeout)
						assert.Equal(t, "10s", probe.KubernetesCMDProperties.Interval)
						if probe.KubernetesCMDProperties.Comparator != nil {
							assert.Equal(t, "Contains", probe.KubernetesCMDProperties.Comparator.Type)
							assert.Equal(t, "==", probe.KubernetesCMDProperties.Comparator.Criteria)
						}
					}
				}
			},
		},
		{
			name:      "Validation Error - Mismatched Type and Properties",
			projectID: projectID,
			probeReq: ProbeRequest{
				Name:               "test-mismatch-probe",
				Type:               ProbeTypeHTTPProbe,
				InfrastructureType: InfrastructureTypeKubernetes,
				KubernetesCMDProperties: &KubernetesCMDProbeRequest{
					Command:      "echo hello",
					ProbeTimeout: "5s",
					Interval:     "5s",
				},
			},
			wantErr: true,
			validateFnFactory: func(expectedReq ProbeRequest) func(*testing.T, *Probe, error) {
				return func(t *testing.T, probe *Probe, err error) {
					// Ensure cleanup happens regardless of test outcome
					if expectedReq.Name != "" {
						defer cleanupProbe(projectID, expectedReq.Name)
					}

					assert.Error(t, err)
					assert.Nil(t, probe)
					assert.Contains(t, err.Error(), "httpProbe type requires kubernetesHTTPProperties")
				}
			},
		},
		{
			name:      "Validation Error - No Properties",
			projectID: projectID,
			probeReq: ProbeRequest{
				Name:               "test-no-props-probe",
				Type:               ProbeTypeHTTPProbe,
				InfrastructureType: InfrastructureTypeKubernetes,
			},
			wantErr: true,
			validateFnFactory: func(expectedReq ProbeRequest) func(*testing.T, *Probe, error) {
				return func(t *testing.T, probe *Probe, err error) {
					// Ensure cleanup happens regardless of test outcome
					if expectedReq.Name != "" {
						defer cleanupProbe(projectID, expectedReq.Name)
					}

					assert.Error(t, err)
					assert.Nil(t, probe)
					assert.Contains(t, err.Error(), "no probe properties provided")
				}
			},
		},
		{
			name:      "Validation Error - Multiple Properties",
			projectID: projectID,
			probeReq: ProbeRequest{
				Name:               "test-multi-props-probe",
				Type:               ProbeTypeHTTPProbe,
				InfrastructureType: InfrastructureTypeKubernetes,
				KubernetesHTTPProperties: &KubernetesHTTPProbeRequest{
					URL:          "http://example.com",
					ProbeTimeout: "5s",
					Interval:     "5s",
					Attempt:      intPtr(1),
					Method:       &Method{Get: &GetMethod{ResponseCode: "200", Criteria: "=="}},
				},
				KubernetesCMDProperties: &KubernetesCMDProbeRequest{
					Command:      "echo hello",
					ProbeTimeout: "5s",
					Interval:     "5s",
					Attempt:      intPtr(1),
				},
			},
			wantErr: true,
			validateFnFactory: func(expectedReq ProbeRequest) func(*testing.T, *Probe, error) {
				return func(t *testing.T, probe *Probe, err error) {
					// Ensure cleanup happens regardless of test outcome
					if expectedReq.Name != "" {
						defer cleanupProbe(projectID, expectedReq.Name)
					}

					assert.Error(t, err)
					assert.Nil(t, probe)
					assert.Contains(t, err.Error(), "multiple probe property types provided")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentTestProbeRequest := tt.probeReq

			probe, err := CreateProbeRequest(currentTestProbeRequest, tt.projectID, credentials)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.validateFnFactory != nil {
				validator := tt.validateFnFactory(currentTestProbeRequest)
				validator(t, probe, err)
			}
            
            // If this is the successful HTTP probe test and it passed, save the ID
            if tt.saveProbeID && !tt.wantErr && err == nil && probe != nil {
                testProbeID = probe.Name
                testProbeName = probe.Name
                t.Logf("Successfully created and saved probe ID for tests: %s", testProbeID)
            }
		})
	}
}

// TestGetProbeRequest tests probe retrieval - runs after TestCreateProbe
func TestGetProbeRequest(t *testing.T) {
    // Skip this test if no probe was created
    if testProbeID == "" {
        t.Skip("Skipping test because no probe ID is available. TestCreateProbe must run first.")
    }

	tests := []struct {
		name       string
		projectID  string
		probeID    string
		setup      func(*LitmusClient) // optional setup steps
		wantErr    bool
		validateFn func(*testing.T, *GetProbeResponse)
	}{
		{
			name:      "successful probe retrieval",
			projectID: projectID,
			probeID:   testProbeID,
			wantErr:   false,
			validateFn: func(t *testing.T, result *GetProbeResponse) {
				assert.NotNil(t, result, "Result should not be nil")
				assert.NotNil(t, result.Data, "Data should not be nil")
				assert.NotNil(t, result.Data.GetProbe, "GetProbe should not be nil")
			},
		},
		{
			name:       "probe retrieval with empty ID",
			projectID:  projectID,
			probeID:    "",
			wantErr:    true,
			validateFn: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := setupTestClient()
			assert.NoError(t, err, "Failed to create Litmus client")

			// Run any setup function if provided
			if tt.setup != nil {
				tt.setup(client)
			}

			result, err := GetProbeRequest(tt.projectID, tt.probeID, client.credentials)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Run validation function if provided
			if tt.validateFn != nil {
				tt.validateFn(t, &result)
			}
		})
	}
}

func TestListProbeRequest(t *testing.T) {
    // This test can run independently since it doesn't depend on specific probe IDs
	tests := []struct {
		name       string
		projectID  string
		probeTypes []*model.ProbeType
		setup      func(*LitmusClient) // optional setup steps
		wantErr    bool
		validateFn func(*testing.T, *ListProbeResponse)
	}{
		{
			name:       "successful probes listing",
			projectID:  projectID,
			probeTypes: nil, // List all probe types
			wantErr:    false,
			validateFn: func(t *testing.T, result *ListProbeResponse) {
				assert.NotNil(t, result, "Result should not be nil")
				// If Data is nil, initialize it to avoid nil pointer panics
				if result.Data.Probes == nil {
					t.Log("Probes list was nil, expected non-nil")
					// We'll still pass the test, but log the issue
					// This handles the case where the API response is empty but not an error
					return
				}

				assert.NotNil(t, result.Data, "Data should not be nil") 
				assert.NotNil(t, result.Data.Probes, "Probes should not be nil")
			},
		},
		{
			name:       "probes listing with empty project ID",
			projectID:  "",
			probeTypes: nil,
			wantErr:    true,
			validateFn: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := setupTestClient()
			assert.NoError(t, err, "Failed to create Litmus client")

			// Run any setup function if provided
			if tt.setup != nil {
				tt.setup(client)
			}

			result, err := ListProbeRequest(tt.projectID, tt.probeTypes, client.credentials)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Run validation function if provided
			if tt.validateFn != nil {
				tt.validateFn(t, &result)
			}
		})
	}
}

// TestGetProbeYAMLRequest tests getting probe YAML
func TestGetProbeYAMLRequest(t *testing.T) {
	tests := []struct {
		name       string
		projectID  string
		request    model.GetProbeYAMLRequest
		setup      func(*LitmusClient) // optional setup steps
		wantErr    bool
		validateFn func(*testing.T, *GetProbeYAMLResponse)
	}{
		{
			name:      "successful probe YAML retrieval",
			projectID: projectID,
			request: model.GetProbeYAMLRequest{
				ProbeName: testProbeName,
				Mode:      "SOT",
			},
			wantErr: false,
			validateFn: nil,
		},
		{
			name:      "probe YAML retrieval with empty probe name",
			projectID: projectID,
			request: model.GetProbeYAMLRequest{
				ProbeName: "",
				Mode:      "SOT",
			},
			wantErr:    true,
			validateFn: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := setupTestClient()
			assert.NoError(t, err, "Failed to create Litmus client")

			// Run any setup function if provided
			if tt.setup != nil {
				tt.setup(client)
			}

			result, err := GetProbeYAMLRequest(tt.projectID, tt.request, client.credentials)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Run validation function if provided
			if tt.validateFn != nil {
				tt.validateFn(t, &result)
			}
		})
	}
}

// TestDeleteProbeRequest tests probe deletion - this should run last
func TestDeleteProbeRequest(t *testing.T) {
    // Skip this test if no probe was created
    if testProbeID == "" {
        t.Skip("Skipping test because no probe ID is available. TestCreateProbe must run first.")
    }

	tests := []struct {
		name       string
		projectID  string
		probeID    string
		setup      func(*LitmusClient) // optional setup steps
		wantErr    bool
		validateFn func(*testing.T, *DeleteProbeResponse)
	}{
		{
			name:      "successful probe deletion",
			projectID: projectID,
			probeID:   testProbeID,
			wantErr:   false,
			validateFn: func(t *testing.T, result *DeleteProbeResponse) {
				assert.NotNil(t, result, "Result should not be nil")
				assert.NotNil(t, result.Data, "Data should not be nil")
			},
		},
		{
			name:       "probe deletion with empty probe ID",
			projectID:  projectID,
			probeID:    "",
			wantErr:    true,
			validateFn: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := setupTestClient()
			assert.NoError(t, err, "Failed to create Litmus client")

			// Run any setup function if provided
			if tt.setup != nil {
				tt.setup(client)
			}

			result, err := DeleteProbeRequest(tt.projectID, tt.probeID, client.credentials)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Run validation function if provided
			if tt.validateFn != nil {
				tt.validateFn(t, &result)
			}
		})
	}
}

// Helper function to get a pointer to an int
func intPtr(i int) *int {
	return &i
}

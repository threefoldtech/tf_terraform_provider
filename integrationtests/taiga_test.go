//go:build integration
// +build integration

// Package integrationtests includes integration tests for deploying solutions on the tf grid, and some utilities to test these solutions.
package integrationtests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"
	"github.com/threefoldtech/terraform-provider-grid/internal/provider/scheduler"
)

func TestTaiga(t *testing.T) {
	/* Test case for deploying a presearch.

	   **Test Scenario**

	   - Deploy a taiga.
	   - Check that the outputs not empty.
	   - Check that vm is reachable.
	   - Check that env variables set successfully.
	   - Check taiga zinit service is running
	   - Destroy the deployment.
	*/

	// retryable errors in terraform testing.
	// generate ssh keys for test
	publicKey, privateKey, err := GenerateSSHKeyPair()
	if err != nil {
		t.Fatalf("failed to generate ssh key pair: %s", err.Error())
	}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./taiga",
		Vars: map[string]interface{}{
			"public_key": publicKey,
		},
	})
	defer terraform.Destroy(t, terraformOptions)

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil &&
		(strings.Contains(err.Error(), scheduler.NoNodesFoundErr.Error()) ||
			strings.Contains(err.Error(), "error creating threefold plugin client")) {
		t.Skip("couldn't find any available nodes")
		return
	}

	require.NoError(t, err)

	// Check that the outputs not empty
	myCeliumIP := terraform.Output(t, terraformOptions, "mycelium_ip")
	require.NotEmpty(t, myCeliumIP)

	fqdn := terraform.Output(t, terraformOptions, "fqdn")
	require.NotEmpty(t, fqdn)

	ok := TestConnection(myCeliumIP, "22")
	require.True(t, ok)

	output, err := RemoteRun("root", myCeliumIP, "zinit list", privateKey)
	require.NoError(t, err)
	require.Contains(t, output, "taiga: Running")

	statusOk := false
	ticker := time.NewTicker(2 * time.Second)
	// taiga takes alot of time to be ready
	for now := time.Now(); time.Since(now) < 10*time.Minute; {
		<-ticker.C
		resp, err := http.Get(fmt.Sprintf("https://%s", fqdn))
		if err == nil && resp.StatusCode == 200 {
			statusOk = true
			break
		}
	}

	require.True(t, statusOk, "website did not respond with 200 status code")
}

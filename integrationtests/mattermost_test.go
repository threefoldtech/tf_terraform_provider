//go:build integration
// +build integration

package integrationtests

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	tests "github.com/threefoldtech/terraform-provider-grid/integrationtests"
)

func TestMattermostDeployment(t *testing.T) {
	/* Test case for deploying a matermost.

	   **Test Scenario**

	   - Deploy a matermost.
	   - Check that the outputs not empty.
	   - Check that vm is reachable
	   - Check that env variables set successfully
	   - Destroy the deployment
	*/

	// retryable errors in terraform testing.
	// generate ssh keys for test
	pk, sk, err := tests.GenerateSSHKeyPair()
	if err != nil {
		t.Log(err)
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./mattermost",
		Vars: map[string]interface{}{
			"public_key": pk,
		},
		Parallelism: 1,
	})
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// Check that the outputs not empty
	planetary := terraform.Output(t, terraformOptions, "ygg_ip")
	assert.NotEmpty(t, planetary)
	fqdn := terraform.Output(t, terraformOptions, "fqdn")
	assert.NotEmpty(t, fqdn)

	// Check that env variables set successfully
	output, err := tests.RemoteRun("root", planetary, "cat /proc/1/environ", sk)
	assert.NoError(t, err)
	assert.Contains(t, string(output), "SSH_KEY")

	// Check that the solution is running successfully

	output, err = tests.RemoteRun("root", planetary, "zinit list", sk)
	assert.NoError(t, err)
	assert.Contains(t, output, "mattermost: Running")

}

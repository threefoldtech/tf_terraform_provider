//go:build integration
// +build integration

package integrationtests

import (
	"os/exec"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestTaigaDeployment(t *testing.T) {
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
		t.Fatal()
	}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./taiga",
		Vars: map[string]interface{}{
			"public_key": publicKey,
		},
		Parallelism: 1,
	})
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// Check that the outputs not empty
	planetary := terraform.Output(t, terraformOptions, "node1_zmachine1_ygg_ip")
	assert.NotEmpty(t, planetary)

	webIp := terraform.Output(t, terraformOptions, "fqdn")
	assert.NotEmpty(t, webIp)

	// Check that env variables set successfully
	output, err := RemoteRun("root", planetary, "cat /proc/1/environ", privateKey)
	assert.NoError(t, err)
	assert.Contains(t, string(output), "SSH_KEY")

	output, err = RemoteRun("root", planetary, "zinit list", privateKey)
	assert.NoError(t, err)
	assert.Contains(t, output, "taiga: Running")

}

package test

import (
	"log"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/threefoldtech/terraform-provider-grid/tests"
)

func TestSingleNodeWithSmallMemDeployment(t *testing.T) {
	// retryable errors in terraform testing.
	// generate ssh keys for test
	pk, _, err := tests.SshKeys()
	if err != nil {
		log.Fatal(err)
	}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./",
		Vars: map[string]interface{}{
			"public_key": pk,
		},
		Parallelism: 1,
	})
	defer terraform.Destroy(t, terraformOptions)

	_, err = terraform.InitAndApplyE(t, terraformOptions)

	if err == nil {
		t.Errorf("Should fail with mem capacity can't be less that 250M but err is null")
	}

}

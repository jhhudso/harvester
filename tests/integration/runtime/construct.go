package runtime

import (
	"context"
	"fmt"

	restclient "k8s.io/client-go/rest"

	"github.com/rancher/harvester/tests/framework/client"
	"github.com/rancher/harvester/tests/framework/env"
	"github.com/rancher/harvester/tests/framework/helm"
	"github.com/rancher/harvester/tests/framework/ready"
)

// Construct prepares runtime if "USE_EXISTING_RUNTIME" is not "true".
func Construct(ctx context.Context, kubeConfig *restclient.Config) error {
	if env.IsUsingExistingRuntime() {
		return nil
	}

	// create namespace
	err := client.CreateNamespace(kubeConfig, testHarvesterNamespace)
	if err != nil {
		return fmt.Errorf("failed to create target namespace, %v", err)
	}

	// install harvester chart
	err = installHarvesterChart(ctx, kubeConfig)
	if err != nil {
		return fmt.Errorf("failed to install harvester chart, %w", err)
	}

	return nil
}

// installHarvesterChart installs the basic components of harvester.
func installHarvesterChart(ctx context.Context, kubeConfig *restclient.Config) error {
	// chart values patches
	patches := map[string]interface{}{
		"replicas":               0,
		"minio.service.type":     "NodePort",
		"minio.mode":             "standalone",
		"minio.persistence.size": "5Gi",
	}

	if env.IsUsingEmulation() {
		patches["kubevirt.spec.configuration.developerConfiguration.useEmulation"] = true
	}

	// install chart
	_, err := helm.InstallChart(testChartReleaseName, testHarvesterNamespace, testChartDir, patches)
	if err != nil {
		return fmt.Errorf("failed to install harvester chart: %w", err)
	}

	// verifies chart installation
	namespaceReadyCondition, err := ready.NewNamespaceCondition(kubeConfig, testHarvesterNamespace)
	if err != nil {
		return fmt.Errorf("faield to create namespace ready condition from kubernetes config: %w", err)
	}
	namespaceReadyCondition.AddDeploymentsReady(testDeploymentManifest...)
	namespaceReadyCondition.AddDaemonSetsReady(testDaemonSetManifest...)

	return namespaceReadyCondition.Wait(ctx)
}

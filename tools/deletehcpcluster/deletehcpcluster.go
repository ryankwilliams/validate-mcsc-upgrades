// Program used to easily delete a rosa hosted control plane cluster that was
// created as part of the test suite `BeforeSuite` and was not properly
// deleted.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openshift/osde2e-framework/pkg/clients/ocm"
	"github.com/openshift/osde2e-framework/pkg/providers/rosa"
)

// getInput returns the user input and validates it is valid
func getInput(message string) (string, error) {
	var input string
	fmt.Printf("%s: ", message)
	fmt.Scanln(&input)

	if input == "" {
		return "", fmt.Errorf("%s is empty, please provide it and try again", message)
	}
	return input, nil
}

func main() {
	var (
		clusterID    string
		clusterName  string
		err          error
		ocmToken     string
		rosaProvider *rosa.Provider
	)

	clusterName, err = getInput("Cluster name")
	if err != nil {
		panic(err)
	}

	clusterID, err = getInput("Cluster id")
	if err != nil {
		panic(err)
	}

	ocmToken = os.Getenv("OCM_TOKEN")
	if ocmToken == "" {
		ocmToken, err = getInput("OCM token")
		if err != nil {
			panic(err)
		}
	}

	ctx := context.Background()

	rosaProvider, err = rosa.New(ctx, ocmToken, ocm.Integration)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Deleting cluster %s/%s\n", clusterName, clusterID)

	err = rosaProvider.DeleteCluster(ctx, &rosa.DeleteClusterOptions{
		ClusterName: clusterName,
		ClusterID:   clusterID,
		HostedCP:    true,
	})
	if err != nil {
		panic(err)
	}
}

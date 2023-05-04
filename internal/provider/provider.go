package provider

import "fmt"

func AddHCPWorkloads() {
	for i := 1; i <= 2; i++ {
		fmt.Printf("Provision hcp cluster #%d\n", i)
	}
}

func RemoveHCPWorkloads() {
	for i := 1; i <= 2; i++ {
		fmt.Printf("Destroy hcp cluster #%d\n", i)
	}
}

package deploy

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func StartDeployment(deploymentParams Parameters) error {
	//deploymentParams := initializeParameters()

	credential, err := azidentity.NewClientSecretCredential(*deploymentParams.TenantId, *deploymentParams.SpnId, *deploymentParams.SpnSecret, nil)
	if err != nil {
		PrintHeaderToConsole("Deployment result")
		return fmt.Errorf("unable to generate a secret credential %v", err)
	}

	deployedPackageVersion, err := GetDeployedPackageVersion(deploymentParams, credential)

	if err != nil {
		PrintHeaderToConsole("Deployment result")
		return fmt.Errorf("unable to get deployed package : %v", err)
	}

	if *deploymentParams.VersionToDeploy == deployedPackageVersion {
		PrintHeaderToConsole("Deployment result")
		fmt.Printf("The deployed package (%s) is the same as the one you want to deploy (%s). Nothing to do. \n", deployedPackageVersion, *deploymentParams.VersionToDeploy)
		return nil
	}

	fmt.Printf("The deployed package (%s) is different from the one you want to deploy (%s). Let's deploy it ! \n", deployedPackageVersion, *deploymentParams.VersionToDeploy)

	err = Deploy(deploymentParams, credential)
	if err != nil {
		PrintHeaderToConsole("Deployment result")
		return err
	}

	PrintHeaderToConsole("Deployment result")
	fmt.Println("Package deployed with success !")
	return nil
}

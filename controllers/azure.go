package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"log"
	"os"
)

const containerWeb, fileToCheck, metadataToCheck = "$web", "index.html", "version"
const azureStorageAccountName, azureAccountKey = "yamaalgolia", "XXXX"

func main() {
	os.Setenv("AZURE_STORAGE_ACCOUNT_KEY", azureAccountKey)

	versionIsDeployed, version, err := isAlreadyDeployed("v1.5.0")
	if err != nil {
		fmt.Errorf("oups %s", err.Error())
	}

	fmt.Printf("is Deployed  %v \n", versionIsDeployed)
	fmt.Printf("Version %s \n", version)
}

func isAlreadyDeployed(versionToDeploy string) (bool, string, error) {
	accountKey, ok := os.LookupEnv("AZURE_STORAGE_ACCOUNT_KEY")

	if !ok {
		println("AZURE_STORAGE_ACCOUNT_KEY could not be found")
	}

	ctx := context.Background()

	credential, err := azblob.NewSharedKeyCredential(azureStorageAccountName, accountKey)
	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
	}

	serviceClient, err := azblob.NewServiceClientWithSharedKey(fmt.Sprintf("https://%s.blob.core.windows.net/", azureStorageAccountName), credential, nil)
	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
	}

	client, err := serviceClient.NewContainerClient(containerWeb)
	if err != nil {
		log.Fatalf("Unable to create a client on %s container", containerWeb)
	}

	_, err = client.GetProperties(ctx, nil)
	if err != nil {
		log.Fatalf("Error when fetching properties on the storage account %s with the following error \n %v ", containerWeb, err)
	}

	a := []azblob.ListBlobsIncludeItem{"metadata"}
	options := azblob.ContainerListBlobsFlatOptions{
		Include: a,
	}

	pager := client.ListBlobsFlat(&options)

	isAlreadyDeployed := false
	var deployedVersion string

	fileIsFound, metadataIsFound := false, false
	for pager.NextPage(ctx) {
		resp := pager.PageResponse()
		for _, v := range resp.ListBlobsFlatSegmentResponse.Segment.BlobItems {
			if *v.Name == fileToCheck {
				fileIsFound = true
				fmt.Printf("Found %s \n", fileToCheck)

				for key, v := range v.Metadata {
					if key == metadataToCheck {
						deployedVersion = *v
						isAlreadyDeployed = *v == versionToDeploy
						metadataIsFound = true
						fmt.Printf("Metadata %v %v \n", key, *v)
					}
				}
				if !metadataIsFound {
					return false, deployedVersion, errors.New(fmt.Sprintf("Unable to find %s metadata in %s file", metadataToCheck, fileToCheck))
				}
			}
			if !fileIsFound {
				return false, deployedVersion, errors.New(fmt.Sprintf("Unable to find %s file", fileToCheck))
			}
		}
	}
	return genial, deployedVersion, nil
}

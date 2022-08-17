package deploy

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func GetDeployedPackageVersion(deploymentParams Parameters, azureClientSecret *azidentity.ClientSecretCredential) (string, error) {
	defer declareNewStep("Checking current deployed version")()

	serviceClient, err := azblob.NewServiceClient(fmt.Sprintf("https://%s.%s/", *deploymentParams.StorageName, AzureBlobDomain), azureClientSecret, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create a storage account client for %s with error %v", *deploymentParams.ContainerName, err)
	}

	client, err := serviceClient.NewContainerClient(*deploymentParams.ContainerName)
	if err != nil {
		return "", fmt.Errorf("unable to get %s container on storage account %s with error: %v", *deploymentParams.ContainerName, *deploymentParams.StorageName, err)
	}

	pager := client.ListBlobsFlat(&azblob.ContainerListBlobsFlatOptions{
		Include: []azblob.ListBlobsIncludeItem{"metadata", "tags"},
	})

	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		for _, v := range resp.ListBlobsFlatSegmentResponse.Segment.BlobItems {
			if *v.Name == *deploymentParams.FileNameToCheck {
				if v.BlobTags == nil {
					return "", errors.New(fmt.Sprintf("Unable to find %s tag in %s file (Container %s Storage Account %s)", *deploymentParams.BlobTagKey, *deploymentParams.FileNameToCheck, *deploymentParams.ContainerName, *deploymentParams.StorageName))
				}

				for _, tag := range v.BlobTags.BlobTagSet {
					if *tag.Key == *deploymentParams.BlobTagKey {
						return *tag.Value, nil
					}
				}
				return "", errors.New(fmt.Sprintf("Unable to find %s tag in %s file (Container %s Storage Account %s)", *deploymentParams.BlobTagKey, *deploymentParams.FileNameToCheck, *deploymentParams.ContainerName, *deploymentParams.StorageName))
			}
		}
	}
	return "", errors.New(fmt.Sprintf("Unable to find %s file in container %s (%s)", *deploymentParams.FileNameToCheck, *deploymentParams.ContainerName, *deploymentParams.StorageName))
}

func Deploy(deploymentParameters Parameters, azureClientSecret *azidentity.ClientSecretCredential) error {

	downloadedData, err := downloadPackage(deploymentParameters, azureClientSecret)
	if err != nil {
		return err
	}

	extractedFiles, err := extractPackage(downloadedData)
	if err != nil {
		return err
	}

	err = deployPackage(deploymentParameters, extractedFiles, azureClientSecret)
	if err != nil {
		return err
	}

	return nil
}

func downloadPackage(deploymentParameters Parameters, azureClientSecret *azidentity.ClientSecretCredential) (*bytes.Buffer, error) {
	defer declareNewStep("Download package to deploy")()

	zipName := fmt.Sprintf("%s%s.zip", deploymentParameters.PackageUrl(), *deploymentParameters.VersionToDeploy)

	fmt.Printf("Trying to fetch %s\n", zipName)

	blobClient, err := azblob.NewBlockBlobClient(zipName, azureClientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s file package metadata before download with error: %v", zipName, err)
	}

	get, err := blobClient.Download(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to download %s file package with error: %v", zipName, err)
	}

	downloadedData := &bytes.Buffer{}
	options := &azblob.RetryReaderOptions{}
	reader := get.Body(options)
	_, err = downloadedData.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("unable to download %s file package with error: %v", zipName, err)
	}
	err = reader.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to download %s file package with error: %v", zipName, err)
	}

	return downloadedData, nil
}

func extractPackage(downloadedData *bytes.Buffer) (map[string]*bytes.Buffer, error) {
	defer declareNewStep("Extracting package")()

	unzippedPackage := downloadedData.Bytes()
	newReader := bytes.NewReader(unzippedPackage)
	decompressor, _ := zip.NewReader(newReader, int64(len(unzippedPackage)))

	extractedFiles := make(map[string]*bytes.Buffer)
	for _, file := range decompressor.File {
		open, err := file.Open()
		downloadedData := &bytes.Buffer{}
		_, err = downloadedData.ReadFrom(open)

		if err != nil {
			return nil, fmt.Errorf("unable to read and extract file %s file from zip package with error: %v", file.Name, err)
		}

		extractedFiles[file.Name] = downloadedData
	}

	fmt.Printf("Package extracted (%d files / %d)\n", len(extractedFiles), len(decompressor.File))
	return extractedFiles, nil
}

func deployPackage(deploymentParameters Parameters, extractedFiles map[string]*bytes.Buffer, azureClientSecret *azidentity.ClientSecretCredential) error {
	url := deploymentParameters.StorageUrl()

	defer declareNewStep("Uploading files")()

	ctx := context.Background()

	for fileName, content := range extractedFiles {
		blobClient, err := azblob.NewBlockBlobClient(fmt.Sprintf("%s%s", url, fileName), azureClientSecret, nil)
		if err != nil {
			return fmt.Errorf("unable to upload %s file in storage %s with error: %v", fileName, url, err)
		}

		_, err = blobClient.UploadBuffer(ctx, content.Bytes(), azblob.UploadOption{
			TagsMap: map[string]string{
				*deploymentParameters.BlobTagKey: *deploymentParameters.VersionToDeploy,
			},
		})
		if err != nil {
			return fmt.Errorf("unable to upload %s file in storage %s with error: %v", fileName, url, err)
		}
	}
	fmt.Printf("Package deployed with success to %s\n (%d files)", url, len(extractedFiles))
	return nil
}

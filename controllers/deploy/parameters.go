package deploy

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

const AzureBlobDomain string = "blob.core.windows.net"

type Parameters struct {
	*AzureCredential
	StorageName     *string
	ContainerName   *string
	FileNameToCheck *string
	BlobTagKey      *string
	VersionToDeploy *string
	Package         *Package
}

type AzureCredential struct {
	TenantId  *string
	SpnId     *string
	SpnSecret *string
}

type Package struct {
	StorageName   *string
	ContainerName *string
}

func InitParameters() Parameters {
	return Parameters{
		AzureCredential: &AzureCredential{
			TenantId:  flag.String("tenantId", "", "Azure Subscription TenantId"),
			SpnId:     flag.String("spnId", "", "Azure SPN Id (Could be found here https://paas-front-end.labpaas.prd.euw.gbis.sg-azure.com/my_spn)"),
			SpnSecret: flag.String("spnSecret", "", "Azure SPN Secret (Could be found here https://paas-front-end.labpaas.prd.euw.gbis.sg-azure.com/my_spn"),
		},
		StorageName:     flag.String("storageName", "", "Azure storage account name where is located the App"),
		ContainerName:   flag.String("containerName", "$web", "Azure storage account container name where is located the file to check"),
		FileNameToCheck: flag.String("fileNameToCheck", "index.html", "The file inside the storage account we need to check app version"),
		BlobTagKey:      flag.String("blobTagKey", "version", "The blob tag key on the file where is located the version"),
		VersionToDeploy: flag.String("versionToDeploy", "", "Version to deploy"),
		Package: &Package{
			StorageName:   flag.String("packageStorageName", "", "Azure storage account name where is located the package to deploy"),
			ContainerName: flag.String("packageContainerName", "packages", "Azure storage account container name where is located the package to deploy"),
		},
	}
}

func (parameters Parameters) String() string {
	var builder strings.Builder

	PrintHeaderToConsole("Script parameters")
	builder.WriteString(fmt.Sprintf("TenantId: %s \n", *parameters.TenantId))
	builder.WriteString(fmt.Sprintf("SpnId: %s \n", *parameters.SpnId))
	builder.WriteString(fmt.Sprintf("BlobTagKey: %s \n", *parameters.BlobTagKey))
	builder.WriteString(fmt.Sprintf("FileNameToCheck: %v \n", *parameters.FileNameToCheck))
	builder.WriteString(fmt.Sprintf("ContainerName: %s \n", *parameters.ContainerName))
	builder.WriteString(fmt.Sprintf("SpnSecret: %s \n", Obfuscate(*parameters.SpnSecret)))
	builder.WriteString(fmt.Sprintf("StorageName: %s \n", *parameters.StorageName))
	builder.WriteString(fmt.Sprintf("PackageStorageName: %s \n", *parameters.Package.StorageName))
	builder.WriteString(fmt.Sprintf("PackageContainerName: %s \n", *parameters.Package.ContainerName))
	builder.WriteString(fmt.Sprintf("versionToDeploy: %s \n", *parameters.VersionToDeploy))
	return builder.String()
}

func (parameters Parameters) PackageUrl() string {
	return fmt.Sprintf("https://%s.%s/%s/", *parameters.Package.StorageName, AzureBlobDomain, *parameters.Package.ContainerName)
}

func (parameters Parameters) StorageUrl() string {
	return fmt.Sprintf("https://%s.%s/%s/", *parameters.StorageName, AzureBlobDomain, *parameters.ContainerName)
}

func (parameters Parameters) Validate() (bool, []string) {
	var parametersError []string
	if *parameters.TenantId == "" {
		parametersError = append(parametersError, "TenantId")
	}

	if *parameters.SpnId == "" {
		parametersError = append(parametersError, "SpnId")
	}

	if *parameters.SpnSecret == "" {
		parametersError = append(parametersError, "SpnSecret")
	}

	if *parameters.StorageName == "" {
		parametersError = append(parametersError, "StorageName")
	}

	if *parameters.ContainerName == "" {
		parametersError = append(parametersError, "ContainerName")
	}

	if *parameters.FileNameToCheck == "" {
		parametersError = append(parametersError, "FileNameToCheck")
	}

	if *parameters.BlobTagKey == "" {
		parametersError = append(parametersError, "BlobTagKey")
	}

	if *parameters.VersionToDeploy == "" {
		parametersError = append(parametersError, "VersionToDeploy")
	}

	if *parameters.Package.ContainerName == "" {
		parametersError = append(parametersError, "PackageContainerName")
	}

	if *parameters.Package.StorageName == "" {
		parametersError = append(parametersError, "PackageStorageName")
	}

	if len(parametersError) == 0 {
		return true, nil
	}
	return false, parametersError
}

func declareNewStep(stepName string) func() {
	start := time.Now()
	PrintHeaderToConsole(stepName)
	return func() {
		fmt.Printf("End %s after %s\n", stepName, time.Since(start))
	}
}

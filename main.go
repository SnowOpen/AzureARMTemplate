package main

import (
	"context"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-10-01/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	tenantID       = "[tenantID]"
	subscriptionID = "[subscriptionID]"
	clientID       = "[clientID]"
	clientSecret   = "[clientSecret]"
)

var (
	ctx        = context.Background()
	authorizer autorest.Authorizer
)

func init() {
	var err error
	authorizer, err = auth.NewClientCredentialsConfig(clientID, clientSecret, tenantID).Authorizer()

	if err != nil {
		log.Fatalf("Failed to get OAuth config: %v", err)
	}
}

func main() {
	//创建资源组
	resourceGroupName := "TestBatchVMGroup1"
	resourceGroupLocation := "eastus"
	group, err := CreateGroup(resourceGroupName, resourceGroupLocation)
	if err != nil {
		log.Fatalf("failed to create group: %v", err)
	}
	log.Printf("Created group: %v", *group.Name)

	//创建部署
	templateURL := "https://raw.githubusercontent.com/SnowOpen/AzureARMTemplate/main/BatchCreateVMs-Template.json"
	deploymentName := "deploy-" + time.Now().UTC().Format("20060102150405")
	log.Printf("Starting deployment: %s", deploymentName)
	result, err := CreateDeployment(deploymentName, resourceGroupName, &templateURL)
	if err != nil {
		log.Fatalf("Failed to deploy: %v", err)
	}
	if result.Name != nil {
		log.Printf("Completed deployment %v: %v", deploymentName, result.Properties.ProvisioningState)
	} else {
		log.Printf("Completed deployment %v (no data returned to SDK)", deploymentName)
	}
}

//CreateGroup Create the Resource Group
func CreateGroup(resourceGroupName string, resourceGroupLocation string) (group resources.Group, err error) {
	groupsClient := resources.NewGroupsClient(subscriptionID)
	groupsClient.Authorizer = authorizer

	return groupsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		resources.Group{
			Location: to.StringPtr(resourceGroupLocation)})
}

//CreateDeployment Create the deployment
func CreateDeployment(deploymentName string, resourceGroupName string, templateURL *string) (deployment resources.DeploymentExtended, err error) {

	params := make(map[string]interface{})
	params["adminPassword"] = map[string]string{
		"value": "password123!@#",
	}

	templateLink := resources.TemplateLink{URI: templateURL}
	deploymentsClient := resources.NewDeploymentsClient(subscriptionID)
	deploymentsClient.Authorizer = authorizer

	deploymentFuture, err := deploymentsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		deploymentName,
		resources.Deployment{
			Properties: &resources.DeploymentProperties{
				TemplateLink: &templateLink,
				Parameters:   params,
				Mode:         resources.Incremental,
			},
		},
	)
	if err != nil {
		return
	}
	err = deploymentFuture.WaitForCompletionRef(ctx, deploymentsClient.BaseClient.Client)
	if err != nil {
		return
	}
	return deploymentFuture.Result(deploymentsClient)
}

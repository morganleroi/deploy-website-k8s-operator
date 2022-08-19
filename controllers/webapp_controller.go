/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	webappv1alpha1 "github.com/morganleroi/deploy-website-k8s-operator/api/v1alpha1"
	"github.com/morganleroi/deploy-website-k8s-operator/controllers/deploy"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// WebappReconciler reconciles a Webapp object
type WebappReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.simpletest.com,resources=webapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.simpletest.com,resources=webapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.simpletest.com,resources=webapps/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;
func (r *WebappReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	webAppCrd := &webappv1alpha1.Webapp{}
	err := r.Get(ctx, req.NamespacedName, webAppCrd)
	log.Log.Info("---------------------------")
	log.Log.Info("Request name", "WebappVersion", req.Name)

	deploymentParameters := deploy.Parameters{
		AzureCredential: &deploy.AzureCredential{
			TenantId:  &webAppCrd.Spec.AzureTenantId,
			SpnId:     &webAppCrd.Spec.AzureSpnId,
			SpnSecret: &webAppCrd.Spec.AzureSpnSecret,
		},
		StorageName:     &webAppCrd.Spec.StorageName,
		ContainerName:   &webAppCrd.Spec.ContainerName,
		FileNameToCheck: &webAppCrd.Spec.FileNameToCheck,
		BlobTagKey:      &webAppCrd.Spec.BlobTagKey,
		VersionToDeploy: &webAppCrd.Spec.VersionToDeploy,
		Package: &deploy.Package{
			StorageName:   &webAppCrd.Spec.PackageStorageName,
			ContainerName: &webAppCrd.Spec.PackageContainerName,
		},
	}

	fmt.Println(deploymentParameters)

	err = deploy.StartDeployment(deploymentParameters)

	var condition v1.Condition

	dateNow := time.Now().Format(time.Layout)
	if err != nil {
		log.Log.Info(fmt.Sprintf("Fail to reconcile (%s) %s - %s", dateNow, req.Name, err))
		webAppCrd.Status.Status = "ERROR"
		condition = v1.Condition{
			Type:    "Degraded",
			Status:  v1.ConditionFalse,
			Reason:  "DeploymentFailed",
			Message: "",
		}
	} else {
		log.Log.Info(fmt.Sprintf("Reconcile is ok (%s) %s", dateNow, req.Name))
		webAppCrd.Status.Status = "SUCCESS"
		webAppCrd.Status.DeployedVersion = webAppCrd.Spec.VersionToDeploy
		condition = v1.Condition{
			Type:    "Available",
			Status:  v1.ConditionTrue,
			Reason:  "Deployed",
			Message: "",
		}
	}

	meta.SetStatusCondition(&webAppCrd.Status.Conditions, condition)
	errStatusUpdate := r.Status().Update(ctx, webAppCrd)

	if errStatusUpdate != nil {
		return ctrl.Result{}, err
	}

	if err != nil {
		log.Log.Info("Ending reconciliation")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *WebappReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1alpha1.Webapp{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}

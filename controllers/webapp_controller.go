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
	"github.com/morganleroi/AzBlobStorage/deploy"
	webappv1alpha1 "github.com/morganleroi/deploy-website-k8s-operator/api/v1alpha1"
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
	log.Log.Info("Parameter", "AzureSubscriptionId", webAppCrd.Spec.AzureSubscriptionId)
	log.Log.Info("Parameter", "StorageAccountName", webAppCrd.Spec.StorageAccountName)
	log.Log.Info("Parameter", "WebappVersion", webAppCrd.Spec.WebappVersion)

	deployedPackage, err := deploy.GetDeployedPackage(webAppCrd.Spec.StorageAccountName, "XXX")
	dateNow := time.Now().Format(time.Layout)
	if err != nil {
		log.Log.Info(fmt.Sprintf("Fail to reconcile (%s) %s - %s", dateNow, req.Name, err))
		webAppCrd.Status.Status = "ERROR"
		webAppCrd.Status.LastUpdate = dateNow
		webAppCrd.Status.Error = fmt.Sprintf("Error happened: %s", err)
	} else {
		log.Log.Info(fmt.Sprintf("Reconcile is ok (%s) %s", dateNow, req.Name))
		webAppCrd.Status.Status = "SUCCESS"
		webAppCrd.Status.LastUpdate = dateNow
		webAppCrd.Status.DeployedVersion = deployedPackage
		webAppCrd.Status.Error = ""
	}

	err = r.Status().Update(ctx, webAppCrd)
	if err != nil {
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

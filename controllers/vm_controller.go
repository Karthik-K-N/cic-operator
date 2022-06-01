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
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudv1 "github.com/Karthik-K-N/cic-operator/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VMReconciler reconciles a VM object
type VMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cloud.ibm.com,resources=vms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud.ibm.com,resources=vms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud.ibm.com,resources=vms/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *VMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var vm cloudv1.VM
	if err := r.Get(ctx, req.NamespacedName, &vm); err != nil {
		klog.Error(err, "unable to fetch vm")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	klog.Infof("Found the vm object %v", vm)

	// Deletion logic
	if !vm.ObjectMeta.DeletionTimestamp.IsZero() {
		vm.Status.Status = "Deleting"
		klog.Info("Updating the status of vm %s to Deleting", vm.Name)

		// TODO: Implement the delete logic here

		if err := r.Status().Update(ctx, &vm); err != nil {
			klog.Error(err, "Unable to update vm status")
			return ctrl.Result{}, err
		}

		klog.Infof("%v: machine deletion successful", vm.Name)
		return reconcile.Result{}, nil
	}

	klog.Infof("Creating VM in CIC with name %s", vm.Name)
	// TODO: Add the CIC logic here

	// TODO: Get the actual status and update the status
	vm.Status.Status = "Running"

	klog.Info("Updating the status of vm %s", vm.Name)
	if err := r.Status().Update(ctx, &vm); err != nil {
		klog.Error(err, "Unable to update vm status")
		return ctrl.Result{}, err
	}

	klog.Infof("Successfully updated the vm %s status ", vm.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudv1.VM{}).
		Complete(r)
}

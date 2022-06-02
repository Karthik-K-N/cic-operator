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
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/utils/openstack/clientconfig"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cloudv1 "github.com/Karthik-K-N/cic-operator/api/v1"
)

const (
	serviceFinalizer = "cic.cloud.ibm.com"
	requeueAfter     = 30 * time.Second
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
	vm := &cloudv1.VM{}
	if err := r.Get(ctx, req.NamespacedName, vm); err != nil {
		klog.Error(err, "unable to fetch vm")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	klog.Infof("Found the VM object with name %s", vm.Name)

	// Delete if necessary
	if vm.ObjectMeta.DeletionTimestamp.IsZero() {
		// Instance is not being deleted, add the finalizer if not present
		if !containsServiceFinalizer(vm) {
			vm.ObjectMeta.Finalizers = append(vm.ObjectMeta.Finalizers, serviceFinalizer)
			if err := r.Update(ctx, vm); err != nil {
				klog.Error(err, "Error adding finalizer", "service", vm.Name)
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsServiceFinalizer(vm) {
			klog.Infof("Deleting the VM %s", vm.Name)
			err := r.DeleteVM(vm)
			if err != nil {
				klog.Error(err, "Delete VM %s failed with error %v", vm.Name, err.Error())
				vm.Status.Status = "Delete Failed"
				klog.Info("Updating the status of vm %s to Delete Failed", vm.Name)
				if err := r.Status().Update(ctx, vm); err != nil {
					klog.Error(err, "Unable to update vm status")
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			vm.ObjectMeta.Finalizers = deleteServiceFinalizer(vm)
			err = r.Update(ctx, vm)
			if err != nil {
				klog.Error(err, "Error removing finalizers")
			}
			klog.Infof("Successfully deleted the VM %s", vm.Name)
			return ctrl.Result{}, err
		}
	}
	klog.Infof("Checking whether VM exists in CIC with name %s", vm.Name)
	server, err := r.CheckVMExists(vm)
	if err != nil {
		klog.Errorf("Failed to check VM %s exists in CIC error %v", vm.Name, err.Error())
		return ctrl.Result{}, err
	}
	if server != nil {
		klog.Infof("VM with name %s already exists in CIC not creating it again", vm.Name)
		// update the status if not equal
		if server.Status != vm.Status.Status {
			vm.Status.Status = server.Status
			if err := r.Status().Update(ctx, vm); err != nil {
				klog.Error("Unable to update vm %s status", vm.Name)
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}
	klog.Infof("Creating VM in CIC with name %s", vm.Name)

	server, err = r.CreateVM(vm)
	if err != nil {
		klog.Errorf("Create VM %s failed with error %v", vm.Name, err.Error())
		return ctrl.Result{}, err
	}
	vm.Status.ID = server.ID
	vm.Status.Status = server.Status
	klog.Infof("Updating the status of vm %s", vm.Name)
	if err := r.Status().Update(ctx, vm); err != nil {
		klog.Error("Unable to update vm %s status", vm.Name)
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

func (r *VMReconciler) CreateVM(vm *cloudv1.VM) (*servers.Server, error) {
	cicClient, err := getClient()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	server, err := servers.Create(cicClient, servers.CreateOpts{
		Name:      vm.Name,
		ImageRef:  vm.Spec.ImageID,
		FlavorRef: vm.Spec.FlavorID,
		Networks: []servers.Network{
			{
				UUID: vm.Spec.NetworkID,
			},
		},
	}).Extract()
	if err != nil {
		klog.Error(err.Error())
		return nil, err
	}
	return server, nil
}

func (r *VMReconciler) DeleteVM(vm *cloudv1.VM) error {
	cicClient, err := getClient()
	if err != nil {
		return err
	}
	id := vm.Status.ID
	if id == "" {
		return fmt.Errorf("Cannot proceed with delete as VM ID is empty ")
	}
	err = servers.Delete(cicClient, vm.Status.ID).ExtractErr()
	if err != nil {
		return err
	}
	return nil
}

func (r *VMReconciler) CheckVMExists(vm *cloudv1.VM) (*servers.Server, error) {
	cicClient, err := getClient()
	if err != nil {
		return nil, err
	}
	opts := servers.ListOpts{Name: vm.Name}

	pager, err := servers.List(cicClient, opts).AllPages()
	if err != nil {
		return nil, err
	}
	allServers, err := servers.ExtractServers(pager)
	if err != nil {
		return nil, err
	}
	for _, server := range allServers {
		if server.Name == vm.Name {
			klog.Infof("Found server with name %s with ID %s Status %s", vm.Name, server.ID, server.Status)
			return &server, nil
		}
	}
	return nil, nil
}

func getClient() (*gophercloud.ServiceClient, error) {
	options := &clientconfig.ClientOpts{}
	cicClient, err := clientconfig.NewServiceClient("compute", options)
	if err != nil {
		return nil, err
	}
	return cicClient, nil
}

// containsServiceFinalizer checks if the instance contains service finalizer
func containsServiceFinalizer(instance *cloudv1.VM) bool {
	for _, finalizer := range instance.ObjectMeta.Finalizers {
		if strings.Contains(finalizer, serviceFinalizer) {
			return true
		}
	}
	return false
}

// deleteServiceFinalizer delete service finalizer
func deleteServiceFinalizer(instance *cloudv1.VM) []string {
	var result []string
	for _, finalizer := range instance.ObjectMeta.Finalizers {
		if finalizer == serviceFinalizer {
			continue
		}
		result = append(result, finalizer)
	}
	return result
}

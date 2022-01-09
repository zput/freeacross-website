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
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kubeAPIErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"math/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	batchv1 "github.com/freeacross/freeacross-website/api/v1"
)

// WebsiteReconciler reconciles a Website object
type WebsiteReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=batch.freecross.com,resources=websites,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.freecross.com,resources=websites/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.freecross.com,resources=websites/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Website object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *WebsiteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	logger := r.Log.WithValues("Website", req.NamespacedName)
	web := &batchv1.Website{}
	err := r.Client.Get(ctx, req.NamespacedName, web)
	if err != nil {
		logger.Error(err, "get website failed")
		return ctrl.Result{Requeue: false}, nil
	}

	// 1. 获取 web name 对应的所有的 deployment/services 的列表
	lbls := client.MatchingLabels{"app": web.Name}

	providerDeploymentList := &appsv1.DeploymentList{}
	err = r.List(ctx, providerDeploymentList, lbls)
	if err != nil && !kubeAPIErrors.IsNotFound(err) {
		logger.Error(err, "get provider deployment list failed")
		return ctrl.Result{}, nil
	}

	providerServiceList := &corev1.ServiceList{}
	err = r.List(ctx, providerServiceList, lbls)
	if err != nil && !kubeAPIErrors.IsNotFound(err) {
		logger.Error(err, "get provider service list failed")
		return ctrl.Result{}, nil
	}
	websiteDeployment, err := r.PullWebsiteDeployment(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, "get website failed")
		return ctrl.Result{Requeue: false}, nil
	}

	expectDeploy := newDeploymentForCR(web, websiteDeployment)

	if len(providerDeploymentList.Items) <= 0 {
		logger.V(1).Info("creating deployment")
		err := r.Client.Create(ctx, expectDeploy)
		if err != nil && !kubeAPIErrors.IsAlreadyExists(err) {
			logger.Error(err, "create deployment failed")
			return ctrl.Result{Requeue: false}, nil
		}
		// TODO
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *WebsiteReconciler) PullWebsiteDeployment(ctx context.Context, namespaceName types.NamespacedName) (*batchv1.WebsiteDeployment, error) {
	logger := r.Log.WithValues("WebsiteDeployment", namespaceName)
	web := &batchv1.WebsiteDeployment{}
	err := r.Client.Get(ctx, namespaceName, web)
	if err != nil {
		logger.Error(err, "get websiteDeployment failed")
		return nil, err
	}
	return web, nil
}

func newDeploymentForCR(cr *batchv1.Website, wd *batchv1.WebsiteDeployment) *appsv1.Deployment {
	var tmpLabels = map[string]string{"app": cr.Name}

	var tmpPod = newPodsForCR(cr, wd)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cr.Name + "-deploy",
			Namespace:    cr.Namespace,
			Labels:       tmpLabels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, batchv1.GroupVersion.WithKind(cr.Kind))},
		},

		Spec: appsv1.DeploymentSpec{
			Replicas: Int32Ptr(1),
			Selector: &metav1.LabelSelector{MatchLabels: tmpLabels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: tmpPod.ObjectMeta,
				Spec:       tmpPod.Spec,
			}}}
}

func Int32Ptr(o int32) *int32 {
	var ret = o
	return &ret
}

func newPodsForCR(wb *batchv1.Website, crs *batchv1.WebsiteDeployment) *corev1.Pod {
	var (
		name = wb.Name

		namespaceName = wb.Namespace

		containers []corev1.Container

		labels = map[string]string{"app": name}
	)

	for _, cr := range crs.Spec.Containers {
		var ports = make([]corev1.ContainerPort, 0)
		for _, e := range cr.Ports {
			ports = append(ports, corev1.ContainerPort{
				ContainerPort: e.Port,
				Protocol:      corev1.Protocol(e.Protocol),
			})

			containers = append(containers, corev1.Container{
				Name:         cr.Name,
				Image:        cr.Image,
				Ports:        ports,
				VolumeMounts: cr.VolumeMounts,
				Env:          cr.Env,
			})
		}
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-pod",
			Namespace:    namespaceName,
			Labels:       labels,
		},
		Spec: corev1.PodSpec{
			Containers: containers,
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebsiteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Website{}).
		Complete(r)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

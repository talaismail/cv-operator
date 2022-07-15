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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"bytes"
	profilev1alpha1 "github.com/talaismail/cv-operator/api/v1alpha1"
	"html/template"
)

// CurriculumVitaeReconciler reconciles a CurriculumVitae object
type CurriculumVitaeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=profile.example.com,resources=curriculumvitaes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=profile.example.com,resources=curriculumvitaes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=profile.example.com,resources=curriculumvitaes/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CurriculumVitae object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *CurriculumVitaeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	profile := &profilev1alpha1.CurriculumVitae{}
	err := r.Get(ctx, req.NamespacedName, profile)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	isMarkedToBeDeleted := profile.GetDeletionTimestamp() != nil
	if isMarkedToBeDeleted {
		return ctrl.Result{}, nil
	}

	t := template.New("index")
	t, err = template.ParseFiles("assets/index.html")
	if err != nil {
		return ctrl.Result{}, nil
	}
	var template bytes.Buffer
	err = t.Execute(&template, profile.Spec)
	if err != nil {
		return ctrl.Result{}, nil
	}
	index := template.String()

	newConfigMap := r.createConfigMap(profile, index)
	err = r.Create(ctx, newConfigMap)
	if err != nil {
		return ctrl.Result{}, err
	}

	deployment := r.createDeployment(profile, newConfigMap)
	err = r.Create(ctx, deployment)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, err
}

func (r *CurriculumVitaeReconciler) createConfigMap(curriculumVitae *profilev1alpha1.CurriculumVitae, index string) *corev1.ConfigMap {
	data := map[string]string{
		"index.html": index,
	}
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      curriculumVitae.Name,
			Namespace: curriculumVitae.Namespace,
		},
		Data: data,
	}

	ownerRef := &metav1.OwnerReference{
		APIVersion: curriculumVitae.APIVersion,
		Kind:       curriculumVitae.Kind,
		Name:       curriculumVitae.Name,
		UID:        curriculumVitae.UID,
	}
	ownerRefs := []metav1.OwnerReference{*ownerRef}
	configmap.SetOwnerReferences(ownerRefs)

	return configmap
}

func (r *CurriculumVitaeReconciler) createDeployment(curriculumVitae *profilev1alpha1.CurriculumVitae, configmap *corev1.ConfigMap) *appsv1.Deployment {
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      curriculumVitae.Name,
			Namespace: curriculumVitae.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "cv-server"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "cv-server"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "quay.io/centos7/httpd-24-centos7",
						Name:  "webserver",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "http",
							Protocol:      "TCP",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      configmap.Name,
							MountPath: "/var/html/www/",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: configmap.Name,
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: configmap.Name,
								},
							},
						},
					}},
				},
			},
		},
	}

	ownerRef := &metav1.OwnerReference{
		APIVersion: curriculumVitae.APIVersion,
		Kind:       curriculumVitae.Kind,
		Name:       curriculumVitae.Name,
		UID:        curriculumVitae.UID,
	}
	ownerRefs := []metav1.OwnerReference{*ownerRef}
	deployment.SetOwnerReferences(ownerRefs)

	return deployment
}

// SetupWithManager sets up the controller with the Manager.
func (r *CurriculumVitaeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&profilev1alpha1.CurriculumVitae{}).
		Complete(r)
}

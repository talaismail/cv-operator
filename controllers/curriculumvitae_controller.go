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

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"bytes"
	"html/template"

	profilev1alpha1 "github.com/talaismail/cv-operator/api/v1alpha1"
)

// CurriculumVitaeReconciler reconciles a CurriculumVitae object
type CurriculumVitaeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=profile.example.com,resources=curriculumvitae,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=profile.example.com,resources=curriculumvitae/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=profile.example.com,resources=curriculumvitae/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete

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

	var err error
	profile := &profilev1alpha1.CurriculumVitae{}
	err = r.Get(ctx, req.NamespacedName, profile)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	isMarkedToBeDeleted := profile.GetDeletionTimestamp() != nil
	if isMarkedToBeDeleted {
		log.Log.Info("Stop reconciling since resource is marked to be deleted.")
		return ctrl.Result{}, nil
	}

	oldRoute := &routev1.Route{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: profile.Name, Namespace: profile.Namespace}, oldRoute)
	if err == nil {
		log.Log.Info("Delete Route as it is outdated.")
		err = r.Delete(context.TODO(), oldRoute)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	oldService := &corev1.Service{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: profile.Name, Namespace: profile.Namespace}, oldService)
	if err == nil {
		log.Log.Info("Delete Service as it is outdated.")
		err = r.Delete(context.TODO(), oldService)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	oldDeployment := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: profile.Name, Namespace: profile.Namespace}, oldDeployment)
	if err == nil {
		log.Log.Info("Delete Deployment as it is outdated.")
		err = r.Delete(context.TODO(), oldDeployment)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	oldIndexConfigMap := &corev1.ConfigMap{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: profile.Name, Namespace: profile.Namespace}, oldIndexConfigMap)
	if err == nil {
		log.Log.Info("Delete ConfigMap as it is outdated.")
		err = r.Delete(context.TODO(), oldIndexConfigMap)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	t := template.New("index")
	t, err = template.ParseFiles("assets/index.html")
	if err != nil {
		log.Log.Info("Stop reconciling since the could not be parsed.")
		return ctrl.Result{}, nil
	}
	var template bytes.Buffer
	err = t.Execute(&template, profile.Spec)
	if err != nil {
		log.Log.Info("Stop reconciling since the template could not be instanciated.")
		return ctrl.Result{}, nil
	}

	index := template.String()

	indexConfigMap := r.createConfigMap(profile, "index.html", index)
	log.Log.Info("Create configmap.")
	err = r.Create(ctx, indexConfigMap)
	if err != nil {
		log.Log.Info("Requeue since there was an error while creating the ConfigMap.")
		return ctrl.Result{}, err
	}

	deployment := r.createDeployment(profile, indexConfigMap)
	log.Log.Info("Create deployment.")
	err = r.Create(ctx, deployment)
	if err != nil {
		log.Log.Info("Requeue since there was an error while creating the Deployment.")
		return ctrl.Result{}, err
	}

	service := r.createService(profile)
	log.Log.Info("Create service.")
	err = r.Create(ctx, service)
	if err != nil {
		log.Log.Info("Requeue since there was an error while creating the Service.")
		return ctrl.Result{}, err
	}

	route := r.createRoute(profile)
	log.Log.Info("Create route.")
	err = r.Create(ctx, route)
	if err != nil {
		log.Log.Info("Requeue since there was an error while creating the Route.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CurriculumVitaeReconciler) createConfigMap(curriculumVitae *profilev1alpha1.CurriculumVitae, key string, value string) *corev1.ConfigMap {
	data := map[string]string{
		key: value,
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

func (r *CurriculumVitaeReconciler) createDeployment(curriculumVitae *profilev1alpha1.CurriculumVitae, indexConfigMap *corev1.ConfigMap) *appsv1.Deployment {
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      curriculumVitae.Name,
			Namespace: curriculumVitae.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": curriculumVitae.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": curriculumVitae.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "quay.io/centos7/httpd-24-centos7",
						Name:  "webserver",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 8080,
								Name:          "http",
								Protocol:      "TCP",
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      indexConfigMap.Name,
								MountPath: "/opt/rh/httpd24/root/var/www/html/",
							},
						},
					}},
					Volumes: []corev1.Volume{
						{
							Name: indexConfigMap.Name,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: indexConfigMap.Name,
									},
								},
							},
						},
					},
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

func (r *CurriculumVitaeReconciler) createService(curriculumVitae *profilev1alpha1.CurriculumVitae) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      curriculumVitae.Name,
			Namespace: curriculumVitae.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": curriculumVitae.Name},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
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
	svc.SetOwnerReferences(ownerRefs)

	return svc
}

func (r *CurriculumVitaeReconciler) createRoute(curriculumVitae *profilev1alpha1.CurriculumVitae) *routev1.Route {
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      curriculumVitae.Name,
			Namespace: curriculumVitae.Namespace,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: curriculumVitae.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(8080),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
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
	route.SetOwnerReferences(ownerRefs)

	return route
}

// SetupWithManager sets up the controller with the Manager.
func (r *CurriculumVitaeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&profilev1alpha1.CurriculumVitae{}).
		Complete(r)
}

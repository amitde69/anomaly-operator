/*
Copyright 2023.

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
	"time"
	monitoringv1alpha1 "github.com/amitde69/anomaly-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	// "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	// "k8s.io/apimachinery/pkg/labels"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"gopkg.in/yaml.v3"
	"strconv"

)

// DetectorReconciler reconciles a Detector object
type DetectorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	finalizers []string = []string{"finalizers.detectors.monitoring.amitdebachar"}
)

//+kubebuilder:rbac:groups=monitoring.amitdebachar,resources=detectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.amitdebachar,resources=detectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitoring.amitdebachar,resources=detectors/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete


// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Detector object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *DetectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	logger := log.FromContext(ctx)
	logger.Info("Reconciling Objects")

	detector := &monitoringv1alpha1.Detector{}

	err := r.Get(ctx, req.NamespacedName, detector)
	if err != nil {
		// if the resource is not found, then just return (might look useless as this usually happens in case of Delete events)
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error occurred while fetching the detector resource")
		return ctrl.Result{}, err
	}

	status := monitoringv1alpha1.DetectorStatus{
		IsCreated: true,
	}

	if !reflect.DeepEqual(detector.Status, status) {
		detector.Status = status
		err := r.Client.Status().Update(ctx, detector)
		if err != nil {
			logger.Error(err, "Error occurred while updating the detector resource")
			return reconcile.Result{}, err
		}
	}
	
	if detector.GetDeletionTimestamp().IsZero() {

		serviceaccount := r.newServiceAccount(detector)
		serviceaccount_exist := serviceaccount.DeepCopy()
		err = r.Get(ctx, req.NamespacedName, serviceaccount_exist)
		if err != nil {
			if errors.IsNotFound(err) {
				// create a new serviceaccount
				logger.Info("Creating new serviceaccount following new detector resource")
				if err = r.Create(ctx, serviceaccount); err != nil {
					return ctrl.Result{}, err
				}
				return ctrl.Result{Requeue: true}, nil
			}
		}

		deployment := r.newDeployment(detector)
		deployment_exist := deployment.DeepCopy()
		
		err := r.Get(ctx, req.NamespacedName, deployment_exist)
		if err != nil {
			if errors.IsNotFound(err) {
				// create a new deployment
				logger.Info("Creating new deployment following new detector resource")
				if err = r.Create(ctx, deployment); err != nil {
					return ctrl.Result{}, err
				}
				return ctrl.Result{Requeue: true}, nil
			}
		} else {
			deployment.Spec.Template.Annotations = deployment_exist.Spec.Template.Annotations 
			if !reflect.DeepEqual(deployment.Spec , deployment_exist.Spec) {
				// Update the Deployment
				deployment_exist.Spec.Template = deployment.Spec.Template
				err = r.Update(ctx, deployment_exist)
				if err != nil {
					return ctrl.Result{}, err
				}
			}
		}
		size := int32(1)
		if *deployment_exist.Spec.Replicas != size {
			deployment_exist.Spec.Replicas = &size
			if err = r.Update(ctx, deployment_exist); err != nil {
				return ctrl.Result{}, err
			}
		}

		if err := r.createFinalizerCallback(ctx, detector); err != nil {
			logger.Error(err, "error occurred while dealing with the creation of finalizer")
		}

		configmap := r.newConfigMap(detector)
		configmap_exist := configmap.DeepCopy()
		err = r.Get(ctx, req.NamespacedName, configmap_exist)
		if err != nil {
			// fmt.Printf("found existing config map %w", err)
			if errors.IsNotFound(err) {
				// create a new configmap
				logger.Info("Creating new configmap following new detector resource")
				if err = r.Create(ctx, configmap); err != nil {
					return ctrl.Result{}, err
				}
				return ctrl.Result{Requeue: true}, nil
			}
		} else {
			if !reflect.DeepEqual(configmap.Data, configmap_exist.Data) {
				// Update the Configmap
				configmap_exist.Data = configmap.Data
				err = r.Update(ctx, configmap_exist)
				if err != nil {
					return ctrl.Result{}, err
				}
				
				deployment_exist.Spec.Template = deployment.Spec.Template
				// Set the annotation to trigger a rollout
				annotations := deployment_exist.Spec.Template.GetAnnotations()
				if annotations == nil {
					annotations = make(map[string]string)
				}
				logger.Info("Detected update in detector config rolling out deployment")
				annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().String()
				deployment_exist.Spec.Template.SetAnnotations(annotations)
				err = r.Update(ctx, deployment_exist)
				if err != nil {
					return ctrl.Result{}, err
				}
			}
		}
		

	} 
	
	isDetectorMarkedToBeDeleted := detector.GetDeletionTimestamp() != nil
	if isDetectorMarkedToBeDeleted {
		logger.Info("Deletion detected! Proceeding to cleanup...")
		
		if err := r.cleanupFinalizerCallback(ctx, detector); err != nil {
			logger.Error(err, "error occurred while dealing with the cleanup finalizer")
		}
		logger.Info("cleaned up the finalizer successfully")
	}
	return ctrl.Result{Requeue: true}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DetectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1alpha1.Detector{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func (r *DetectorReconciler) cleanupFinalizerCallback(ctx context.Context, detector *monitoringv1alpha1.Detector) error {
	// parse the table and the id of the row to delete

	// remove the cleanup-row finalizer from the postgresWriterObject
	for _, finalizer := range finalizers {
		if controllerutil.ContainsFinalizer(detector, finalizer) {
			controllerutil.RemoveFinalizer(detector, finalizer)
			if err := r.Update(ctx, detector); err != nil {
				return fmt.Errorf("error occurred while removing the finalizer: %w", err)
			}
		}
	}
	return nil
}
func (r *DetectorReconciler) createFinalizerCallback(ctx context.Context, detector *monitoringv1alpha1.Detector) error {
	// parse the table and the id of the row to delete

	// remove the cleanup-row finalizer from the postgresWriterObject
	for _, finalizer := range finalizers {
		if controllerutil.ContainsFinalizer(detector, finalizer) {
			// fmt.Printf("Detected a new resource, creating a finalizer for it")
			detector.SetFinalizers(finalizers)
			if err := r.Update(ctx, detector); err != nil {
				return fmt.Errorf("error occurred while setting the finalizers of the detector resource: %w", err)
			}
		}
	}
	// logger.Info("created the finalizer successfully")
	return nil
}

func (r *DetectorReconciler) newDeployment(cr *monitoringv1alpha1.Detector) *appsv1.Deployment {
	replicas := int32(1)
	revisions := int32(2)
	labels :=  map[string]string{
		"app": cr.Name,
	}
	podTemplateSpec := corev1.PodTemplateSpec{}
	// check if PodSpec in config is empty
	if !reflect.DeepEqual(cr.Spec.PodSpec, podTemplateSpec) {
		podTemplateSpec = cr.Spec.PodSpec
	} else {
		podTemplateSpec = corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: labels,
						},
						Spec: corev1.PodSpec{
							Volumes: []corev1.Volume{
								{
									Name: cr.Name,
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cr.Name,
											},
										},
									},
								},
							},
							ServiceAccountName: cr.Name,
							Containers: []corev1.Container{{
									Image: cr.Spec.Image,
									Name: cr.Name,
									Env: []corev1.EnvVar{
										{
											Name: "LOG_LEVEL",
											Value: "INFO",
										},
									},
									Ports: []corev1.ContainerPort{{
										ContainerPort: 9090,
										Name: "http",
									}},
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											"cpu":    resourceQuantity("1"),
											"memory": resourceQuantity("500M"),
										},
										Limits: corev1.ResourceList{
											"cpu":    resourceQuantity("1"),
											"memory": resourceQuantity("500M"),
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      cr.Name,
											MountPath: "/app/config.yaml",
											SubPath:   cr.Name + "-conf.yaml",
										},
									},
								},
							},
					},
				}
			// cr.Spec.PodSpec = podTemplateSpec
	}
	deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: cr.Name,
				Namespace: cr.Namespace,
				Labels: labels,
			},

			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{
							MatchLabels: labels,
						},
				RevisionHistoryLimit: &revisions,
				Template: podTemplateSpec,
			},
	}
	deployment.Spec.Template.ObjectMeta.Labels = labels

	controllerutil.SetControllerReference(cr, deployment, r.Scheme)
	return deployment
  }


func (r *DetectorReconciler) newServiceAccount(cr *monitoringv1alpha1.Detector) *corev1.ServiceAccount {
	serviceaccount := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
	}
	controllerutil.SetControllerReference(cr, serviceaccount, r.Scheme)
	return serviceaccount
}

func (r *DetectorReconciler) newConfigMap(cr *monitoringv1alpha1.Detector) *corev1.ConfigMap {
	intervalMins := cr.Spec.IntervalMins
	querySpec := r.convertQuerySpectoQuery(cr)
	configData := monitoringv1alpha1.Config{
		IntervalMins: intervalMins,
		PromURL: cr.Spec.PromUrl,
		Queries: querySpec,
	}
	
	// Convert the configData struct to YAML format
	configDataYaml, err := yaml.Marshal(configData)
	if err != nil {
		fmt.Errorf("Cant parse config yaml: %w.", err)
	}
    configMap := &corev1.ConfigMap{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "v1",
            Kind:       "ConfigMap",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name:      cr.Name,
            Namespace: cr.Namespace,
        },
		Data: map[string]string{
			cr.Name + "-conf.yaml": string(configDataYaml),
		},
    }
    controllerutil.SetControllerReference(cr, configMap, r.Scheme)
    return configMap
}

func (r *DetectorReconciler) convertQuerySpectoQuery(cr *monitoringv1alpha1.Detector) []monitoringv1alpha1.Query {
	querySpec := []monitoringv1alpha1.Query{}
	for _, query := range cr.Spec.Queries {
		newQuery := monitoringv1alpha1.Query{}
		if query.Flexibility != "" {
			var err error
			newQuery.Flexibility, err = strconv.ParseFloat(query.Flexibility, 4)
			if err != nil {
				fmt.Printf("Error while parsing QuerySpec")
			}
		}
		if query.Buffer_Pct != 0 {
			newQuery.Buffer_Pct = GetIntPointer(int(query.Buffer_Pct))
		}
		if query.Resolution != 0 {
			newQuery.Resolution = int(query.Resolution)
		}
		if query.Detection_Window_Hours != 0 {
			newQuery.Detection_Window_Hours = int64(query.Detection_Window_Hours)
		}
		newQuery.Name = query.Name
		newQuery.Train_Window = query.Train_Window
		newQuery.Query = query.Query
		querySpec = append(querySpec, newQuery)
	}
	return querySpec
}

func resourceQuantity(s string) resource.Quantity {
    q, err := resource.ParseQuantity(s)
    if err != nil {
        panic(err)
    }
    return q
}


func GetIntPointer(value int) *int {
    return &value
}
/*
Copyright 2021.

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
	"errors"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	log2 "log"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1 "ash.lt/secret-creator/api/v1"
)

// SecretCreatorReconciler reconciles a SecretCreator object
type SecretCreatorReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Max time after creation of namespace to create secrets on it
const MAX_OLDNESS_IN_SECONDS = 300

//+kubebuilder:rbac:groups=api.ash.lt,resources=secretcreators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.ash.lt,resources=secretcreators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.ash.lt,resources=secretcreators/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecretCreator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *SecretCreatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	r.Log.WithValues("reconcile: ", r, "\n", req.NamespacedName)

	timenow := time.Now().Unix()

	is_namespace := false
	is_secretcreator := false

	//log2.Println("test", r.Scheme.Name(), req, ctx)

	var sc apiv1.SecretCreator
	var ns corev1.Namespace

	if err := r.Get(ctx, req.NamespacedName, &sc); err != nil {
		if err := r.Get(ctx, req.NamespacedName, &ns); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		} else {
			is_namespace = true
		}
	} else {
		is_secretcreator = true
	}

	if is_namespace {
		if ns.Status.Phase == corev1.NamespaceActive {
			if timenow-ns.CreationTimestamp.Unix() > MAX_OLDNESS_IN_SECONDS {
				log2.Println(ns.Name + " will not check, too old")
				return ctrl.Result{}, nil
			}

			log2.Println("Will be check secrets in namespace " + ns.Name)

			var secretCreators apiv1.SecretCreatorList
			if err := r.Client.List(ctx, &secretCreators); err != nil {
				return ctrl.Result{}, errors.New("Can not get list of secretcreators: " + err.Error())
			} else {
				for _, secretCreator := range secretCreators.Items {
					nsname := types.NamespacedName{
						Namespace: ns.Name,
						Name:      secretCreator.Spec.SecretName,
					}

					err := createSecretIfNotExists(ctx, r.Client, secretCreator.Spec.Secret, nsname)
					if err != nil {
						continue
					}
					r.Recorder.Event(&secretCreator, corev1.EventTypeNormal, "", "Secret "+nsname.Name+" for namespace "+nsname.Namespace+" created")

				}
			}

		}
	} else if is_secretcreator {
		nss := &corev1.NamespaceList{}
		err := r.Client.List(ctx, nss)
		if err != nil {
			return ctrl.Result{}, errors.New("Can not get list of namespaces: " + err.Error())
		}
		for _, ns := range nss.Items {

			if timenow-ns.CreationTimestamp.Unix() > MAX_OLDNESS_IN_SECONDS {
				log2.Println(ns.Name + " will not check, too old")
				continue
			}

			err := createSecretIfNotExists(ctx, r.Client, sc.Spec.Secret, types.NamespacedName{
				Name:      sc.Spec.SecretName,
				Namespace: ns.Name,
			})

			if err != nil {
				continue
			}
			r.Recorder.Event(&sc, corev1.EventTypeNormal, "", "Secret "+ns.Name+" for namespace "+ns.Name+" created")
		}

	}

	return ctrl.Result{}, nil
}

func createSecretIfNotExists(ctx context.Context, cl client.Client, secret corev1.Secret, nsname types.NamespacedName) error {
	var oldsecret = &corev1.Secret{}
	err := cl.Get(ctx, nsname, oldsecret)
	if err != nil && client.IgnoreNotFound(err) != nil {
		if oldsecret.Type == secret.Type && reflect.DeepEqual(oldsecret.Data, secret.Data) {
			log2.Println(nsname, " already exists, all ok")
			return nil
		} else {
			log2.Println(nsname, " exists but not same, be careful")
		}
	}

	secret.ObjectMeta = metav1.ObjectMeta{
		Name:      nsname.Name,
		Namespace: nsname.Namespace,
	}
	return cl.Create(ctx, &secret)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretCreatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.SecretCreator{}).
		Watches(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForObject{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}

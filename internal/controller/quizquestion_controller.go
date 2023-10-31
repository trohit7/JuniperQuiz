package controller

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	//"sigs.k8s.io/controller-runtime/pkg/reconcile"

	juniperv1 "github.com/juniper/quiz-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
)

// QuizQuestionReconciler reconciles a QuizQuestion object
type QuizQuestionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=juniper.juniper.com,resources=quizquestions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=juniper.juniper.com,resources=quizquestions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=juniper.juniper.com,resources=quizquestions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *QuizQuestionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	logger := log.FromContext(ctx)
	logger.Info("Reconciling QuizQuestion", "Namespace", req.Namespace, "Name", req.Name)

	// Fetch the QuizQuestion resource
	quizQuestion := &juniperv1.QuizQuestion{}
	if err := r.Get(ctx, req.NamespacedName, quizQuestion); err != nil {
		if errors.IsNotFound(err) {
			// QuizQuestion resource not found, may have been deleted, so no need to reconcile
			logger.Info("QuizQuestion resource not found, skipping reconciliation")
			return ctrl.Result{}, nil
		}
		// Error fetching the resource; return an error
		return ctrl.Result{}, err
	}

	// Define a ConfigMap based on the QuizQuestion resource
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.Namespace,
			Name:      quizQuestion.Name + "-configmap", //
		},
		Data: map[string]string{
			"Question": quizQuestion.Spec.Question,
			"Answer":   quizQuestion.Spec.Answer,
		},
	}

	// Set the owner reference to the QuizQuestion resource
	if err := controllerutil.SetControllerReference(quizQuestion, configMap, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if the ConfigMap already exists
	foundConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: configMap.Name}, foundConfigMap)
	if err != nil && errors.IsNotFound(err) {
		// ConfigMap doesn't exist, so create it
		logger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
		err = r.Create(ctx, configMap)
		if err != nil {
			logger.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
			return ctrl.Result{}, err
		}
		// ConfigMap created successfully
		return ctrl.Result{}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get ConfigMap")
		return ctrl.Result{}, err
	}

	// ConfigMap already exists, so update
	if !configMapsEqual(configMap, foundConfigMap) {
		logger.Info("Updating ConfigMap", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)

		// Use retry.OnError to handle conflicts during updates
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Fetching the latest version of the ConfigMap
			err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: configMap.Name}, foundConfigMap)
			if err != nil {
				logger.Error(err, "Failed to fetch ConfigMap during update")
				return err
			}

			// Updating the ConfigMap data
			foundConfigMap.Data = configMap.Data

			// Updating the ConfigMap
			err = r.Update(ctx, foundConfigMap)
			if err != nil {
				logger.Error(err, "Failed to update ConfigMap", "ConfigMap.Namespace", foundConfigMap.Namespace, "ConfigMap.Name", foundConfigMap.Name)
				return err
			}

			return nil
		})

		if err != nil {
			logger.Error(err, "Failed to update ConfigMap with retry")
			return ctrl.Result{}, err
		}

	}

	return ctrl.Result{}, nil
}

func configMapsEqual(cm1, cm2 *corev1.ConfigMap) bool {
	return cm1.Data["Question"] == cm2.Data["Question"] &&
		cm1.Data["Answer"] == cm2.Data["Answer"]
}

// SetupWithManager sets up the controller with the Manager.
func (r *QuizQuestionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&juniperv1.QuizQuestion{}).
		Complete(r)
}

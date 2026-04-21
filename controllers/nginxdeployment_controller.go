package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	deployv1alpha1 "nginx-helm-operator/api/v1alpha1"
)

const (
	// Finalizer name for cleanup
	nginxDeploymentFinalizer = "deploy.example.com/nginx-deployment-finalizer"

	// Default values
	defaultImage = "nginx:latest"

	// Status phases
	phaseDeploying = "Deploying"
	phaseDeployed  = "Deployed"
	phaseFailed    = "Failed"
	phaseDeleting  = "Deleting"

	// Conflict resolution constants
	maxRetryAttempts  = 5
	initialRetryDelay = 1 * time.Second
	maxRetryDelay     = 32 * time.Second
	backoffMultiplier = 2.0
)

var (
	defaultReplicas int32 = 1
)

// NginxDeploymentReconciler reconciles a NginxDeployment object
type NginxDeploymentReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	Log            logr.Logger
	HelmChartsPath string
}

// +kubebuilder:rbac:groups=deploy.example.com,resources=nginxdeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=deploy.example.com,resources=nginxdeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=deploy.example.com,resources=nginxdeployments/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NginxDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Implement retry logic with exponential backoff for conflict resolution
	retryBackoff := wait.Backoff{
		Steps:    maxRetryAttempts,
		Duration: initialRetryDelay,
		Factor:   backoffMultiplier,
		Cap:      maxRetryDelay,
	}

	var lastErr error
	err := wait.ExponentialBackoff(retryBackoff, func() (bool, error) {
		// Fetch the NginxDeployment instance with latest resource version
		nginxDeployment := &deployv1alpha1.NginxDeployment{}
		err := r.Get(ctx, req.NamespacedName, nginxDeployment)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// Request object not found, could have been deleted after reconcile request.
				return true, nil // Success, no need to retry
			}
			// Error reading the object - retry
			logger.V(1).Info("Failed to get NginxDeployment, retrying", "error", err)
			lastErr = err
			return false, nil // Retry
		}

		// Check if the NginxDeployment instance is marked to be deleted
		if nginxDeployment.GetDeletionTimestamp() != nil {
			_, err := r.handleDeletion(ctx, nginxDeployment, logger)
			if err != nil {
				if apierrors.IsConflict(err) {
					logger.V(1).Info("Conflict during deletion, retrying", "error", err)
					lastErr = err
					return false, nil // Retry
				}
				lastErr = err
				return true, err // Don't retry non-conflict errors
			}
			return true, nil // Success
		}

		// Add finalizer if it doesn't exist
		if !controllerutil.ContainsFinalizer(nginxDeployment, nginxDeploymentFinalizer) {
			controllerutil.AddFinalizer(nginxDeployment, nginxDeploymentFinalizer)
			if err := r.updateWithConflictRetry(ctx, nginxDeployment, logger); err != nil {
				if apierrors.IsConflict(err) {
					logger.V(1).Info("Conflict adding finalizer, retrying", "error", err)
					lastErr = err
					return false, nil // Retry
				}
				lastErr = err
				return true, err // Don't retry non-conflict errors
			}
			return true, nil // Success, will requeue
		}

		// Handle the deployment
		_, deployErr := r.handleDeployment(ctx, nginxDeployment, logger)
		if deployErr != nil {
			if apierrors.IsConflict(deployErr) {
				logger.V(1).Info("Conflict during deployment, retrying", "error", deployErr)
				lastErr = deployErr
				return false, nil // Retry
			}
			lastErr = deployErr
			return true, deployErr // Don't retry non-conflict errors
		}
		return true, nil // Success
	})

	if err != nil {
		if lastErr != nil {
			logger.Error(lastErr, "Failed to reconcile after retries")
			return ctrl.Result{RequeueAfter: time.Minute * 2}, lastErr
		}
		logger.Error(err, "Retry logic failed")
		return ctrl.Result{RequeueAfter: time.Minute * 2}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

func (r *NginxDeploymentReconciler) handleDeployment(ctx context.Context, nginxDeployment *deployv1alpha1.NginxDeployment, logger logr.Logger) (ctrl.Result, error) {
	// Update status to deploying with conflict retry
	nginxDeployment.Status.Phase = phaseDeploying
	nginxDeployment.Status.Message = "Starting deployment"
	nginxDeployment.Status.LastUpdated = metav1.Now()

	if err := r.updateStatusWithConflictRetry(ctx, nginxDeployment, logger); err != nil {
		logger.Error(err, "Failed to update status to deploying")
		return ctrl.Result{}, err
	}

	// Determine target namespace
	targetNamespace := nginxDeployment.Spec.Namespace
	if targetNamespace == "" {
		targetNamespace = nginxDeployment.Namespace
	}

	// Ensure target namespace exists
	if err := r.ensureNamespace(ctx, targetNamespace); err != nil {
		logger.Error(err, "Failed to ensure namespace exists", "namespace", targetNamespace)
		return r.updateStatusWithError(ctx, nginxDeployment, fmt.Sprintf("Failed to ensure namespace: %v", err))
	}

	// Deploy or update Helm chart
	if err := r.deployHelmChart(ctx, nginxDeployment, targetNamespace, logger); err != nil {
		logger.Error(err, "Failed to deploy Helm chart")
		return r.updateStatusWithError(ctx, nginxDeployment, fmt.Sprintf("Failed to deploy Helm chart: %v", err))
	}

	// Update status to deployed with conflict retry
	nginxDeployment.Status.Phase = phaseDeployed
	nginxDeployment.Status.Message = "Deployment successful"
	nginxDeployment.Status.LastUpdated = metav1.Now()
	nginxDeployment.Status.HelmReleaseStatus = "deployed"

	if err := r.updateStatusWithConflictRetry(ctx, nginxDeployment, logger); err != nil {
		logger.Error(err, "Failed to update status to deployed")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully deployed nginx", "deployment", nginxDeployment.Spec.DeploymentName, "namespace", targetNamespace)
	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil // Requeue for periodic checks
}

func (r *NginxDeploymentReconciler) handleDeletion(ctx context.Context, nginxDeployment *deployv1alpha1.NginxDeployment, logger logr.Logger) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(nginxDeployment, nginxDeploymentFinalizer) {
		return ctrl.Result{}, nil
	}

	// Update status to deleting with conflict retry
	nginxDeployment.Status.Phase = phaseDeleting
	nginxDeployment.Status.Message = "Deleting deployment"
	nginxDeployment.Status.LastUpdated = metav1.Now()

	if err := r.updateStatusWithConflictRetry(ctx, nginxDeployment, logger); err != nil {
		logger.Error(err, "Failed to update status to deleting")
	}

	// Determine target namespace
	targetNamespace := nginxDeployment.Spec.Namespace
	if targetNamespace == "" {
		targetNamespace = nginxDeployment.Namespace
	}

	// Uninstall Helm release
	if err := r.uninstallHelmChart(ctx, nginxDeployment, targetNamespace, logger); err != nil {
		logger.Error(err, "Failed to uninstall Helm chart")
		// Continue with finalizer removal even if uninstall fails
	}

	// Remove finalizer with conflict retry
	controllerutil.RemoveFinalizer(nginxDeployment, nginxDeploymentFinalizer)
	if err := r.updateWithConflictRetry(ctx, nginxDeployment, logger); err != nil {
		logger.Error(err, "Failed to remove finalizer")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully deleted nginx deployment", "deployment", nginxDeployment.Spec.DeploymentName, "namespace", targetNamespace)
	return ctrl.Result{}, nil
}

func (r *NginxDeploymentReconciler) ensureNamespace(ctx context.Context, namespace string) error {
	ns := &corev1.Namespace{}
	err := r.Get(ctx, types.NamespacedName{Name: namespace}, ns)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create namespace if it doesn't exist
			ns = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			return r.Create(ctx, ns)
		}
		return err
	}
	return nil
}

func (r *NginxDeploymentReconciler) deployHelmChart(ctx context.Context, nginxDeployment *deployv1alpha1.NginxDeployment, namespace string, logger logr.Logger) error {
	// Setup Helm configuration
	settings := cli.New()
	// Ensure we're using the correct namespace for Helm operations
	settings.SetNamespace(namespace)
	actionConfig := new(action.Configuration)

	// Initialize Helm action configuration with the target namespace
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, "secret", func(format string, v ...interface{}) {
		logger.Info(fmt.Sprintf(format, v...))
	}); err != nil {
		return fmt.Errorf("failed to initialize Helm action config: %w", err)
	}

	// Load chart from the charts directory
	chartPath := filepath.Join(r.HelmChartsPath, "nginx")
	chart, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("failed to load chart from %s: %w", chartPath, err)
	}

	// Prepare values
	values := r.prepareHelmValues(nginxDeployment)

	// Check if release already exists
	releaseName := nginxDeployment.Spec.DeploymentName
	getAction := action.NewGet(actionConfig)
	_, err = getAction.Run(releaseName)

	if err != nil {
		// Release doesn't exist, install it
		installAction := action.NewInstall(actionConfig)
		installAction.Namespace = namespace
		installAction.ReleaseName = releaseName
		installAction.CreateNamespace = true
		installAction.Wait = true
		installAction.Timeout = time.Minute * 5

		release, err := installAction.Run(chart, values)
		if err != nil {
			return fmt.Errorf("failed to install Helm chart: %w", err)
		}

		nginxDeployment.Status.DeployedRevision = release.Version
		logger.Info("Installed Helm chart", "release", releaseName, "revision", release.Version)
	} else {
		// Release exists, upgrade it
		upgradeAction := action.NewUpgrade(actionConfig)
		upgradeAction.Namespace = namespace
		upgradeAction.Wait = true
		upgradeAction.Timeout = time.Minute * 5
		upgradeAction.ResetValues = false
		upgradeAction.ReuseValues = false

		release, err := upgradeAction.Run(releaseName, chart, values)
		if err != nil {
			return fmt.Errorf("failed to upgrade Helm chart: %w", err)
		}

		nginxDeployment.Status.DeployedRevision = release.Version
		logger.Info("Upgraded Helm chart", "release", releaseName, "revision", release.Version)
	}

	return nil
}

func (r *NginxDeploymentReconciler) uninstallHelmChart(ctx context.Context, nginxDeployment *deployv1alpha1.NginxDeployment, namespace string, logger logr.Logger) error {
	// Setup Helm configuration
	settings := cli.New()
	// Ensure we're using the correct namespace for Helm operations
	settings.SetNamespace(namespace)
	actionConfig := new(action.Configuration)

	// Initialize Helm action configuration with the target namespace
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, "secret", func(format string, v ...interface{}) {
		logger.Info(fmt.Sprintf(format, v...))
	}); err != nil {
		return fmt.Errorf("failed to initialize Helm action config: %w", err)
	}

	// Uninstall the release
	uninstallAction := action.NewUninstall(actionConfig)
	uninstallAction.Wait = true
	uninstallAction.Timeout = time.Minute * 5

	releaseName := nginxDeployment.Spec.DeploymentName
	_, err := uninstallAction.Run(releaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall Helm chart: %w", err)
	}

	logger.Info("Uninstalled Helm chart", "release", releaseName)
	return nil
}

func (r *NginxDeploymentReconciler) prepareHelmValues(nginxDeployment *deployv1alpha1.NginxDeployment) map[string]interface{} {
	values := make(map[string]interface{})

	// Set default values
	image := nginxDeployment.Spec.Image
	if image == "" {
		image = defaultImage
	}

	replicas := nginxDeployment.Spec.Replicas
	if replicas == nil {
		defaultReplicasCopy := defaultReplicas
		replicas = &defaultReplicasCopy
	}

	// Basic values
	values["image"] = map[string]interface{}{
		"repository": "nginx",
		"tag":        "latest",
	}

	// Parse image if it contains tag
	if image != "" {
		if imageMap, ok := parseImageString(image); ok {
			values["image"] = imageMap
		}
	}

	values["replicaCount"] = *replicas
	values["nameOverride"] = nginxDeployment.Spec.DeploymentName
	values["fullnameOverride"] = nginxDeployment.Spec.DeploymentName

	// Merge custom Helm values from RawExtension
	if nginxDeployment.Spec.HelmValues != nil {
		customValues := make(map[string]interface{})
		if err := json.Unmarshal(nginxDeployment.Spec.HelmValues.Raw, &customValues); err == nil {
			for key, value := range customValues {
				values[key] = value
			}
		}
	}

	return values
}

func parseImageString(imageStr string) (map[string]interface{}, bool) {
	// Simple parsing for image:tag format
	if imageStr == "" {
		return nil, false
	}

	parts := strings.Split(imageStr, ":")
	if len(parts) == 1 {
		return map[string]interface{}{
			"repository": parts[0],
			"tag":        "latest",
		}, true
	} else if len(parts) == 2 {
		return map[string]interface{}{
			"repository": parts[0],
			"tag":        parts[1],
		}, true
	}

	return nil, false
}

func (r *NginxDeploymentReconciler) updateStatusWithError(ctx context.Context, nginxDeployment *deployv1alpha1.NginxDeployment, errorMsg string) (ctrl.Result, error) {
	nginxDeployment.Status.Phase = phaseFailed
	nginxDeployment.Status.Message = errorMsg
	nginxDeployment.Status.LastUpdated = metav1.Now()

	if err := r.updateStatusWithConflictRetry(ctx, nginxDeployment, r.Log); err != nil {
		r.Log.Error(err, "Failed to update status with error")
	}

	return ctrl.Result{RequeueAfter: time.Minute * 2}, nil
}

// updateWithConflictRetry updates the resource with retry logic for conflict resolution
func (r *NginxDeploymentReconciler) updateWithConflictRetry(ctx context.Context, nginxDeployment *deployv1alpha1.NginxDeployment, logger logr.Logger) error {
	retryBackoff := wait.Backoff{
		Steps:    3, // Fewer retries for update operations
		Duration: 100 * time.Millisecond,
		Factor:   2.0,
		Cap:      1 * time.Second,
	}

	return wait.ExponentialBackoff(retryBackoff, func() (bool, error) {
		// Get the latest version of the resource
		latestNginxDeployment := &deployv1alpha1.NginxDeployment{}
		err := r.Get(ctx, client.ObjectKeyFromObject(nginxDeployment), latestNginxDeployment)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, err // Don't retry if resource is deleted
			}
			logger.V(1).Info("Failed to get latest resource version for update", "error", err)
			return false, nil // Retry
		}

		// Copy the desired changes to the latest version
		latestNginxDeployment.Finalizers = nginxDeployment.Finalizers

		// Attempt the update
		err = r.Update(ctx, latestNginxDeployment)
		if err != nil {
			if apierrors.IsConflict(err) {
				logger.V(1).Info("Update conflict, retrying", "error", err)
				return false, nil // Retry
			}
			return true, err // Don't retry non-conflict errors
		}

		// Update successful
		return true, nil
	})
}

// updateStatusWithConflictRetry updates the status with retry logic for conflict resolution
func (r *NginxDeploymentReconciler) updateStatusWithConflictRetry(ctx context.Context, nginxDeployment *deployv1alpha1.NginxDeployment, logger logr.Logger) error {
	retryBackoff := wait.Backoff{
		Steps:    3, // Fewer retries for status operations
		Duration: 100 * time.Millisecond,
		Factor:   2.0,
		Cap:      1 * time.Second,
	}

	return wait.ExponentialBackoff(retryBackoff, func() (bool, error) {
		// Get the latest version of the resource
		latestNginxDeployment := &deployv1alpha1.NginxDeployment{}
		err := r.Get(ctx, client.ObjectKeyFromObject(nginxDeployment), latestNginxDeployment)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, err // Don't retry if resource is deleted
			}
			logger.V(1).Info("Failed to get latest resource version for status update", "error", err)
			return false, nil // Retry
		}

		// Copy the desired status changes to the latest version
		latestNginxDeployment.Status = nginxDeployment.Status

		// Attempt the status update
		err = r.Status().Update(ctx, latestNginxDeployment)
		if err != nil {
			if apierrors.IsConflict(err) {
				logger.V(1).Info("Status update conflict, retrying", "error", err)
				return false, nil // Retry
			}
			return true, err // Don't retry non-conflict errors
		}

		// Status update successful
		return true, nil
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&deployv1alpha1.NginxDeployment{}).
		Complete(r)
}

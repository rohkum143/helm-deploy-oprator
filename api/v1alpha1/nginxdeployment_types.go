package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NginxDeploymentSpec defines the desired state of NginxDeployment
type NginxDeploymentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make generate" to regenerate code after modifying this file

	// DeploymentName is the name of the nginx deployment
	DeploymentName string `json:"deploymentName"`

	// Image is the nginx image to use (optional, defaults to nginx:latest)
	// +optional
	Image string `json:"image,omitempty"`

	// Replicas is the number of nginx replicas (optional, defaults to 1)
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Namespace is the target namespace for deployment (optional, uses CR namespace if not specified)
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// HelmValues allows overriding specific Helm chart values
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	HelmValues *runtime.RawExtension `json:"helmValues,omitempty"`

	// ChartVersion specifies the version of the Helm chart to use (optional)
	// +optional
	ChartVersion string `json:"chartVersion,omitempty"`
}

// NginxDeploymentStatus defines the observed state of NginxDeployment
type NginxDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make generate" to regenerate code after modifying this file

	// Phase represents the current phase of the deployment
	Phase string `json:"phase,omitempty"`

	// Message provides additional information about the current status
	Message string `json:"message,omitempty"`

	// LastUpdated is the timestamp of the last status update
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`

	// HelmReleaseStatus contains the status of the Helm release
	HelmReleaseStatus string `json:"helmReleaseStatus,omitempty"`

	// DeployedRevision is the revision number of the deployed Helm release
	DeployedRevision int `json:"deployedRevision,omitempty"`

	// Conditions represent the latest available observations of the deployment's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Deployment",type="string",JSONPath=".spec.deploymentName"
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.image"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// NginxDeployment is the Schema for the nginxdeployments API
type NginxDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NginxDeploymentSpec   `json:"spec,omitempty"`
	Status NginxDeploymentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NginxDeploymentList contains a list of NginxDeployment
type NginxDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NginxDeployment `json:"items"`
}

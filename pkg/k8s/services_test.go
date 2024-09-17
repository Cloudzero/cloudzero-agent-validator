package k8s_test

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/cloudzero/cloudzero-agent-validator/pkg/k8s"
)

// TestGetServiceURLs tests the GetServiceURLs function
func TestGetServiceURLs(t *testing.T) {
	tests := []struct {
		name                    string
		services                []corev1.Service
		expectedKubeStateURL    string
		expectedNodeExporterURL string
		expectError             bool
	}{
		{
			name: "Both services found",
			services: []corev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-state-metrics",
						Namespace: "default",
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{Port: 8080},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "node-exporter",
						Namespace: "default",
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{Port: 9100},
						},
					},
				},
			},
			expectedKubeStateURL:    "http://kube-state-metrics.default.svc.cluster.local:8080",
			expectedNodeExporterURL: "http://node-exporter.default.svc.cluster.local:9100",
			expectError:             false,
		},
		{
			name: "Kube-state-metrics service not found",
			services: []corev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "node-exporter",
						Namespace: "default",
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{Port: 9100},
						},
					},
				},
			},
			expectedKubeStateURL:    "",
			expectedNodeExporterURL: "",
			expectError:             true,
		},
		{
			name: "Node-exporter service not found",
			services: []corev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-state-metrics",
						Namespace: "default",
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{Port: 8080},
						},
					},
				},
			},
			expectedKubeStateURL:    "",
			expectedNodeExporterURL: "",
			expectError:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(&corev1.ServiceList{Items: tt.services})

			kubeStateMetricsURL, nodeExporterURL, err := k8s.GetServiceURLs(context.Background(), clientset)
			if (err != nil) != tt.expectError {
				t.Errorf("GetServiceURLs() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if kubeStateMetricsURL != tt.expectedKubeStateURL {
				t.Errorf("GetServiceURLs() kubeStateMetricsURL = %v, expected %v", kubeStateMetricsURL, tt.expectedKubeStateURL)
			}
			if nodeExporterURL != tt.expectedNodeExporterURL {
				t.Errorf("GetServiceURLs() nodeExporterURL = %v, expected %v", nodeExporterURL, tt.expectedNodeExporterURL)
			}
		})
	}
}

// TestGetConfigMap tests the GetConfigMap function
func TestGetConfigMap(t *testing.T) {
	tests := []struct {
		name          string
		configMaps    []corev1.ConfigMap
		namespace     string
		configMapName string
		expectError   bool
	}{
		{
			name: "ConfigMap found",
			configMaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: "default",
					},
					Data: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			namespace:     "default",
			configMapName: "test-configmap",
			expectError:   false,
		},
		{
			name:          "ConfigMap not found",
			configMaps:    []corev1.ConfigMap{},
			namespace:     "default",
			configMapName: "nonexistent-configmap",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(&corev1.ConfigMapList{Items: tt.configMaps})

			configMap, err := k8s.GetConfigMap(context.Background(), clientset, tt.namespace, tt.configMapName)
			if (err != nil) != tt.expectError {
				t.Errorf("GetConfigMap() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && configMap.Name != tt.configMapName {
				t.Errorf("GetConfigMap() configMap.Name = %v, expected %v", configMap.Name, tt.configMapName)
			}
		})
	}
}

// TestUpdateConfigMap tests the UpdateConfigMap function
func TestUpdateConfigMap(t *testing.T) {
	tests := []struct {
		name             string
		initialConfigMap *corev1.ConfigMap
		updatedData      map[string]string
		expectError      bool
	}{
		{
			name: "Update ConfigMap successfully",
			initialConfigMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-configmap",
					Namespace: "default",
				},
				Data: map[string]string{
					"key1": "value1",
				},
			},
			updatedData: map[string]string{
				"key1": "new-value1",
				"key2": "value2",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(tt.initialConfigMap)

			// Update the ConfigMap data
			tt.initialConfigMap.Data = tt.updatedData

			err := k8s.UpdateConfigMap(context.Background(), clientset, tt.initialConfigMap.Namespace, tt.initialConfigMap)
			if (err != nil) != tt.expectError {
				t.Errorf("UpdateConfigMap() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// Verify the ConfigMap was updated
			updatedConfigMap, err := k8s.GetConfigMap(context.Background(), clientset, tt.initialConfigMap.Namespace, tt.initialConfigMap.Name)
			if err != nil {
				t.Errorf("GetConfigMap() error = %v", err)
				return
			}

			for key, expectedValue := range tt.updatedData {
				if updatedConfigMap.Data[key] != expectedValue {
					t.Errorf("ConfigMap data mismatch for key %s: got %v, expected %v", key, updatedConfigMap.Data[key], expectedValue)
				}
			}
		})
	}
}

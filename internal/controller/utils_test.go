package controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Utils", func() {
	Describe("applyResourceDefaults", func() {
		It("should set default CPU and memory limits", func() {
			resources := corev1.ResourceRequirements{}

			result := applyResourceDefaults(resources)

			// Check that defaults are set when empty
			_, cpuReqExists := result.Requests[corev1.ResourceCPU]
			_, memReqExists := result.Requests[corev1.ResourceMemory]
			_, cpuLimitExists := result.Limits[corev1.ResourceCPU]
			_, memLimitExists := result.Limits[corev1.ResourceMemory]

			Expect(cpuReqExists).To(BeTrue())
			Expect(memReqExists).To(BeTrue())
			Expect(cpuLimitExists).To(BeTrue())
			Expect(memLimitExists).To(BeTrue())
		})

		It("should preserve existing resource values", func() {
			resources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("1000m"),
				},
			}

			result := applyResourceDefaults(resources)

			// Should preserve existing CPU and add missing memory/CPU limit
			_, cpuReqExists := result.Requests[corev1.ResourceCPU]
			_, memReqExists := result.Requests[corev1.ResourceMemory]
			_, cpuLimitExists := result.Limits[corev1.ResourceCPU]
			_, memLimitExists := result.Limits[corev1.ResourceMemory]

			Expect(cpuReqExists).To(BeTrue())   // existed before
			Expect(memReqExists).To(BeTrue())   // added by function
			Expect(cpuLimitExists).To(BeTrue()) // added by function
			Expect(memLimitExists).To(BeTrue()) // added by function
		})
	})

	Describe("GetSecureGameServerContainer", func() {
		It("should create a secure container with defaults", func() {
			ports := []corev1.ContainerPort{
				{ContainerPort: 16261, Name: "tcp", Protocol: corev1.ProtocolTCP},
			}

			container := GetSecureGameServerContainer("test-server", "test-image:latest", corev1.ResourceRequirements{}, ports)

			Expect(container.Name).To(Equal("test-server"))
			Expect(container.Image).To(Equal("test-image:latest"))
			// These fields are not explicitly set, so they should be nil
			Expect(container.SecurityContext.RunAsNonRoot).To(BeNil()) // LinuxGSM needs root, so this is not set
			Expect(container.SecurityContext.AllowPrivilegeEscalation).To(BeNil())
			Expect(container.SecurityContext.ReadOnlyRootFilesystem).To(BeNil())
			Expect(len(container.Ports)).To(Equal(1))
		})
	})

	Describe("GetSecureCodeServerContainer", func() {
		It("should create a secure code-server container", func() {
			container := GetSecureCodeServerContainer("test-password")

			Expect(container.Name).To(Equal("code-server"))
			Expect(container.Image).To(Equal("codercom/code-server:latest"))
			Expect(container.Resources.Requests.Cpu().String()).To(Equal("100m"))
			Expect(container.Resources.Limits.Memory().String()).To(Equal("512Mi"))
			// This field is not explicitly set, so it should be nil
			Expect(container.SecurityContext.RunAsNonRoot).To(BeNil())
		})
	})

	Describe("CompareDeployments", func() {
		It("should return true for equivalent deployments", func() {
			dep1 := createTestDeployment(1)
			dep2 := createTestDeployment(1)

			result := CompareDeployments(dep1, dep2)
			Expect(result).To(BeTrue())
		})

		It("should return false for deployments with different replicas", func() {
			dep1 := createTestDeployment(1)
			dep2 := createTestDeployment(2)

			result := CompareDeployments(dep1, dep2)
			Expect(result).To(BeFalse())
		})
	})
})

func createTestDeployment(replicas int32) *appsv1.Deployment {
	if replicas == 0 {
		return &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"test": "true"}},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"test": "true"}},
				},
			},
		}
	}

	return &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"test": "true"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"test": "true"}},
			},
		},
	}
}

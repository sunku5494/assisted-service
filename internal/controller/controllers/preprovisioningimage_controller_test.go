package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	metal3_v1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	aiv1beta1 "github.com/openshift/assisted-service/api/v1beta1"
	"github.com/openshift/assisted-service/internal/bminventory"
	"github.com/openshift/assisted-service/internal/common"
	"github.com/openshift/assisted-service/internal/ignition"
	"github.com/openshift/assisted-service/models"
	"github.com/openshift/assisted-service/restapi/operations/installer"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newPreprovisioningImageRequest(image *metal3_v1alpha1.PreprovisioningImage) ctrl.Request {
	namespacedName := types.NamespacedName{
		Namespace: image.ObjectMeta.Namespace,
		Name:      image.ObjectMeta.Name,
	}
	return ctrl.Request{NamespacedName: namespacedName}
}

func newPreprovisioningImage(name, namespace string, labelKey string, labelValue string) *metal3_v1alpha1.PreprovisioningImage {
	return &metal3_v1alpha1.PreprovisioningImage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    AddLabel(nil, labelKey, labelValue),
		},
		Spec: metal3_v1alpha1.PreprovisioningImageSpec{},
	}
}

func newInfraEnv(name, namespace string, spec aiv1beta1.InfraEnvSpec) *aiv1beta1.InfraEnv {
	return &aiv1beta1.InfraEnv{
		TypeMeta: metav1.TypeMeta{
			Kind:       "InfraEnv",
			APIVersion: fmt.Sprintf("%s/%s", aiv1beta1.GroupVersion.Group, aiv1beta1.GroupVersion.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
}

var _ = Describe("PreprovisioningImage reconcile", func() {
	var (
		c                     client.Client
		pr                    *PreprovisioningImageReconciler
		mockCtrl              *gomock.Controller
		mockInstallerInternal *bminventory.MockInstallerInternals
		mockCRDEventsHandler  *MockCRDEventsHandler
		ironicIgnitionBuilder ignition.IronicIgniotionBuilder
		ctx                   = context.Background()
		sId                   strfmt.UUID
		backendInfraEnv       = &common.InfraEnv{InfraEnv: models.InfraEnv{ClusterID: sId, ID: &sId}}
		downloadURL           = "https://downloadurl"
		infraEnvArch          = "x86_64"
		infraEnv              *aiv1beta1.InfraEnv
		ppi                   *metal3_v1alpha1.PreprovisioningImage
	)

	BeforeEach(func() {
		c = fakeclient.NewClientBuilder().WithScheme(scheme.Scheme).Build()
		mockCtrl = gomock.NewController(GinkgoT())
		mockInstallerInternal = bminventory.NewMockInstallerInternals(mockCtrl)
		mockCRDEventsHandler = NewMockCRDEventsHandler(mockCtrl)
		ironicIgnitionBuilder = ignition.NewIronicIgniotionBuilder(ignition.IronicIgniotionBuilderConfig{BaremetalIronicAgentImage: "ironic-agent-image:latest"})
		sId = strfmt.UUID(uuid.New().String())
		pr = &PreprovisioningImageReconciler{
			Client:                 c,
			Log:                    common.GetTestLog(),
			Installer:              mockInstallerInternal,
			CRDEventsHandler:       mockCRDEventsHandler,
			IronicIgniotionBuilder: ironicIgnitionBuilder,
			IronicServiceURL:       "ironic.url",
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("reconcile new PreprovisioningImage - success", func() {
		BeforeEach(func() {
			infraEnv = newInfraEnv("testInfraEnv", testNamespace, aiv1beta1.InfraEnvSpec{})
			infraEnv.Status.ISODownloadURL = downloadURL
			createdAt := metav1.Now().Add(-InfraEnvImageCooldownPeriod)
			infraEnv.Status.CreatedTime = &metav1.Time{Time: createdAt}
			infraEnv.Status.Conditions = []conditionsv1.Condition{{Type: aiv1beta1.ImageCreatedCondition,
				Status:  corev1.ConditionTrue,
				Reason:  "some reason",
				Message: "Some message",
			}}
			ppi = newPreprovisioningImage("testPPI", testNamespace, InfraEnvLabel, "testInfraEnv")
			Expect(c.Create(ctx, ppi)).To(BeNil())
		})
		AfterEach(func() {
			mockCtrl.Finish()
		})
		It("Add the ironic Agent to the infraEnv", func() {
			Expect(c.Create(ctx, infraEnv)).To(BeNil())
			mockInstallerInternal.EXPECT().GetInfraEnvByKubeKey(gomock.Any()).Return(backendInfraEnv, nil)
			mockInstallerInternal.EXPECT().UpdateInfraEnvInternal(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, params installer.UpdateInfraEnvParams, internalIgnitionConfig *string) {
					Expect(params.InfraEnvID).To(Equal(*backendInfraEnv.ID))
					Expect(params.InfraEnvUpdateParams.IgnitionConfigOverride).To(Equal(""))
					Expect(*internalIgnitionConfig).Should(ContainSubstring("ironic"))
				}).Return(
				&common.InfraEnv{InfraEnv: models.InfraEnv{ClusterID: sId, ID: &sId, DownloadURL: downloadURL, CPUArchitecture: infraEnvArch}, GeneratedAt: strfmt.DateTime(time.Now())}, nil).Times(1)
			mockCRDEventsHandler.EXPECT().NotifyInfraEnvUpdates(infraEnv.Name, infraEnv.Namespace).Times(1)
			res, err := pr.Reconcile(ctx, newPreprovisioningImageRequest(ppi))
			Expect(err).To(BeNil())
			Expect(res).To(Equal(ctrl.Result{}))

			key := types.NamespacedName{
				Namespace: testNamespace,
				Name:      "testInfraEnv",
			}
			Expect(c.Get(ctx, key, infraEnv)).To(BeNil())
			Expect(infraEnv.ObjectMeta.Annotations[EnableIronicAgentAnnotation]).To(Equal("true"))
		})
		It("Wait for InfraEnv cool down", func() {
			infraEnv.Status.ISODownloadURL = downloadURL
			infraEnv.Status.CreatedTime = &metav1.Time{Time: time.Now()}
			infraEnv.Status.Conditions = []conditionsv1.Condition{{Type: aiv1beta1.ImageCreatedCondition,
				Status:  corev1.ConditionTrue,
				Reason:  "some reason",
				Message: "Some message",
			}}
			infraEnv.ObjectMeta.Annotations = make(map[string]string)
			infraEnv.ObjectMeta.Annotations[EnableIronicAgentAnnotation] = "true"
			Expect(c.Create(ctx, infraEnv)).To(BeNil())
			Expect(infraEnv.ObjectMeta.Annotations[EnableIronicAgentAnnotation]).To(Equal("true"))
			res, err := pr.Reconcile(ctx, newPreprovisioningImageRequest(ppi))
			Expect(err).To(BeNil())
			Expect(res.Requeue).To(Equal(true))
			key := types.NamespacedName{
				Namespace: testNamespace,
				Name:      "testPPI",
			}
			Expect(c.Get(ctx, key, ppi)).To(BeNil())
			validateStatus("",
				&conditionsv1.Condition{
					Reason:  "Waiting for InfraEnv image to cool down",
					Message: "Waiting for InfraEnv image to cool down",
					Status:  corev1.ConditionFalse},
				ppi,
			)
		})

		It("Set the image on the PPI", func() {
			infraEnv.Status.ISODownloadURL = downloadURL
			createdAt := metav1.Now().Add(-InfraEnvImageCooldownPeriod)
			infraEnv.Status.CreatedTime = &metav1.Time{Time: createdAt}
			infraEnv.Status.Conditions = []conditionsv1.Condition{{Type: aiv1beta1.ImageCreatedCondition,
				Status:  corev1.ConditionTrue,
				Reason:  "some reason",
				Message: "Some message",
			}}
			infraEnv.ObjectMeta.Annotations = make(map[string]string)
			infraEnv.ObjectMeta.Annotations[EnableIronicAgentAnnotation] = "true"
			Expect(c.Create(ctx, infraEnv)).To(BeNil())
			Expect(infraEnv.ObjectMeta.Annotations[EnableIronicAgentAnnotation]).To(Equal("true"))

			res, err := pr.Reconcile(ctx, newPreprovisioningImageRequest(ppi))
			Expect(err).To(BeNil())
			Expect(res).To(Equal(ctrl.Result{}))
			key := types.NamespacedName{
				Namespace: testNamespace,
				Name:      "testPPI",
			}
			Expect(c.Get(ctx, key, ppi)).To(BeNil())
			validateStatus(infraEnv.Status.ISODownloadURL, conditionsv1.FindStatusCondition(infraEnv.Status.Conditions, aiv1beta1.ImageCreatedCondition), ppi)
		})
		It("PreprovisioningImage ImageUrl is up to date", func() {
			infraEnv.Status.ISODownloadURL = downloadURL
			createdAt := metav1.Now().Add(-InfraEnvImageCooldownPeriod)
			infraEnv.Status.CreatedTime = &metav1.Time{Time: createdAt}
			infraEnv.Status.Conditions = []conditionsv1.Condition{{Type: aiv1beta1.ImageCreatedCondition,
				Status:  corev1.ConditionTrue,
				Reason:  "some reason",
				Message: "Some message",
			}}
			infraEnv.ObjectMeta.Annotations = make(map[string]string)
			infraEnv.ObjectMeta.Annotations[EnableIronicAgentAnnotation] = "true"
			Expect(c.Create(ctx, infraEnv)).To(BeNil())
			Expect(infraEnv.ObjectMeta.Annotations[EnableIronicAgentAnnotation]).To(Equal("true"))
			SetImageUrl(ppi, *infraEnv)
			Expect(c.Update(ctx, ppi)).To(BeNil())

			res, err := pr.Reconcile(ctx, newPreprovisioningImageRequest(ppi))
			Expect(err).To(BeNil())
			Expect(res).To(Equal(ctrl.Result{}))
			key := types.NamespacedName{
				Namespace: testNamespace,
				Name:      "testPPI",
			}
			Expect(c.Get(ctx, key, ppi)).To(BeNil())
			validateStatus(infraEnv.Status.ISODownloadURL, conditionsv1.FindStatusCondition(infraEnv.Status.Conditions, aiv1beta1.ImageCreatedCondition), ppi)
		})

		It("infraEnv not found", func() {
			res, err := pr.Reconcile(ctx, newPreprovisioningImageRequest(ppi))
			Expect(err).To(BeNil())
			Expect(res).To(Equal(ctrl.Result{}))
		})
	})
	It("PreprovisioningImage not found", func() {
		res, err := pr.Reconcile(ctx, newPreprovisioningImageRequest(ppi))
		Expect(err).To(BeNil())
		Expect(res).To(Equal(ctrl.Result{}))
	})
	It("internalInfraEnv not found", func() {
		ppi = newPreprovisioningImage("testPPI", testNamespace, InfraEnvLabel, "testInfraEnv")
		Expect(c.Create(ctx, ppi)).To(BeNil())
		infraEnv = newInfraEnv("testInfraEnv", testNamespace, aiv1beta1.InfraEnvSpec{})
		Expect(c.Create(ctx, infraEnv)).To(BeNil())
		mockInstallerInternal.EXPECT().GetInfraEnvByKubeKey(gomock.Any()).Return(backendInfraEnv, errors.New("Failed to get internal infra env"))

		res, err := pr.Reconcile(ctx, newPreprovisioningImageRequest(ppi))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Failed to get internal infra env"))
		Expect(res).To(Equal(ctrl.Result{}))
	})
	It("Failed during getIronicIgnitionConfig", func() {
		ppi = newPreprovisioningImage("testPPI", testNamespace, InfraEnvLabel, "testInfraEnv")
		Expect(c.Create(ctx, ppi)).To(BeNil())
		infraEnv = newInfraEnv("testInfraEnv", testNamespace, aiv1beta1.InfraEnvSpec{})
		infraEnv.ObjectMeta.Annotations = make(map[string]string)
		infraEnv.ObjectMeta.Annotations[EnableIronicAgentAnnotation] = "invalid value"
		mockInstallerInternal.EXPECT().GetInfraEnvByKubeKey(gomock.Any()).Return(backendInfraEnv, nil)
		Expect(c.Create(ctx, infraEnv)).To(BeNil())
		// This should fail the IronicIgnitionBuilder
		pr.IronicServiceURL = ""
		res, err := pr.Reconcile(ctx, newPreprovisioningImageRequest(ppi))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("ironicBaseURL is required"))
		Expect(res).To(Equal(ctrl.Result{}))
	})
	It("Failed to UpdateInfraEnvInternal", func() {
		ppi = newPreprovisioningImage("testPPI", testNamespace, InfraEnvLabel, "testInfraEnv")
		Expect(c.Create(ctx, ppi)).To(BeNil())
		infraEnv = newInfraEnv("testInfraEnv", testNamespace, aiv1beta1.InfraEnvSpec{})
		infraEnv.ObjectMeta.Annotations = make(map[string]string)
		infraEnv.ObjectMeta.Annotations[EnableIronicAgentAnnotation] = "invalid value"
		mockInstallerInternal.EXPECT().GetInfraEnvByKubeKey(gomock.Any()).Return(backendInfraEnv, nil)
		Expect(c.Create(ctx, infraEnv)).To(BeNil())
		mockInstallerInternal.EXPECT().UpdateInfraEnvInternal(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("Failed to update infraEnvInternal"))

		res, err := pr.Reconcile(ctx, newPreprovisioningImageRequest(ppi))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Failed to update infraEnvInternal"))
		Expect(res).To(Equal(ctrl.Result{}))
	})

})

func SetImageUrl(ppi *metal3_v1alpha1.PreprovisioningImage, infraEnv aiv1beta1.InfraEnv) {
	ppi.Status.ImageUrl = infraEnv.Status.ISODownloadURL
	ppi.Status.Conditions = []metav1.Condition{
		{Type: string(metal3_v1alpha1.ConditionImageReady),
			Reason:  infraEnv.Status.Conditions[0].Reason,
			Message: infraEnv.Status.Conditions[0].Message,
			Status:  metav1.ConditionStatus(infraEnv.Status.Conditions[0].Status)},
		{Type: string(metal3_v1alpha1.ConditionImageError),
			Reason:  infraEnv.Status.Conditions[0].Reason,
			Message: infraEnv.Status.Conditions[0].Message,
			Status:  metav1.ConditionFalse},
	}
}

func validateStatus(imageURL string, ExpectedImageReadyCondition *conditionsv1.Condition, ppi *metal3_v1alpha1.PreprovisioningImage) {
	Expect(imageURL).To(Equal(ppi.Status.ImageUrl))
	readyCondition := meta.FindStatusCondition(ppi.Status.Conditions, string(metal3_v1alpha1.ConditionImageReady))
	Expect(metav1.ConditionStatus(ExpectedImageReadyCondition.Status)).To(Equal(readyCondition.Status))
	Expect(ExpectedImageReadyCondition.Message).To(Equal(readyCondition.Message))
	Expect(ExpectedImageReadyCondition.Reason).To(Equal(readyCondition.Reason))
	Expect(corev1.ConditionFalse).To(Not(Equal(
		meta.FindStatusCondition(ppi.Status.Conditions, string(metal3_v1alpha1.ConditionImageError)).Status)))

}
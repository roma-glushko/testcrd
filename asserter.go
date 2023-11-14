package testcrd

import (
	"fmt"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubeyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kube-openapi/pkg/validation/validate"
	"os"
)

// TestingT is an interface wrapper around *testing.T
type TestingT interface {
	Errorf(format string, args ...interface{})
}

type ResourceAsserter struct {
	validators map[schema.GroupVersionKind]*validate.SchemaValidator
}

func NewResourceAsserter() *ResourceAsserter {
	return &ResourceAsserter{
		validators: make(map[schema.GroupVersionKind]*validate.SchemaValidator),
	}
}

// Load CRD from a byte buffer
func (a *ResourceAsserter) Load(crds []*apiextensions.CustomResourceDefinition) error {
	for _, crd := range crds {
		versions := crd.Spec.Versions

		if len(versions) == 0 {
			versions = []apiextensions.CustomResourceDefinitionVersion{{Name: crd.Spec.Version}} // nolint: staticcheck
		}

		for _, ver := range versions {
			defVer := schema.GroupVersionKind{
				Group:   crd.Spec.Group,
				Version: ver.Name,
				Kind:    crd.Spec.Names.Kind,
			}

			crdSchema := ver.Schema

			if crdSchema == nil {
				crdSchema = crd.Spec.Validation
			}

			if crdSchema == nil {
				return fmt.Errorf("crd did not have validation defined")
			}

			schemaValidator, _, err := validation.NewSchemaValidator(crdSchema)

			if err != nil {
				return fmt.Errorf("error on building schema validator: %v", err)
			}

			a.validators[defVer] = schemaValidator
		}
	}

	return nil
}

func (a *ResourceAsserter) LoadFromFiles(crdPaths []string) error {
	for _, crdPath := range crdPaths {
		rawCrd, err := os.ReadFile(crdPath)

		if err != nil {
			return fmt.Errorf("error reading %v CRD: %v", crdPath, err)
		}

		decodedCrd := &unstructured.Unstructured{}
		err = kubeyaml.Unmarshal(rawCrd, &decodedCrd)

		if err != nil {
			return fmt.Errorf("error decoding %v CRD: %v", crdPath, err)
		}

		crd, err := a.mapCRD(decodedCrd)

		if err != nil {
			return err
		}

		err = a.Load([]*apiextensions.CustomResourceDefinition{crd})

		if err != nil {
			return fmt.Errorf("error on %v CRD: %v", crdPath, err)
		}
	}

	return nil
}

func (a *ResourceAsserter) Assert(t TestingT, resource *unstructured.Unstructured) error {
	defVer := resource.GroupVersionKind()

	if schemaValidator, ok := a.validators[defVer]; ok {
		results := (*schemaValidator).Validate(resource)

		if len(results.Errors) == 0 {
			return nil
		}

		t.Errorf(
			"%v resource failed to pass CRD validation:\n%v",
			resource.GetName(),
			a.formatErrors(results.Errors),
		)

		return nil
	}

	return fmt.Errorf("no CRD is found for the resource of %v kind. Use Load() and LoadFromFiles() methods to register the CRD", defVer)
}

func (a *ResourceAsserter) AssertFiles(t TestingT, resourcePaths []string) error {
	for _, resourcePath := range resourcePaths {
		resourceData, err := os.ReadFile(resourcePath)

		if err != nil {
			return fmt.Errorf("error reading %v resource: %v", resourcePath, err)
		}

		resource := &unstructured.Unstructured{}
		err = kubeyaml.Unmarshal(resourceData, &resource)

		if err != nil {
			return fmt.Errorf("error decoding %v resource: %v", resourcePath, err)
		}

		err = a.Assert(t, resource)

		if err != nil {
			return fmt.Errorf("error on asserting %v resource: %v", resourcePath, err)
		}
	}

	return nil
}

func (a *ResourceAsserter) mapCRD(decodedCrd *unstructured.Unstructured) (*apiextensions.CustomResourceDefinition, error) {
	crd := apiextensions.CustomResourceDefinition{}

	switch decodedCrd.GroupVersionKind() {
	case schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Version: "v1",
		Kind:    "CustomResourceDefinition",
	}:
		crdv1 := apiextensionsv1.CustomResourceDefinition{}

		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(decodedCrd.UnstructuredContent(), &crdv1); err != nil {
			return nil, err
		}

		if err := apiextensionsv1.Convert_v1_CustomResourceDefinition_To_apiextensions_CustomResourceDefinition(&crdv1, &crd, nil); err != nil {
			return nil, err
		}
	case schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Version: "v1beta1",
		Kind:    "CustomResourceDefinition",
	}:
		crdv1beta1 := apiextensionsv1beta1.CustomResourceDefinition{}

		if err := runtime.DefaultUnstructuredConverter.
			FromUnstructured(decodedCrd.UnstructuredContent(), &crdv1beta1); err != nil {
			return nil, err
		}

		if err := apiextensionsv1beta1.Convert_v1beta1_CustomResourceDefinition_To_apiextensions_CustomResourceDefinition(&crdv1beta1, &crd, nil); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown CRD type: %v", decodedCrd.GroupVersionKind())
	}

	return &crd, nil
}

func (a *ResourceAsserter) formatErrors(errors []error) string {
	errorMsg := ""

	for _, err := range errors {
		errorMsg += fmt.Sprintf("- %v\n", err.Error())
	}

	return errorMsg
}

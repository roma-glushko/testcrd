package testcrd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockReporter struct {
	errors []string
}

func (r *MockReporter) Errorf(format string, args ...interface{}) {
	r.errors = append(r.errors, fmt.Sprint(format, args))
}

func (r *MockReporter) IsFailed() bool {
	return len(r.errors) > 0
}

func TestAsserter_LoadExternalSecretsCRDFile(t *testing.T) {
	asserter := NewResourceAsserter()

	_ = asserter.LoadFromFiles([]string{
		"fixtures/externalsecrets/crd.yaml",
	})

	assert.Len(t, asserter.validators, 2)
}

func TestAsserter_AssertValidExternalSecret(t *testing.T) {
	asserter := NewResourceAsserter()

	_ = asserter.LoadFromFiles([]string{
		"fixtures/externalsecrets/crd.yaml",
	})

	_ = asserter.AssertFiles(t, []string{
		"fixtures/externalsecrets/resource.valid.1.yaml",
	})
}

func TestAsserter_AssertInvalidExternalSecret(t *testing.T) {
	mt := MockReporter{}
	asserter := NewResourceAsserter()

	_ = asserter.LoadFromFiles([]string{
		"fixtures/externalsecrets/crd.yaml",
	})

	_ = asserter.AssertFiles(&mt, []string{
		"fixtures/externalsecrets/resource.invalid.1.yaml",
	})

	assert.True(t, mt.IsFailed())
}

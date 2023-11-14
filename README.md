# testcrd

Unit test your Kubernetes Custom Resource Definitions (CRDs) against concreate resource samples
without a need for a running Kubernetes cluster.

The primary use case would be to ensure backward-compatibility of CRD changes.

## Installation

```bash
go get github.com/roma-glushko/testcrd
```

## Usage

```go
import (
	"github.com/roma-glushko/testcrd"
	"testing"
)

func TestCRD_AssertDefinitionBackwardCompatible(t *testing.T) {
	asserter := testcrd.NewResourceAsserter()

	_ = asserter.LoadFromFiles([]string{
		"crd/my_crd.yaml",
	})

	_ = asserter.AssertFiles(t, []string{
		"fixtures/my_crd/resource.sample.yaml",
	})
}
```

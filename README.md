# testcrd

Unit test your Kubernetes Custom Resource Definitions (CRDs) against concreate resource samples
without a need for a running Kubernetes cluster.

The primary use case would be to ensure backward-compatibility of CRD changes.

## Usage

```go
import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCRD_AssertDefinitionBackwardCompatible(t *testing.T) {
	asserter := NewResourceAsserter()

	_ = asserter.LoadFromFiles([]string{
		"crd/my_crd.yaml",
	})

	_ = asserter.AssertFiles(t, []string{
		"fixtures/my_crd/resource.sample.yaml",
	})
}
```

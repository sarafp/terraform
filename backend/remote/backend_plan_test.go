package remote

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform/backend"
	"github.com/hashicorp/terraform/config/module"
)

func testOperationPlan() *backend.Operation {
	return &backend.Operation{
		Type: backend.OperationTypePlan,
	}
}

func TestRemote_planBasic(t *testing.T) {
	b := testBackendDefault(t)

	mod, modCleanup := module.TestTree(t, "./test-fixtures/plan")
	defer modCleanup()

	op := testOperationPlan()
	op.Module = mod
	op.PlanRefresh = true
	op.Workspace = backend.DefaultStateName

	run, err := b.Operation(context.Background(), op)
	if err != nil {
		t.Fatalf("error starting operation: %v", err)
	}

	<-run.Done()
	if run.Err != nil {
		t.Fatalf("error running operation: %v", run.Err)
	}
}

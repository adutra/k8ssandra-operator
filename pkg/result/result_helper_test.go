package result

import (
	"errors"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMerge(t *testing.T) {
	tests := []struct {
		name     string
		input    []ReconcileResult
		expected ReconcileResult
	}{

		{"continue vs continue", []ReconcileResult{Continue(), Continue()}, Continue()},
		{"continue vs done", []ReconcileResult{Continue(), Done()}, Done()},
		{"continue vs continue and requeue", []ReconcileResult{Continue(), ContinueAndRequeue(10 * time.Second)}, ContinueAndRequeue(10 * time.Second)},
		{"continue vs complete and requeue", []ReconcileResult{Continue(), CompleteAndRequeue(10 * time.Second)}, CompleteAndRequeue(10 * time.Second)},
		{"continue vs continue with error", []ReconcileResult{Continue(), ContinueWithError(errors.New("ouch"))}, ContinueWithError(errors.New("ouch"))},
		{"continue vs complete with error", []ReconcileResult{Continue(), CompleteWithError(errors.New("ouch"))}, CompleteWithError(errors.New("ouch"))},

		{"done vs continue", []ReconcileResult{Done(), Continue()}, Done()},
		{"done vs done", []ReconcileResult{Done(), Done()}, Done()},
		{"done vs continue and requeue", []ReconcileResult{Done(), ContinueAndRequeue(10 * time.Second)}, CompleteAndRequeue(10 * time.Second)},
		{"done vs complete and requeue", []ReconcileResult{Done(), CompleteAndRequeue(10 * time.Second)}, CompleteAndRequeue(10 * time.Second)},
		{"done vs continue with error", []ReconcileResult{Done(), ContinueWithError(errors.New("ouch"))}, CompleteWithError(errors.New("ouch"))},
		{"done vs complete with error", []ReconcileResult{Done(), CompleteWithError(errors.New("ouch"))}, CompleteWithError(errors.New("ouch"))},

		{"continue and requeue vs continue", []ReconcileResult{ContinueAndRequeue(10 * time.Second), Continue()}, ContinueAndRequeue(10 * time.Second)},
		{"continue and requeue vs done", []ReconcileResult{ContinueAndRequeue(10 * time.Second), Done()}, CompleteAndRequeue(10 * time.Second)},
		{"continue and requeue vs continue and requeue", []ReconcileResult{ContinueAndRequeue(10 * time.Second), ContinueAndRequeue(5 * time.Second)}, ContinueAndRequeue(5 * time.Second)},
		{"continue and requeue vs complete and requeue", []ReconcileResult{ContinueAndRequeue(10 * time.Second), CompleteAndRequeue(5 * time.Second)}, CompleteAndRequeue(5 * time.Second)},
		{"continue and requeue vs continue with error", []ReconcileResult{ContinueAndRequeue(10 * time.Second), ContinueWithError(errors.New("ouch"))}, ContinueWithError(errors.New("ouch"))},
		{"continue and requeue vs complete with error", []ReconcileResult{ContinueAndRequeue(10 * time.Second), CompleteWithError(errors.New("ouch"))}, CompleteWithError(errors.New("ouch"))},

		{"complete and requeue vs continue", []ReconcileResult{CompleteAndRequeue(10 * time.Second), Continue()}, CompleteAndRequeue(10 * time.Second)},
		{"complete and requeue vs done", []ReconcileResult{CompleteAndRequeue(10 * time.Second), Done()}, CompleteAndRequeue(10 * time.Second)},
		{"complete and requeue vs continue and requeue", []ReconcileResult{CompleteAndRequeue(10 * time.Second), ContinueAndRequeue(5 * time.Second)}, CompleteAndRequeue(5 * time.Second)},
		{"complete and requeue vs complete and requeue", []ReconcileResult{CompleteAndRequeue(10 * time.Second), CompleteAndRequeue(5 * time.Second)}, CompleteAndRequeue(5 * time.Second)},
		{"complete and requeue vs continue with error", []ReconcileResult{CompleteAndRequeue(10 * time.Second), ContinueWithError(errors.New("ouch"))}, CompleteWithError(errors.New("ouch"))},
		{"complete and requeue vs complete with error", []ReconcileResult{CompleteAndRequeue(10 * time.Second), CompleteWithError(errors.New("ouch"))}, CompleteWithError(errors.New("ouch"))},

		{"continue with error vs continue", []ReconcileResult{ContinueWithError(errors.New("ouch")), Continue()}, ContinueWithError(errors.New("ouch"))},
		{"continue with error vs done", []ReconcileResult{ContinueWithError(errors.New("ouch")), Done()}, CompleteWithError(errors.New("ouch"))},
		{"continue with error vs continue and requeue", []ReconcileResult{ContinueWithError(errors.New("ouch")), ContinueAndRequeue(5 * time.Second)}, ContinueWithError(errors.New("ouch"))},
		{"continue with error vs complete and requeue", []ReconcileResult{ContinueWithError(errors.New("ouch")), CompleteAndRequeue(5 * time.Second)}, CompleteWithError(errors.New("ouch"))},
		{"continue with error vs continue with error", []ReconcileResult{ContinueWithError(errors.New("ouch")), ContinueWithError(errors.New("ouch"))}, ContinueWithError(multierror.Append(errors.New("ouch"), errors.New("ouch")))},
		{"continue with error vs complete with error", []ReconcileResult{ContinueWithError(errors.New("ouch")), CompleteWithError(errors.New("ouch"))}, CompleteWithError(multierror.Append(errors.New("ouch"), errors.New("ouch")))},

		{"complete with error vs continue", []ReconcileResult{CompleteWithError(errors.New("ouch")), Continue()}, CompleteWithError(errors.New("ouch"))},
		{"complete with error vs done", []ReconcileResult{CompleteWithError(errors.New("ouch")), Done()}, CompleteWithError(errors.New("ouch"))},
		{"complete with error vs continue and requeue", []ReconcileResult{CompleteWithError(errors.New("ouch")), ContinueAndRequeue(5 * time.Second)}, CompleteWithError(errors.New("ouch"))},
		{"complete with error vs complete and requeue", []ReconcileResult{CompleteWithError(errors.New("ouch")), CompleteAndRequeue(5 * time.Second)}, CompleteWithError(errors.New("ouch"))},
		{"complete with error vs continue with error", []ReconcileResult{CompleteWithError(errors.New("ouch")), ContinueWithError(errors.New("ouch"))}, CompleteWithError(multierror.Append(errors.New("ouch"), errors.New("ouch")))},
		{"complete with error vs complete with error", []ReconcileResult{CompleteWithError(errors.New("ouch")), CompleteWithError(errors.New("ouch"))}, CompleteWithError(multierror.Append(errors.New("ouch"), errors.New("ouch")))},

		{
			"N items",
			[]ReconcileResult{
				Continue(),
				Continue(),
				Done(),
				Done(),
			},
			Done(),
		},
		{
			"N items with requeue",
			[]ReconcileResult{
				Continue(),
				ContinueAndRequeue(5 * time.Second),
				ContinueAndRequeue(10 * time.Second),
				ContinueAndRequeue(15 * time.Second),
				ContinueAndRequeue(20 * time.Second),
				Done(),
			},
			CompleteAndRequeue(5 * time.Second),
		},
		{
			"N items with error",
			[]ReconcileResult{
				Continue(),
				ContinueAndRequeue(5 * time.Second),
				ContinueAndRequeue(10 * time.Second),
				ContinueWithError(errors.New("ouch 1")),
				ContinueWithError(errors.New("ouch 2")),
				ContinueWithError(errors.New("ouch 3")),
				Done(),
			},
			CompleteWithError(multierror.Append(errors.New("ouch 1"), errors.New("ouch 2"), errors.New("ouch 3"))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Merge(tt.input...)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

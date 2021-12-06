package result

import (
	"github.com/hashicorp/go-multierror"
	"math"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

// Copyright DataStax, Inc.
// Please see the included license file for details.

type ReconcileResult struct {
	completed bool
	err       error
	delay     time.Duration
}

func (r ReconcileResult) Completed() bool {
	return r.completed
}

func (r ReconcileResult) Output() (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: r.delay}, r.err
}

func Continue() ReconcileResult {
	return ReconcileResult{}
}

func Done() ReconcileResult {
	return ReconcileResult{completed: true}
}

func CompleteAndRequeue(delay time.Duration) ReconcileResult {
	return ReconcileResult{delay: delay, completed: true}
}

func ContinueAndRequeue(delay time.Duration) ReconcileResult {
	return ReconcileResult{delay: delay, completed: false}
}

func CompleteWithError(e error) ReconcileResult {
	return ReconcileResult{err: e, completed: true}
}

func ContinueWithError(e error) ReconcileResult {
	return ReconcileResult{err: e, completed: false}
}

func Merge(results ...ReconcileResult) ReconcileResult {
	var completed bool
	var delay time.Duration
	var err error
	for _, result := range results {
		// the merged result is completed if any of the results is completed
		completed = completed || result.completed
		// the requeue delay is the smallest one, excluding zero delays
		if result.delay > 0 {
			if delay == 0 {
				delay = result.delay
			} else {
				delay = time.Duration(math.Min(float64(delay), float64(result.delay)))
			}
		}
		if result.err != nil {
			if err == nil {
				err = result.err
			} else {
				err = multierror.Append(err, result.err)
			}
		}
	}
	if err != nil {
		// if the result is an error, no need to retain the delay, it's going to be requeued immediately when complete.
		delay = 0
	}
	return ReconcileResult{completed: completed, delay: delay, err: err}
}

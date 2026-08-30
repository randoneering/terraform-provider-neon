package provider

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	neon "github.com/kislerdm/neon-sdk-go"
)

// delay wraps a Neon API call with retry-on-transient-error semantics.
//
// This mirrors provider/retry.go (the SDK v2 version) but adapted for the
// Plugin Framework: the wrapped function only needs a context, and the
// result is a diag.Diagnostics that the resource methods append to their
// response. ponytail: same shape as SDK v2, ported not redesigned.
type delay struct {
	delay  time.Duration
	maxCnt uint8
}

// Retry runs fn, retrying on transient Neon errors (429, 500, 423) up to
// maxCnt times. Non-neon errors and non-transient neon errors are returned
// immediately.
func (r *delay) Retry(fn func(ctx context.Context) error, ctx context.Context) diag.Diagnostics {
	var lastErr error
	for i := uint8(0); i < r.maxCnt; i++ {
		tflog.Debug(ctx, "API call attempt", map[string]interface{}{"i": int(i)})
		err := fn(ctx)
		if err == nil {
			return nil
		}
		switch e := err.(type) {
		case neon.Error:
			tflog.Debug(ctx, "API call error code", map[string]interface{}{"code": e.HTTPCode})
			switch e.HTTPCode {
			case http.StatusOK:
				return nil
			case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusLocked:
				lastErr = e
				time.Sleep(r.delay)
				continue
			default:
				return errorDiagnostic(e)
			}
		default:
			return errorDiagnostic(e)
		}
	}
	return errorDiagnostic(lastErr)
}

func errorDiagnostic(err error) diag.Diagnostics {
	if err == nil {
		return nil
	}
	return diag.Diagnostics{diag.NewErrorDiagnostic("Neon API call failed", err.Error())}
}

// apiKeyReadiness is the retry policy for API key operations. Matches
// projectReadiness in provider/retry.go to keep behaviour consistent
// across the migration window.
var apiKeyReadiness = delay{
	delay:  1 * time.Second,
	maxCnt: 120,
}

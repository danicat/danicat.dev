package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/status"
	mwexp "github.com/firebase/genkit/go/plugins/middleware/exp"
)

const (
	sessionBudgetPublicMessage = "This session reached its token budget. Click New chat to start a new session."
	runBudgetPublicMessage     = "That question used too many tokens in a single turn. Try a narrower question, or click New chat to start a new session."
	maxTurnsPublicMessage      = "That question needed too many lookups in a single turn. Try asking about a specific article or topic."
)

// publicBudgetMiddleware wraps Genkit's experimental Budget middleware so that
// budget stops carry a safe public error message across genkit.Handler in
// production instead of being redacted to generic "aborted", while keeping raw
// token counts in server logs.
type publicBudgetMiddleware struct {
	budget *mwexp.Budget
}

func newPublicBudgetMiddleware(budget *mwexp.Budget) ai.Middleware {
	return &publicBudgetMiddleware{budget: budget}
}

func (m *publicBudgetMiddleware) Name() string {
	if m.budget == nil {
		return "budget"
	}
	return m.budget.Name()
}

func (m *publicBudgetMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	if m.budget == nil {
		return nil, status.Errorf(status.ErrInvalidArgument, "publicBudgetMiddleware: budget is required")
	}
	hooks, err := m.budget.New(ctx)
	if err != nil {
		return nil, err
	}
	if hooks == nil || hooks.WrapGenerate == nil {
		return hooks, nil
	}
	scope := m.budget.Scope
	if scope == "" {
		scope = mwexp.BudgetScopeRun
	}
	h := &publicBudgetHook{
		scope: scope,
		inner: hooks.WrapGenerate,
	}
	hooks.WrapGenerate = h.wrapGenerate
	return hooks, nil
}

type publicBudgetHook struct {
	scope mwexp.BudgetScope
	inner func(context.Context, *ai.GenerateParams, ai.GenerateNext) (*ai.ModelResponse, error)
}

func (h *publicBudgetHook) wrapGenerate(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
	nextCalled := false
	resp, err := h.inner(ctx, params, func(ctx context.Context, p *ai.GenerateParams) (*ai.ModelResponse, error) {
		nextCalled = true
		return next(ctx, p)
	})
	if err != nil && !nextCalled && errors.Is(err, ai.ErrBudgetExceeded) {
		slog.Warn("budget_exceeded", "scope", h.scope, "detail", err.Error())
		msg := runBudgetPublicMessage
		if h.scope == mwexp.BudgetScopeSession {
			msg = sessionBudgetPublicMessage
		}
		pubErr := status.PublicErrorf(ai.ErrBudgetExceeded, "%s", msg).WithCause(err)
		if resp != nil {
			resp.Error = status.Convert(pubErr)
		}
		return resp, pubErr
	}
	return resp, err
}

// publicMaxTurnsMiddleware marks ai.ErrMaxTurnsExceeded public so clients can
// distinguish tool-loop turn limits from token budget exhaustion in production.
type publicMaxTurnsMiddleware struct{}

func newPublicMaxTurnsMiddleware() ai.Middleware {
	return &publicMaxTurnsMiddleware{}
}

func (*publicMaxTurnsMiddleware) Name() string {
	return "public-max-turns"
}

func (m *publicMaxTurnsMiddleware) New(context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		WrapGenerate: m.wrapGenerate,
	}, nil
}

func (*publicMaxTurnsMiddleware) wrapGenerate(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
	resp, err := next(ctx, params)
	if err != nil && errors.Is(err, ai.ErrMaxTurnsExceeded) {
		if _, isPublic := status.PublicMessage(err); !isPublic {
			slog.Warn("max_turns_exceeded", "detail", err.Error())
			pubErr := status.PublicErrorf(ai.ErrMaxTurnsExceeded, "%s", maxTurnsPublicMessage).WithCause(err)
			if resp != nil {
				resp.Error = status.Convert(pubErr)
			}
			return resp, pubErr
		}
	}
	return resp, err
}

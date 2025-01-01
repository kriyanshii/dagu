// Copyright (C) 2024 Yota Hamada
// SPDX-License-Identifier: GPL-3.0-or-later

package digraph

import (
	"context"

	"github.com/dagu-org/dagu/internal/cmdutil"
	"github.com/dagu-org/dagu/internal/mailer"
)

type StepContext struct {
	Context

	outputVariables *SyncMap
	step            Step
	envs            map[string]string
}

func NewStepContext(ctx context.Context, step Step) StepContext {
	return StepContext{
		Context: GetContext(ctx),

		outputVariables: &SyncMap{},
		step:            step,
		envs: map[string]string{
			EnvKeyDAGStepName: step.Name,
		},
	}
}

func (c StepContext) AllEnvs() []string {
	envs := c.Context.AllEnvs()
	for k, v := range c.envs {
		envs = append(envs, k+"="+v)
	}
	c.outputVariables.Range(func(_, value any) bool {
		envs = append(envs, value.(string))
		return true
	})
	return envs
}

func (c StepContext) LoadOutputVariables(vars *SyncMap) {
	vars.Range(func(key, value any) bool {
		// Skip if the key already exists
		if _, ok := c.outputVariables.Load(key); ok {
			return true
		}
		c.outputVariables.Store(key, value)
		return true
	})
}

func (c StepContext) MailerConfig() (mailer.Config, error) {
	return EvalStringFields(c, mailer.Config{
		Host:     c.dag.SMTP.Host,
		Port:     c.dag.SMTP.Port,
		Username: c.dag.SMTP.Username,
		Password: c.dag.SMTP.Password,
	})
}

func (c StepContext) EvalString(s string) (string, error) {
	return cmdutil.EvalString(s,
		cmdutil.WithVariables(c.envs),
		cmdutil.WithVariables(c.outputVariables.Variables()),
	)
}

func (c StepContext) WithEnv(key, value string) StepContext {
	c.envs[key] = value
	return c
}

func WithStepContext(ctx context.Context, stepContext StepContext) context.Context {
	return context.WithValue(ctx, stepCtxKey{}, stepContext)
}

func GetStepContext(ctx context.Context) StepContext {
	contextValue, ok := ctx.Value(stepCtxKey{}).(StepContext)
	if !ok {
		return NewStepContext(ctx, Step{})
	}
	return contextValue
}

func IsStepContext(ctx context.Context) bool {
	_, ok := ctx.Value(stepCtxKey{}).(StepContext)
	return ok
}

type stepCtxKey struct{}

func EvalStringFields[T any](stepContext StepContext, obj T) (T, error) {
	return cmdutil.EvalStringFields(obj,
		cmdutil.WithVariables(stepContext.outputVariables.Variables()))
}
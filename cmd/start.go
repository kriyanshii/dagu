// Copyright (C) 2024 Yota Hamada
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dagu-org/dagu/internal/agent"
	"github.com/dagu-org/dagu/internal/config"
	"github.com/dagu-org/dagu/internal/digraph"
	"github.com/dagu-org/dagu/internal/logger"
	"github.com/spf13/cobra"
)

const startPrefix = "start_"

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [flags] /path/to/spec.yaml",
		Short: "Runs the DAG",
		Long:  `dagu start [--params="param1 param2"] /path/to/spec.yaml`,
		Args:  cobra.MinimumNArgs(1),
		RunE:  wrapRunE(runStart),
	}

	initStartFlags(cmd)
	return cmd
}

func initStartFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("params", "p", "", "parameters")
	cmd.Flags().StringP("requestID", "r", "", "specify request ID")
	cmd.Flags().BoolP("quiet", "q", false, "suppress output")
}

func runStart(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	setup := newSetup(cfg)

	// Get quiet flag
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return fmt.Errorf("failed to get quiet flag: %w", err)
	}

	// Get request ID if specified
	requestID, err := cmd.Flags().GetString("requestID")
	if err != nil {
		return fmt.Errorf("failed to get request ID: %w", err)
	}

	ctx := setup.loggerContext(cmd.Context(), quiet)

	// Get parameters
	// Get parameters from the new syntax
	var params string
	if argsLenAtDash := cmd.ArgsLenAtDash(); argsLenAtDash != -1 {
		params = strings.Join(args[argsLenAtDash:], " ")
		args = args[:1]
	} else {
		// Get parameters from the deprecated flag
		params, err = cmd.Flags().GetString("params")
		if err != nil {
			logger.Error(ctx, "Failed to get parameters", "err", err)
			return fmt.Errorf("failed to get parameters: %w", err)
		}
	}

	// Initialize and run DAG
	return executeDag(ctx, setup, args[0], removeQuotes(params), quiet, requestID)
}

func executeDag(ctx context.Context, setup *setup, specPath, params string, quiet bool, requestID string) error {
	// Load DAG
	dag, err := digraph.Load(ctx, setup.cfg.Paths.BaseConfig, specPath, params)
	if err != nil {
		logger.Error(ctx, "Failed to load DAG", "path", specPath, "err", err)
		return fmt.Errorf("failed to load DAG from %s: %w", specPath, err)
	}

	// Generate request ID
	if requestID == "" {
		var err error
		requestID, err = generateRequestID()
		if err != nil {
			logger.Error(ctx, "Failed to generate request ID", "err", err)
			return fmt.Errorf("failed to generate request ID: %w", err)
		}
	}

	// Setup logging
	logFile, err := setup.openLogFile(startPrefix, dag, requestID)
	if err != nil {
		logger.Error(ctx, "failed to initialize log file", "DAG", dag.Name, "err", err)
		return fmt.Errorf("failed to initialize log file for DAG %s: %w", dag.Name, err)
	}
	defer logFile.Close()

	ctx = setup.loggerContextWithFile(ctx, quiet, logFile)

	logger.Info(ctx, "DAG execution initiated", "DAG", dag.Name, "requestID", requestID, "logFile", logFile.Name())

	dagStore, err := setup.dagStore()
	if err != nil {
		logger.Error(ctx, "Failed to initialize DAG store", "err", err)
		return fmt.Errorf("failed to initialize DAG store: %w", err)
	}

	cli, err := setup.client()
	if err != nil {
		logger.Error(ctx, "Failed to initialize client", "err", err)
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	// Create and run agent
	agt := agent.New(
		requestID,
		dag,
		filepath.Dir(logFile.Name()),
		logFile.Name(),
		cli,
		dagStore,
		setup.historyStore(),
		&agent.Options{},
	)

	listenSignals(ctx, agt)

	if err := agt.Run(ctx); err != nil {
		logger.Error(ctx, "Failed to execute DAG", "DAG", dag.Name, "requestID", requestID, "err", err)

		if quiet {
			os.Exit(1)
		} else {
			agt.PrintSummary(ctx)
			return fmt.Errorf("failed to execute DAG %s (requestID: %s): %w", dag.Name, requestID, err)
		}
	}

	return nil
}

// removeQuotes removes the surrounding quotes from the string.
func removeQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// © 2022 Nokia.
//
// This code is a Contribution to the gNMIc project ("Work") made under the Google Software Grant and Corporate Contributor License Agreement ("CLA") and governed by the Apache License 2.0.
// No other rights or licenses in or to any of Nokia's intellectual property are granted for any other purpose.
// This code is provided on an "as is" basis without any warranties of any kind.
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"sync"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/openconfig/gnmic/pkg/api/types"
	"github.com/openconfig/gnmic/pkg/formatters"
	"github.com/openconfig/gnmic/pkg/outputs"
)

var (
	getRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "gnmic",
		Subsystem: "get_poller",
		Name:      "requests_total",
		Help:      "Total number of Get requests sent",
	}, []string{"target", "get_name"})

	getRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "gnmic",
		Subsystem: "get_poller",
		Name:      "request_duration_seconds",
		Help:      "Duration of Get requests in seconds",
		Buckets:   prometheus.DefBuckets,
	}, []string{"target", "get_name"})

	getRequestErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "gnmic",
		Subsystem: "get_poller",
		Name:      "request_errors_total",
		Help:      "Total number of Get request errors",
	}, []string{"target", "get_name"})

	getResponseNotifications = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "gnmic",
		Subsystem: "get_poller",
		Name:      "response_notifications_total",
		Help:      "Total number of notifications received from Get responses",
	}, []string{"target", "get_name"})
)

func init() {
	prometheus.MustRegister(getRequestsTotal)
	prometheus.MustRegister(getRequestDuration)
	prometheus.MustRegister(getRequestErrors)
	prometheus.MustRegister(getResponseNotifications)
}

// StartGetPoller starts the periodic Get request poller for all configured gets.
// It creates a goroutine for each target/get combination that polls at the configured interval.
func (a *App) StartGetPoller(ctx context.Context) {
	if len(a.Config.Gets) == 0 {
		if a.Config.Debug {
			a.Logger.Printf("no get configurations found, skipping get poller")
		}
		return
	}

	a.Logger.Printf("starting get poller with %d get configurations", len(a.Config.Gets))

	for getName, gc := range a.Config.Gets {
		targets := a.getTargetsForGet(gc)
		if len(targets) == 0 {
			a.Logger.Printf("get %q: no matching targets found", getName)
			continue
		}

		for _, tc := range targets {
			a.wg.Add(1)
			go a.runGetPollLoop(ctx, tc, gc)
		}
	}
}

// getTargetsForGet returns the list of targets that a GetConfig applies to.
// If GetConfig.Targets is specified, only those targets are returned.
// Otherwise, all configured targets are returned.
func (a *App) getTargetsForGet(gc *types.GetConfig) []*types.TargetConfig {
	a.configLock.RLock()
	defer a.configLock.RUnlock()

	if len(gc.Targets) > 0 {
		targets := make([]*types.TargetConfig, 0, len(gc.Targets))
		for _, targetName := range gc.Targets {
			if tc, ok := a.Config.Targets[targetName]; ok {
				targets = append(targets, tc)
			} else {
				a.Logger.Printf("get %q: target %q not found in configuration", gc.Name, targetName)
			}
		}
		return targets
	}

	// Return all targets if no specific targets are configured
	targets := make([]*types.TargetConfig, 0, len(a.Config.Targets))
	for _, tc := range a.Config.Targets {
		targets = append(targets, tc)
	}
	return targets
}

// runGetPollLoop runs the polling loop for a single target/get combination.
// It executes Get requests at the configured interval and exports results to outputs.
func (a *App) runGetPollLoop(ctx context.Context, tc *types.TargetConfig, gc *types.GetConfig) {
	defer a.wg.Done()

	a.Logger.Printf("starting get poller for target %q, get %q with interval %s",
		tc.Name, gc.Name, gc.Interval)

	ticker := time.NewTicker(gc.Interval)
	defer ticker.Stop()

	// Execute immediately on start
	a.executeAndExportGet(ctx, tc, gc)

	for {
		select {
		case <-ticker.C:
			a.executeAndExportGet(ctx, tc, gc)
		case <-ctx.Done():
			a.Logger.Printf("stopping get poller for target %q, get %q", tc.Name, gc.Name)
			return
		}
	}
}

// executeAndExportGet executes a single Get request and exports the result to outputs.
func (a *App) executeAndExportGet(ctx context.Context, tc *types.TargetConfig, gc *types.GetConfig) {
	if a.Config.Debug {
		a.Logger.Printf("executing get %q for target %q", gc.Name, tc.Name)
	}

	// Create the Get request
	req, err := a.Config.CreateGetRequestFromConfig(gc, tc)
	if err != nil {
		a.Logger.Printf("get %q: failed to create request for target %q: %v", gc.Name, tc.Name, err)
		getRequestErrors.WithLabelValues(tc.Name, gc.Name).Inc()
		return
	}

	// Execute the Get request with timing
	startTime := time.Now()
	getRequestsTotal.WithLabelValues(tc.Name, gc.Name).Inc()

	rsp, err := a.ClientGet(ctx, tc, req)
	duration := time.Since(startTime)
	getRequestDuration.WithLabelValues(tc.Name, gc.Name).Observe(duration.Seconds())

	if err != nil {
		a.Logger.Printf("get %q: request failed for target %q: %v", gc.Name, tc.Name, err)
		getRequestErrors.WithLabelValues(tc.Name, gc.Name).Inc()
		return
	}

	if rsp == nil {
		if a.Config.Debug {
			a.Logger.Printf("get %q: nil response from target %q", gc.Name, tc.Name)
		}
		return
	}

	// Track notifications
	notifCount := formatters.GetResponseNotificationCount(rsp)
	getResponseNotifications.WithLabelValues(tc.Name, gc.Name).Add(float64(notifCount))

	if a.Config.Debug {
		a.Logger.Printf("get %q: received %d notifications from target %q in %s",
			gc.Name, notifCount, tc.Name, duration)
	}

	// Export the response to outputs
	a.ExportGetResponse(ctx, rsp, tc, gc)
}

// ExportGetResponse exports a GetResponse to the configured outputs.
// It converts the GetResponse to SubscribeResponse format for compatibility with existing outputs.
func (a *App) ExportGetResponse(ctx context.Context, rsp *gnmi.GetResponse, tc *types.TargetConfig, gc *types.GetConfig) {
	if rsp == nil {
		return
	}

	// Convert GetResponse notifications to SubscribeResponse format
	subscribeResponses := formatters.GetResponseToSubscribeResponses(rsp)
	if len(subscribeResponses) == 0 {
		return
	}

	// Build metadata
	meta := outputs.Meta{
		"source":   tc.Name,
		"format":   a.Config.Format,
		"get-name": gc.Name,
	}
	for k, v := range tc.EventTags {
		meta[k] = v
	}

	// Determine which outputs to use
	var outs []string
	if len(gc.Outputs) > 0 {
		outs = gc.Outputs
	} else {
		outs = tc.Outputs
	}

	// Export each notification as a SubscribeResponse
	wg := new(sync.WaitGroup)
	for _, sr := range subscribeResponses {
		a.exportSingleResponse(ctx, sr, meta, outs, wg)
	}
	wg.Wait()
}

// exportSingleResponse exports a single SubscribeResponse to outputs.
func (a *App) exportSingleResponse(ctx context.Context, rsp *gnmi.SubscribeResponse, meta outputs.Meta, outs []string, wg *sync.WaitGroup) {
	// No specific outputs defined - write to all outputs
	if len(outs) == 0 {
		a.operLock.RLock()
		for _, o := range a.Outputs {
			wg.Add(1)
			go func(o outputs.Output) {
				defer wg.Done()
				o.Write(ctx, rsp, meta)
			}(o)
		}
		a.operLock.RUnlock()
		return
	}

	// Write to specified outputs only
	a.operLock.RLock()
	for _, name := range outs {
		if o, ok := a.Outputs[name]; ok {
			wg.Add(1)
			go func(o outputs.Output) {
				defer wg.Done()
				o.Write(ctx, rsp, meta)
			}(o)
		}
	}
	a.operLock.RUnlock()
}

// LoadGetsConfig loads the Gets configuration from the config file.
// This should be called during initialization.
func (a *App) LoadGetsConfig() error {
	_, err := a.Config.GetGets()
	if err != nil {
		return err
	}
	return nil
}

// StartGetPollerForTarget starts GET polling for a specific target.
// This is used when tunnel targets connect dynamically after startup.
func (a *App) StartGetPollerForTarget(ctx context.Context, tc *types.TargetConfig) {
	if len(a.Config.Gets) == 0 {
		return
	}

	// If target has specific gets configured, only run those
	if len(tc.Gets) > 0 {
		for _, getName := range tc.Gets {
			if gc, ok := a.Config.Gets[getName]; ok {
				a.wg.Add(1)
				go a.runGetPollLoop(ctx, tc, gc)
			} else {
				a.Logger.Printf("get %q not found in configuration for target %q", getName, tc.Name)
			}
		}
		return
	}

	// Otherwise, check all gets to see if this target matches
	for _, gc := range a.Config.Gets {
		if len(gc.Targets) > 0 {
			// GetConfig has specific targets, check if this target is included
			for _, targetName := range gc.Targets {
				if targetName == tc.Name {
					a.wg.Add(1)
					go a.runGetPollLoop(ctx, tc, gc)
					break
				}
			}
		}
		// If GetConfig has no specific targets, don't auto-apply to tunnel targets
		// (they should explicitly list gets in their config)
	}
}

// © 2022 Nokia.
//
// This code is a Contribution to the gNMIc project ("Work") made under the Google Software Grant and Corporate Contributor License Agreement ("CLA") and governed by the Apache License 2.0.
// No other rights or licenses in or to any of Nokia's intellectual property are granted for any other purpose.
// This code is provided on an "as is" basis without any warranties of any kind.
//
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GetConfig defines a periodic gNMI Get request configuration.
// Similar to SubscriptionConfig but for polling Get requests at regular intervals.
type GetConfig struct {
	Name            string        `mapstructure:"name,omitempty" json:"name,omitempty"`
	Paths           []string      `mapstructure:"paths,omitempty" json:"paths,omitempty"`
	Prefix          string        `mapstructure:"prefix,omitempty" json:"prefix,omitempty"`
	Target          string        `mapstructure:"target,omitempty" json:"target,omitempty"`
	SetTarget       bool          `mapstructure:"set-target,omitempty" json:"set-target,omitempty"`
	Type            string        `mapstructure:"type,omitempty" json:"type,omitempty"`
	Encoding        *string       `mapstructure:"encoding,omitempty" json:"encoding,omitempty"`
	Models          []string      `mapstructure:"models,omitempty" json:"models,omitempty"`
	Interval        time.Duration `mapstructure:"interval,omitempty" json:"interval,omitempty"`
	Outputs         []string      `mapstructure:"outputs,omitempty" json:"outputs,omitempty"`
	Depth           uint32        `mapstructure:"depth,omitempty" json:"depth,omitempty"`
	EventProcessors []string      `mapstructure:"event-processors,omitempty" json:"event-processors,omitempty"`
	// Targets allows specifying which targets this get config applies to.
	// If empty, applies to all configured targets.
	Targets []string `mapstructure:"targets,omitempty" json:"targets,omitempty"`
}

// String returns a JSON representation of the GetConfig.
func (gc *GetConfig) String() string {
	b, err := json.Marshal(gc)
	if err != nil {
		return ""
	}
	return string(b)
}

// PathsString returns a formatted string of the paths.
func (gc *GetConfig) PathsString() string {
	return fmt.Sprintf("- %s", strings.Join(gc.Paths, "\n- "))
}

// PrefixString returns the prefix or "NA" if empty.
func (gc *GetConfig) PrefixString() string {
	if gc.Prefix == "" {
		return "NA"
	}
	return gc.Prefix
}

// TypeString returns the data type or "ALL" if empty.
func (gc *GetConfig) TypeString() string {
	if gc.Type == "" {
		return "ALL"
	}
	return strings.ToUpper(gc.Type)
}

// IntervalString returns the interval as a string.
func (gc *GetConfig) IntervalString() string {
	return gc.Interval.String()
}

// ModelsString returns a formatted string of the models.
func (gc *GetConfig) ModelsString() string {
	if len(gc.Models) == 0 {
		return "NA"
	}
	return fmt.Sprintf("- %s", strings.Join(gc.Models, "\n- "))
}

// Clone returns a deep copy of the GetConfig.
func (gc *GetConfig) Clone() *GetConfig {
	if gc == nil {
		return nil
	}
	clone := &GetConfig{
		Name:      gc.Name,
		Prefix:    gc.Prefix,
		Target:    gc.Target,
		SetTarget: gc.SetTarget,
		Type:      gc.Type,
		Interval:  gc.Interval,
		Depth:     gc.Depth,
	}
	if gc.Paths != nil {
		clone.Paths = make([]string, len(gc.Paths))
		copy(clone.Paths, gc.Paths)
	}
	if gc.Encoding != nil {
		enc := *gc.Encoding
		clone.Encoding = &enc
	}
	if gc.Models != nil {
		clone.Models = make([]string, len(gc.Models))
		copy(clone.Models, gc.Models)
	}
	if gc.Outputs != nil {
		clone.Outputs = make([]string, len(gc.Outputs))
		copy(clone.Outputs, gc.Outputs)
	}
	if gc.EventProcessors != nil {
		clone.EventProcessors = make([]string, len(gc.EventProcessors))
		copy(clone.EventProcessors, gc.EventProcessors)
	}
	if gc.Targets != nil {
		clone.Targets = make([]string, len(gc.Targets))
		copy(clone.Targets, gc.Targets)
	}
	return clone
}

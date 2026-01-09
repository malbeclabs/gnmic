// © 2022 Nokia.
//
// This code is a Contribution to the gNMIc project ("Work") made under the Google Software Grant and Corporate Contributor License Agreement ("CLA") and governed by the Apache License 2.0.
// No other rights or licenses in or to any of Nokia's intellectual property are granted for any other purpose.
// This code is provided on an "as is" basis without any warranties of any kind.
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/protobuf/proto"

	"github.com/openconfig/gnmic/pkg/api/types"
)

func TestValidateGetConfig(t *testing.T) {
	tests := map[string]struct {
		name    string
		gc      *types.GetConfig
		wantErr bool
	}{
		"valid_config": {
			name: "valid",
			gc: &types.GetConfig{
				Paths:    []string{"/interfaces/interface"},
				Interval: 30 * time.Second,
				Type:     "ALL",
			},
			wantErr: false,
		},
		"missing_paths": {
			name: "missing_paths",
			gc: &types.GetConfig{
				Interval: 30 * time.Second,
			},
			wantErr: true,
		},
		"zero_interval": {
			name: "zero_interval",
			gc: &types.GetConfig{
				Paths:    []string{"/interfaces"},
				Interval: 0,
			},
			wantErr: true,
		},
		"negative_interval": {
			name: "negative_interval",
			gc: &types.GetConfig{
				Paths:    []string{"/interfaces"},
				Interval: -1 * time.Second,
			},
			wantErr: true,
		},
		"invalid_type": {
			name: "invalid_type",
			gc: &types.GetConfig{
				Paths:    []string{"/interfaces"},
				Interval: 30 * time.Second,
				Type:     "INVALID",
			},
			wantErr: true,
		},
		"valid_config_type": {
			name: "valid_config_type",
			gc: &types.GetConfig{
				Paths:    []string{"/interfaces"},
				Interval: 30 * time.Second,
				Type:     "CONFIG",
			},
			wantErr: false,
		},
		"valid_state_type": {
			name: "valid_state_type",
			gc: &types.GetConfig{
				Paths:    []string{"/interfaces"},
				Interval: 30 * time.Second,
				Type:     "STATE",
			},
			wantErr: false,
		},
		"valid_operational_type": {
			name: "valid_operational_type",
			gc: &types.GetConfig{
				Paths:    []string{"/interfaces"},
				Interval: 30 * time.Second,
				Type:     "OPERATIONAL",
			},
			wantErr: false,
		},
		"valid_with_prefix_only": {
			name: "valid_with_prefix_only",
			gc: &types.GetConfig{
				Prefix:   "/interfaces",
				Interval: 30 * time.Second,
			},
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateGetConfig(tt.name, tt.gc)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGetConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateGetRequestFromConfig(t *testing.T) {
	encoding := "json"

	tests := map[string]struct {
		gc      *types.GetConfig
		tc      *types.TargetConfig
		want    *gnmi.GetRequest
		wantErr bool
	}{
		"basic_get_request": {
			gc: &types.GetConfig{
				Name:     "test_get",
				Paths:    []string{"/interfaces/interface"},
				Encoding: &encoding,
				Interval: 30 * time.Second,
			},
			tc: &types.TargetConfig{
				Name: "target1",
			},
			want: &gnmi.GetRequest{
				Path: []*gnmi.Path{
					{
						Elem: []*gnmi.PathElem{
							{Name: "interfaces"},
							{Name: "interface"},
						},
					},
				},
				Encoding: gnmi.Encoding_JSON,
			},
			wantErr: false,
		},
		"get_request_with_prefix": {
			gc: &types.GetConfig{
				Name:     "test_get_prefix",
				Prefix:   "/interfaces",
				Paths:    []string{"/interface/state"},
				Encoding: &encoding,
				Interval: 30 * time.Second,
			},
			tc: &types.TargetConfig{
				Name: "target1",
			},
			want: &gnmi.GetRequest{
				Prefix: &gnmi.Path{
					Elem: []*gnmi.PathElem{
						{Name: "interfaces"},
					},
				},
				Path: []*gnmi.Path{
					{
						Elem: []*gnmi.PathElem{
							{Name: "interface"},
							{Name: "state"},
						},
					},
				},
				Encoding: gnmi.Encoding_JSON,
			},
			wantErr: false,
		},
		"get_request_with_type": {
			gc: &types.GetConfig{
				Name:     "test_get_type",
				Paths:    []string{"/interfaces"},
				Encoding: &encoding,
				Type:     "STATE",
				Interval: 30 * time.Second,
			},
			tc: &types.TargetConfig{
				Name: "target1",
			},
			want: &gnmi.GetRequest{
				Path: []*gnmi.Path{
					{
						Elem: []*gnmi.PathElem{
							{Name: "interfaces"},
						},
					},
				},
				Encoding: gnmi.Encoding_JSON,
				Type:     gnmi.GetRequest_STATE,
			},
			wantErr: false,
		},
		"get_request_multiple_paths": {
			gc: &types.GetConfig{
				Name:     "test_get_multi",
				Paths:    []string{"/interfaces", "/system/state"},
				Encoding: &encoding,
				Interval: 30 * time.Second,
			},
			tc: &types.TargetConfig{
				Name: "target1",
			},
			want: &gnmi.GetRequest{
				Path: []*gnmi.Path{
					{
						Elem: []*gnmi.PathElem{
							{Name: "interfaces"},
						},
					},
					{
						Elem: []*gnmi.PathElem{
							{Name: "system"},
							{Name: "state"},
						},
					},
				},
				Encoding: gnmi.Encoding_JSON,
			},
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := New()
			got, err := c.CreateGetRequestFromConfig(tt.gc, tt.tc)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateGetRequestFromConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !proto.Equal(got, tt.want) {
				t.Errorf("CreateGetRequestFromConfig() mismatch:\ngot:  %v\nwant: %v", got, tt.want)
			}
		})
	}
}

func TestGetConfigClone(t *testing.T) {
	encoding := "json"
	original := &types.GetConfig{
		Name:            "test",
		Paths:           []string{"/path1", "/path2"},
		Prefix:          "/prefix",
		Target:          "target",
		SetTarget:       true,
		Type:            "STATE",
		Encoding:        &encoding,
		Models:          []string{"model1"},
		Interval:        30 * time.Second,
		Outputs:         []string{"output1"},
		Depth:           5,
		EventProcessors: []string{"proc1"},
		Targets:         []string{"target1"},
	}

	clone := original.Clone()

	// Verify all fields are equal
	if clone.Name != original.Name {
		t.Errorf("Clone() Name = %v, want %v", clone.Name, original.Name)
	}
	if clone.Prefix != original.Prefix {
		t.Errorf("Clone() Prefix = %v, want %v", clone.Prefix, original.Prefix)
	}
	if clone.Target != original.Target {
		t.Errorf("Clone() Target = %v, want %v", clone.Target, original.Target)
	}
	if clone.SetTarget != original.SetTarget {
		t.Errorf("Clone() SetTarget = %v, want %v", clone.SetTarget, original.SetTarget)
	}
	if clone.Type != original.Type {
		t.Errorf("Clone() Type = %v, want %v", clone.Type, original.Type)
	}
	if clone.Interval != original.Interval {
		t.Errorf("Clone() Interval = %v, want %v", clone.Interval, original.Interval)
	}
	if clone.Depth != original.Depth {
		t.Errorf("Clone() Depth = %v, want %v", clone.Depth, original.Depth)
	}

	// Verify slices are deep copied
	clone.Paths[0] = "modified"
	if original.Paths[0] == "modified" {
		t.Error("Clone() Paths is not a deep copy")
	}

	// Verify encoding is deep copied
	newEncoding := "proto"
	clone.Encoding = &newEncoding
	if *original.Encoding != "json" {
		t.Error("Clone() Encoding is not a deep copy")
	}
}

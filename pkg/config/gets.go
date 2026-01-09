// © 2022 Nokia.
//
// This code is a Contribution to the gNMIc project ("Work") made under the Google Software Grant and Corporate Contributor License Agreement ("CLA") and governed by the Apache License 2.0.
// No other rights or licenses in or to any of Nokia's intellectual property are granted for any other purpose.
// This code is provided on an "as is" basis without any warranties of any kind.
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/mitchellh/mapstructure"
	"github.com/openconfig/gnmi/proto/gnmi"

	"github.com/openconfig/gnmic/pkg/api"
	"github.com/openconfig/gnmic/pkg/api/types"
)

const (
	getDefaultType     = "ALL"
	getDefaultEncoding = "JSON"
	getDefaultInterval = 30 * time.Second
)

// GetGets loads get configurations from the config file.
func (c *Config) GetGets() (map[string]*types.GetConfig, error) {
	getDef := c.FileConfig.GetStringMap("gets")
	if c.Debug {
		c.logger.Printf("gets map: %#v", getDef)
	}
	for gn, g := range getDef {
		switch g := g.(type) {
		case map[string]interface{}:
			gc, err := c.decodeGetConfig(gn, g)
			if err != nil {
				return nil, err
			}
			c.Gets[gn] = gc
		default:
			return nil, fmt.Errorf("%w: gets map: unexpected type %T", ErrConfig, g)
		}
	}
	if c.Debug {
		c.logger.Printf("gets: %+v", c.Gets)
	}
	err := validateGetsConfig(c.Gets)
	if err != nil {
		return nil, err
	}
	return c.Gets, nil
}

func (c *Config) decodeGetConfig(gn string, g interface{}) (*types.GetConfig, error) {
	gc := new(types.GetConfig)
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
			Result:     gc,
		})
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(g)
	if err != nil {
		return nil, err
	}
	gc.Name = gn
	c.setGetDefaults(gc)
	expandGetEnv(gc)
	return gc, nil
}

func (c *Config) setGetDefaults(gc *types.GetConfig) {
	if gc.Type == "" {
		gc.Type = getDefaultType
	}
	if gc.Interval == 0 {
		gc.Interval = getDefaultInterval
	}
	if gc.Encoding == nil {
		if c.Encoding != "" {
			gc.Encoding = pointer.ToString(c.Encoding)
		} else {
			gc.Encoding = pointer.ToString(getDefaultEncoding)
		}
	}
}

func expandGetEnv(gc *types.GetConfig) {
	gc.Name = os.ExpandEnv(gc.Name)
	for i := range gc.Paths {
		gc.Paths[i] = os.ExpandEnv(gc.Paths[i])
	}
	gc.Prefix = os.ExpandEnv(gc.Prefix)
	gc.Target = os.ExpandEnv(gc.Target)
	gc.Type = os.ExpandEnv(gc.Type)
	if gc.Encoding != nil {
		gc.Encoding = pointer.ToString(os.ExpandEnv(*gc.Encoding))
	}
	for i := range gc.Models {
		gc.Models[i] = os.ExpandEnv(gc.Models[i])
	}
	for i := range gc.Outputs {
		gc.Outputs[i] = os.ExpandEnv(gc.Outputs[i])
	}
	for i := range gc.Targets {
		gc.Targets[i] = os.ExpandEnv(gc.Targets[i])
	}
}

func validateGetsConfig(gets map[string]*types.GetConfig) error {
	for name, gc := range gets {
		if err := validateGetConfig(name, gc); err != nil {
			return err
		}
	}
	return nil
}

func validateGetConfig(name string, gc *types.GetConfig) error {
	if len(gc.Paths) == 0 && gc.Prefix == "" {
		return fmt.Errorf("%w: get %q: missing paths", ErrConfig, name)
	}
	if gc.Interval <= 0 {
		return fmt.Errorf("%w: get %q: interval must be greater than 0", ErrConfig, name)
	}
	// Validate type
	switch strings.ToUpper(gc.Type) {
	case "ALL", "CONFIG", "STATE", "OPERATIONAL", "":
	default:
		return fmt.Errorf("%w: get %q: unknown type %q, must be one of: ALL, CONFIG, STATE, OPERATIONAL", ErrConfig, name, gc.Type)
	}
	// Validate encoding if specified
	if gc.Encoding != nil {
		switch strings.ToUpper(strings.ReplaceAll(*gc.Encoding, "-", "_")) {
		case "JSON", "BYTES", "PROTO", "ASCII", "JSON_IETF", "":
		default:
			return fmt.Errorf("%w: get %q: unknown encoding %q", ErrConfig, name, *gc.Encoding)
		}
	}
	return nil
}

// GetGetsFromFile returns a sorted slice of GetConfigs.
func (c *Config) GetGetsFromFile() []*types.GetConfig {
	gets, err := c.GetGets()
	if err != nil {
		return nil
	}
	result := make([]*types.GetConfig, 0, len(gets))
	for _, gc := range gets {
		result = append(result, gc)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// CreateGetRequestFromConfig builds a gnmi.GetRequest from a GetConfig.
func (c *Config) CreateGetRequestFromConfig(gc *types.GetConfig, tc *types.TargetConfig) (*gnmi.GetRequest, error) {
	gnmiOpts := make([]api.GNMIOption, 0, 8)

	// Add prefix
	if gc.Prefix != "" {
		gnmiOpts = append(gnmiOpts, api.Prefix(gc.Prefix))
	}

	// Add paths
	for _, p := range gc.Paths {
		gnmiOpts = append(gnmiOpts, api.Path(p))
	}

	// Add encoding
	switch {
	case gc.Encoding != nil:
		gnmiOpts = append(gnmiOpts, api.Encoding(*gc.Encoding))
	case tc != nil && tc.Encoding != nil:
		gnmiOpts = append(gnmiOpts, api.Encoding(*tc.Encoding))
	default:
		gnmiOpts = append(gnmiOpts, api.Encoding(c.Encoding))
	}

	// Add data type
	if gc.Type != "" {
		gnmiOpts = append(gnmiOpts, api.DataType(gc.Type))
	}

	// Add target
	if gc.Target != "" {
		gnmiOpts = append(gnmiOpts, api.Target(gc.Target))
	} else if gc.SetTarget && tc != nil {
		gnmiOpts = append(gnmiOpts, api.Target(tc.Name))
	}

	// Add models
	for _, m := range gc.Models {
		gnmiOpts = append(gnmiOpts, api.UseModel(m, "", ""))
	}

	// Add depth extension
	if gc.Depth > 0 {
		gnmiOpts = append(gnmiOpts, api.Extension_Depth(gc.Depth))
	}

	return api.NewGetRequest(gnmiOpts...)
}

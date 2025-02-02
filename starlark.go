// Copyright 2025 Michael Vittrup Larsen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/starlark"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	starlarkRunGroup      = "fn.kpt.dev"
	starlarkRunVersion    = "v1alpha1"
	starlarkRunAPIVersion = starlarkRunGroup + "/" + starlarkRunVersion
	starlarkRunKind       = "StarlarkRun"
)

type FunctionConfig struct {
	yaml.ResourceMeta `yaml:",inline" json:",inline"`
	Params            map[string]any `yaml:"params,omitempty" json:"params,omitempty"`
	Source            string         `yaml:"source,omitempty" json:"source,omitempty"`
}

func (fncfg *FunctionConfig) Default() error { //nolint:unparam // this return is part of the Defaulter interface
	return nil
}

func (fncfg *FunctionConfig) Validate() error {
	if fncfg.TypeMeta.APIVersion != starlarkRunAPIVersion || fncfg.TypeMeta.Kind != starlarkRunKind {
		return fmt.Errorf("unknown function-config: %v/%v", fncfg.TypeMeta.APIVersion, fncfg.TypeMeta.Kind)
	}
	if fncfg.Source == "" {
		return fmt.Errorf("starlark source cannot be empty")
	}
	return nil
}

func Processor() framework.ResourceListProcessor {
	return framework.ResourceListProcessorFunc(func(rl *framework.ResourceList) error {
		fncfg := &FunctionConfig{}
		if err := framework.LoadFunctionConfig(rl.FunctionConfig, fncfg); err != nil {
			return fmt.Errorf("reading function-config: %w", err)
		}
		fltr := starlark.Filter{
			Name:           fncfg.NameMeta.Name,
			Program:        fncfg.Source,
			FunctionFilter: runtimeutil.FunctionFilter{FunctionConfig: rl.FunctionConfig},
		}

		out, err := fltr.Filter(rl.Items)
		if err == nil {
			rl.Items = out
		}

		return err
	})
}

func main() {
	cmd := command.Build(Processor(), command.StandaloneEnabled, false)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

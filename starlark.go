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
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	starlarkRunGroup      = "fn.kpt.dev"
	starlarkRunVersion    = "v1alpha1"
	starlarkRunAPIVersion = starlarkRunGroup + "/" + starlarkRunVersion
	starlarkRunKind       = "StarlarkRun"
)

type FunctionConfig struct {
	yaml.ResourceMeta `yaml:".inline" json:".inline"`
	Params            map[string]any `yaml:"params,omitempty" json:"params,omitempty"`
	Source            string         `yaml:"source,omitempty" json:"source,omitempty"`
}

// type FilterState struct {
// 	fnConfig  *FunctionConfig
// 	Results   framework.Results
// }

func (fncfg *FunctionConfig) Default() error { //nolint:unparam // this return is part of the Defaulter interface
	return nil
}

func (fncfg *FunctionConfig) Validate() error {
	// TODO: Check api version and kind
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
		// filter := FilterState{
		// 	fnConfig:  config,
		// }
		fltr := starlark.Filter{
			Name:           fncfg.NameMeta.Name,
			Program:        fncfg.Source,
			FunctionFilter: runtimeutil.FunctionFilter{FunctionConfig: rl.FunctionConfig},
		}

		var out []*yaml.RNode
		err := kio.Pipeline{
			Inputs:  []kio.Reader{&kio.PackageBuffer{}},
			Filters: []kio.Filter{&fltr},
			Outputs: []kio.Writer{&kio.PackageBuffer{out}},
		}.Execute()
		fmt.Fprintf(os.Stderr, ">> %v\n", len(out))

		//rl.Results = append(rl.Results, filter.Results...)

		return err
	})
}

func main() {
	cmd := command.Build(Processor(), command.StandaloneEnabled, false)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

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

	"go.starlark.net/resolve"
	corev1 "k8s.io/api/core/v1"
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

func (fnCfg *FunctionConfig) LoadFunctionConfig(o *yaml.RNode) error {
	if o.GetKind() == "ConfigMap" && o.GetApiVersion() == "v1" {
		var cm corev1.ConfigMap
		if err := yaml.Unmarshal([]byte(o.MustString()), &cm); err != nil {
			return err
		}
		_, ok := cm.Data["source"]
		if !ok {
			return fmt.Errorf("no 'source' in function-config")
		}
		fnCfg.Source = cm.Data["source"]
		fnCfg.Params = map[string]any{}
		for k, v := range cm.Data {
			if k != "source" {
				fnCfg.Params[k] = v
			}
		}
		return nil
	} else if o.GetKind() == starlarkRunKind && o.GetApiVersion() == starlarkRunAPIVersion {
		if err := yaml.Unmarshal([]byte(o.MustString()), &fnCfg); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("unknown function config")
}

func Processor() framework.ResourceListProcessor {
	return framework.ResourceListProcessorFunc(func(rl *framework.ResourceList) error {
		fncfg := &FunctionConfig{}
		if err := fncfg.LoadFunctionConfig(rl.FunctionConfig); err != nil {
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
	resolve.AllowRecursion = true
	resolve.AllowGlobalReassign = true
	resolve.AllowSet = true
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

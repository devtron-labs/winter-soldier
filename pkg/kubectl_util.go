/*
Copyright 2021 Devtron Labs Pvt Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/restmapper"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
)

type Mapper struct {
	mapper             meta.RESTMapper
	categoryExpanderFn func() (restmapper.CategoryExpander, error)
}

type ArgsProcessor struct {
	mapper         *Mapper
	errs           []error
	resourceTuples []resourceTuple
	names          []string
	resources      []string
}

type resourceTuple struct {
	Resource string
	Name     string
}

func NewMapperFactory() *Mapper {
	restConfig := ctrl.GetConfigOrDie()
	disco, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		panic(err)
	}
	categoryExpanderFn := func() (restmapper.CategoryExpander, error) {
		return restmapper.NewDiscoveryCategoryExpander(disco), err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(disco))
	shortcutExpander := restmapper.NewShortcutExpander(mapper, disco)
	return &Mapper{
		mapper:             shortcutExpander,
		categoryExpanderFn: categoryExpanderFn,
	}
}

func NewFactory(mapper *Mapper) *ArgsProcessor {
	return &ArgsProcessor{
		mapper: mapper,
	}
}

func (a *ArgsProcessor) ResourceTuples() []resourceTuple {
	if len(a.resourceTuples) > 0 {
		return a.resourceTuples
	}
	if len(a.resources) > 0 && len(a.names) == 0 {
		return []resourceTuple{
			{
				Resource: a.resources[0],
			},
		}
	}
	if len(a.resources) > 0 && len(a.names) > 0 {
		var resourceTuples []resourceTuple
		for _, name := range a.names {
			resourceTuples = append(resourceTuples, resourceTuple{
				Resource: a.resources[0],
				Name:     name,
			})
		}
		return resourceTuples
	}
	return []resourceTuple{}
}

// ResourceTypeOrNameArgs indicates that the builder should accept arguments
// of the form `(<type1>[,<type2>,...]|<type> <name1>[,<name2>,...])`. When one argument is
// received, the types provided will be retrieved from the server (and be comma delimited).
// When two or more arguments are received, they must be a single type and resource name(s).
// The allowEmptySelector permits to select all the resources (via Everything func).
func (a *ArgsProcessor) ResourceTypeOrNameArgs(args ...string) {
	args = normalizeMultipleResourcesArgs(args)
	if ok, err := hasCombinedTypeArgs(args); ok {
		if err != nil {
			a.errs = append(a.errs, err)
			return
		}
		for _, s := range args {
			tuple, ok, err := splitResourceTypeName(s)
			if err != nil {
				a.errs = append(a.errs, err)
				return
			}
			if ok {
				a.resourceTuples = append(a.resourceTuples, tuple)
			}
		}
		return
	}
	if len(args) > 0 {
		// Try replacing aliases only in types
		args[0] = a.ReplaceAliases(args[0])
	}
	switch {
	case len(args) > 2:
		a.names = append(a.names, args[1:]...)
		a.ResourceTypes(SplitResourceArgument(args[0])...)
	case len(args) == 2:
		a.names = append(a.names, args[1])
		a.ResourceTypes(SplitResourceArgument(args[0])...)
	case len(args) == 1:
		a.ResourceTypes(SplitResourceArgument(args[0])...)
	case len(args) == 0:
	default:
		a.errs = append(a.errs, fmt.Errorf("arguments must consist of a resource or a resource and name"))
	}
}

// Normalize args convert multiple resources to resource tuples, a,b,c d
// as a transform to a/d b/d c/d
func normalizeMultipleResourcesArgs(args []string) []string {
	if len(args) >= 2 {
		resources := []string{}
		resources = append(resources, SplitResourceArgument(args[0])...)
		if len(resources) > 1 {
			names := []string{}
			names = append(names, args[1:]...)
			newArgs := []string{}
			for _, resource := range resources {
				for _, name := range names {
					newArgs = append(newArgs, strings.Join([]string{resource, name}, "/"))
				}
			}
			return newArgs
		}
	}
	return args
}

// splitResourceTypeName handles type/name resource formats and returns a resource tuple
// (empty or not), whether it successfully found one, and an error
func splitResourceTypeName(s string) (resourceTuple, bool, error) {
	if !strings.Contains(s, "/") {
		return resourceTuple{}, false, nil
	}
	seg := strings.Split(s, "/")
	if len(seg) != 2 {
		return resourceTuple{}, false, fmt.Errorf("arguments in resource/name form may not have more than one slash")
	}
	resource, name := seg[0], seg[1]
	if len(resource) == 0 || len(name) == 0 || len(SplitResourceArgument(resource)) != 1 {
		return resourceTuple{}, false, fmt.Errorf("arguments in resource/name form must have a single resource and name")
	}
	return resourceTuple{Resource: resource, Name: name}, true, nil
}

func hasCombinedTypeArgs(args []string) (bool, error) {
	hasSlash := 0
	for _, s := range args {
		if strings.Contains(s, "/") {
			hasSlash++
		}
	}
	switch {
	case hasSlash > 0 && hasSlash == len(args):
		return true, nil
	case hasSlash > 0 && hasSlash != len(args):
		baseCmd := "cmd"
		if len(os.Args) > 0 {
			baseCmdSlice := strings.Split(os.Args[0], "/")
			baseCmd = baseCmdSlice[len(baseCmdSlice)-1]
		}
		return true, fmt.Errorf("there is no need to specify a resource type as a separate argument when passing arguments in resource/name form (e.g. '%s get resource/<resource_name>' instead of '%s get resource resource/<resource_name>'", baseCmd, baseCmd)
	default:
		return false, nil
	}
}

// SplitResourceArgument splits the argument with commas and returns unique
// strings in the original order.
func SplitResourceArgument(arg string) []string {
	out := []string{}
	set := sets.NewString()
	for _, s := range strings.Split(arg, ",") {
		if set.Has(s) {
			continue
		}
		set.Insert(s)
		out = append(out, s)
	}
	return out
}

// ReplaceAliases accepts an argument and tries to expand any existing
// aliases found in it
func (a *ArgsProcessor) ReplaceAliases(input string) string {
	replaced := []string{}
	for _, arg := range strings.Split(input, ",") {
		if a.mapper.categoryExpanderFn == nil {
			continue
		}
		categoryExpander, err := a.mapper.categoryExpanderFn()
		if err != nil {
			a.AddError(err)
			continue
		}

		if resources, ok := categoryExpander.Expand(arg); ok {
			asStrings := []string{}
			for _, resource := range resources {
				if len(resource.Group) == 0 {
					asStrings = append(asStrings, resource.Resource)
					continue
				}
				asStrings = append(asStrings, resource.Resource+"."+resource.Group)
			}
			arg = strings.Join(asStrings, ",")
		}
		replaced = append(replaced, arg)
	}
	return strings.Join(replaced, ",")
}

func (a *ArgsProcessor) AddError(err error) {
	if err == nil {
		return
	}
	a.errs = append(a.errs, err)
}

// ResourceTypes is a list of types of resources to operate on, when listing objects on
// the server or retrieving objects that match a selector.
func (a *ArgsProcessor) ResourceTypes(types ...string) {
	a.resources = append(a.resources, types...)
}

// mappingFor returns the RESTMapping for the Kind given, or the Kind referenced by the resource.
// Prefers a fully specified GroupVersionResource match. If one is not found, we match on a fully
// specified GroupVersionKind, or fallback to a match on GroupKind.
func (a *ArgsProcessor) MappingFor(resourceOrKindArg string) (*meta.RESTMapping, error) {
	fullySpecifiedGVR, groupResource := schema.ParseResourceArg(resourceOrKindArg)
	gvk := schema.GroupVersionKind{}
	restMapper := a.mapper.mapper

	if fullySpecifiedGVR != nil {
		gvk, _ = restMapper.KindFor(*fullySpecifiedGVR)
	}
	if gvk.Empty() {
		gvk, _ = restMapper.KindFor(groupResource.WithVersion(""))
	}
	if !gvk.Empty() {
		return restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	}

	fullySpecifiedGVK, groupKind := schema.ParseKindArg(resourceOrKindArg)
	if fullySpecifiedGVK == nil {
		gvk := groupKind.WithVersion("")
		fullySpecifiedGVK = &gvk
	}

	if !fullySpecifiedGVK.Empty() {
		if mapping, err := restMapper.RESTMapping(fullySpecifiedGVK.GroupKind(), fullySpecifiedGVK.Version); err == nil {
			return mapping, nil
		}
	}

	mapping, err := restMapper.RESTMapping(groupKind, gvk.Version)
	if err != nil {
		// if we error out here, it is because we could not match a resource or a kind
		// for the given argument. To maintain consistency with previous behavior,
		// announce that a resource type could not be found.
		// if the error is _not_ a *meta.NoKindMatchError, then we had trouble doing discovery,
		// so we should return the original error since it may help a user diagnose what is actually wrong
		if meta.IsNoMatchError(err) {
			return nil, fmt.Errorf("the server doesn't have a resource type %q", groupResource.Resource)
		}
		return nil, err
	}

	return mapping, nil
}

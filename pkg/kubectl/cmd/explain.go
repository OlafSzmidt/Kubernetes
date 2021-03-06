/*
Copyright 2014 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"io"

	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/api/latest"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const (
	explainExamples = `# Get the documentation of the resource and its fields
$ kubectl explain pods

# Get the documentation of a specific field of a resource
$ kubectl explain pods.spec.containers`

	explainLong = `Documentation of resources.

Possible resource types include: pods (po), services (svc),
replicationcontrollers (rc), nodes (no), events (ev), componentstatuses (cs),
limitranges (limits), persistentvolumes (pv), persistentvolumeclaims (pvc),
resourcequotas (quota), namespaces (ns), horizontalpodautoscalers (hpa)
or endpoints (ep).`
)

// NewCmdExplain returns a cobra command for swagger docs
func NewCmdExplain(f *cmdutil.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "explain RESOURCE",
		Short:   "Documentation of resources.",
		Long:    explainLong,
		Example: explainExamples,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunExplain(f, out, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	cmd.Flags().Bool("recursive", false, "Print the fields of fields (Currently only 1 level deep)")
	return cmd
}

// RunExplain executes the appropriate steps to print a model's documentation
func RunExplain(f *cmdutil.Factory, out io.Writer, cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmdutil.UsageError(cmd, "We accept only this format: explain RESOURCE")
	}

	client, err := f.Client()
	if err != nil {
		return err
	}

	recursive := cmdutil.GetFlagBool(cmd, "recursive")
	apiVersionString := cmdutil.GetFlagString(cmd, "api-version")
	apiVersion := unversioned.GroupVersion{}

	mapper, _ := f.Object()
	// TODO: After we figured out the new syntax to separate group and resource, allow
	// the users to use it in explain (kubectl explain <group><syntax><resource>).
	// Refer to issue #16039 for why we do this. Refer to PR #15808 that used "/" syntax.
	inModel, fieldsPath, err := kubectl.SplitAndParseResourceRequest(args[0], mapper)
	if err != nil {
		return err
	}

	// TODO: We should deduce the group for a resource by discovering the supported resources at server.
	gvk, err := mapper.KindFor(inModel)
	if err != nil {
		return err
	}

	if len(apiVersionString) == 0 {
		groupMeta, err := latest.Group(gvk.Group)
		if err != nil {
			return err
		}
		apiVersion = groupMeta.GroupVersion

	} else {
		apiVersion, err = unversioned.ParseGroupVersion(apiVersionString)
		if err != nil {
			return nil
		}
	}

	swagSchema, err := kubectl.GetSwaggerSchema(apiVersion, client)
	if err != nil {
		return err
	}

	return kubectl.PrintModelDescription(inModel, fieldsPath, out, swagSchema, recursive)
}

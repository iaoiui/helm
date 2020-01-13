/*
Copyright The Helm Authors.

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

package main

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/releaseutil"
)

const defaultDirectoryPermission = 0755

var getManifestHelp = `
This command fetches the generated manifest for a given release.

A manifest is a YAML-encoded representation of the Kubernetes resources that
were generated from this release's chart(s). If a chart is dependent on other
charts, those resources will also be included in the manifest.
`

func newGetManifestCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {

	client := action.NewGet(cfg)
	var showFiles []string
	cmd := &cobra.Command{
		Use:   "manifest RELEASE_NAME",
		Short: "download the manifest for a named release",
		Long:  getManifestHelp,
		Args:  require.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := client.Run(args[0])
			if err != nil {
				return err
			}
			var manifests bytes.Buffer
			fmt.Fprintln(&manifests, strings.TrimSpace(res.Manifest))
			// fmt.Printf("%v\n", res.Manifest)
			splitManifests := releaseutil.SplitManifests(manifests.String())
			manifestNameRegex := regexp.MustCompile("# Source: [^/]+/(.+)")
			// var manifestsToRender []string
			fileWritten := make(map[string]bool)
			// missing := true
			for _, manifest := range splitManifests {

				submatch := manifestNameRegex.FindStringSubmatch(manifest)
				if len(submatch) == 0 {
					continue
				}
				manifestName := submatch[1]

				// manifest.Name is rendered using linux-style filepath separators on Windows as
				// well as macOS/linux.
				manifestPathSplit := strings.Split(manifestName, "/")
				manifestPath := filepath.Join(manifestPathSplit...)
				// fmt.Printf("%v\n", manifestPath)
				var outputDir = res.Name

				err = action.WriteToFile(outputDir, manifestPath, manifest, fileWritten[manifestPath])
				if err != nil {
					return err
				}
			}
			// fmt.Printf("%v", res.Name)

			return nil
		},
	}

	f := cmd.Flags()
	f.IntVar(&client.Version, "revision", 0, "get the named release with revision")
	f.StringArrayVarP(&showFiles, "show-only", "s", []string{}, "only show manifests rendered from the given templates")
	return cmd
}

// func writeToFile(outputDir string, name string, data string, append bool) error {
// 	outfileName := strings.Join([]string{outputDir, name}, string(filepath.Separator))

// 	err := ensureDirectoryForFile(outfileName)
// 	if err != nil {
// 		return err
// 	}

// 	f, err := createOrOpenFile(outfileName, append)
// 	if err != nil {
// 		return err
// 	}

// 	defer f.Close()

// 	_, err = f.WriteString(fmt.Sprintf("---\n# Source: %s\n%s\n", name, data))
// 	// _, err = f.WriteString(fmt.Sprintf("%s\n", data))

// 	if err != nil {
// 		return err
// 	}

// 	fmt.Printf("wrote %s\n", outfileName)
// 	return nil
// }

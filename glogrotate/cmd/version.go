// Copyright © 2019 Moises P. Sena <moisespsena@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var Version struct {
	Version, Commit, Date, GoVersion string
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show binary version",
	Run: func(cmd *cobra.Command, args []string) {
		b, _ := yaml.Marshal(Version)
		os.Stdout.Write(b)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

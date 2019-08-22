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
	"io"
	"os"

	"github.com/papertrail/go-tail/follower"

	"github.com/spf13/cobra"
)

var tailCmd = &cobra.Command{
	Use:   "follower OUT",
	Args:  cobra.ExactArgs(1),
	Short: "tail with follower OUT file",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var t *follower.Follower
		t, err = follower.New(args[0], follower.Config{
			Whence: io.SeekEnd,
			Offset: 0,
			Reopen: true,
		})
		if err != nil {
			return
		}

		for line := range t.Lines() {
			os.Stdout.Write(line.Bytes())
			println()
		}

		return t.Err()
	},
}

func init() {
	rootCmd.AddCommand(tailCmd)
}

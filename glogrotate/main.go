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

package main

import (
	"encoding/base64"
	"strings"

	"github.com/moisespsena-go/glogrotate/cli/cmd"
)

var goversion, version, commit, date string

func init() {
	if strings.HasPrefix(goversion, "b64:") {
		b, _ := base64.StdEncoding.DecodeString(goversion[4:])
		goversion = string(b)
	}

	cmd.Version.Version,
		cmd.Version.Commit,
		cmd.Version.Date,
		cmd.Version.GoVersion =
		version,
		commit,
		date,
		strings.TrimSpace(goversion)
}

func main() {
	cmd.Execute()
}

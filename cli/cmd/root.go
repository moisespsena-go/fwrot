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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	logrotate "github.com/moisespsena-go/glogrotate"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var name = filepath.Base(os.Args[0])

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   name,
	Short: "Starts file writer rotation reading STDIN and writes to OUT",

	Long: strings.ReplaceAll(`Starts file writer rotation reading STDIN and writes to OUT.
ENV VARIABLES:
	{N}_ROOT, {N}_PATH, {N}_OUT, {N}_DURATION, {N}_MAX_SIZE, {N}_MAX_COUNT, {N}_DIR_MODE, {N}_FILE_MODE
	
TIME FORMAT:
	%Y - Year. (example: 2006)
	%M - Month with left zero pad. (examples: 01, 12)
	%D - Day with left zero pad. (examples: 01, 31)
	%h - Hour with left zero pad. (examples: 00, 05, 23)
	%m - Minute with left zero pad. (examples: 00, 05, 59)
	%s - Second with left zero pad. (examples: 00, 05, 59)
	%Z - Time Zone. If not set, uses UTC time. (examples: +0700, -0330)
`, "{N}", strings.Trim(strings.ToUpper(name), "_")),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var (
			flags          = cmd.Flags()
			out            = viper.GetString("out")
			silent, _      = flags.GetBool("silent")
			printConfig, _ = flags.GetBool("print")
		)

		for _, v := range args {
			if v != "" {
				out = v
			}
		}

		if out == "" {
			return errors.New("OUT not defined")
		}

		var cfg = logrotate.Config{
			HistoryDir:   strings.ReplaceAll(viper.GetString("history-dir"), "OUT", out),
			HistoryPath:  viper.GetString("history-path"),
			MaxSize:      viper.GetString("max-size"),
			HistoryCount: viper.GetInt("history-count"),
			Duration:     viper.GetString("duration"),
			FileMode:     os.FileMode(viper.GetInt("file-mode")),
			DirMode:      os.FileMode(viper.GetInt("dir-mode")),
		}

		var opt logrotate.Options
		if opt, err = cfg.Options(); err != nil {
			return
		}

		if printConfig {
			fmt.Fprintln(os.Stdout, "out: "+out)
			fmt.Fprintln(os.Stdout, cfg.Yaml())
			return
		}

		Rotator := logrotate.New(out, opt)
		defer Rotator.Close()
		var r io.Reader = os.Stdin
		if !silent {
			r = io.TeeReader(r, os.Stdout)
		}

		_, err = io.Copy(Rotator, r)
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")

	flags := rootCmd.Flags()
	flags.StringP("out", "o", "", "the output file.")
	flags.StringP("history-dir", "r", "OUT.history", "history root directory")
	flags.StringP("history-path", "p", "%Y/%M", "dynamic direcotry path inside ROOT DIR using TIME FORMAT")
	flags.StringP("duration", "d", "M", "rotates every DURATION. Accepted values: Y - yearly, M - monthly, W - weekly, D - daily, h - hourly, m - minutely")
	flags.StringP("max-size", "S", "50M", "Forces rotation if current log size is greather then MAX_SIZE. Values in bytes. Examples: 100, 100K, 50M, 1G, 1T")
	flags.IntP("history-count", "C", 0, "Max history log count")
	flags.IntP("dir-mode", "M", 0750, "directory perms")
	flags.Lookup("dir-mode").DefValue = "0750"
	flags.IntP("file-mode", "m", 0640, "file perms")
	flags.Lookup("file-mode").DefValue = "0640"
	flags.Bool("print", false, "print current config")
	flags.Bool("silent", false, "disable tee to STDOUT")

	for _, v := range []string{
		"out", "history-dir", "history-path", "duration",
		"max-size", "history-count", "dir-mode",
		"file-mode", "silent",
	} {
		viper.BindPFlag(v, flags.Lookup(v))
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(name)

	viper.AddConfigPath(".")
	viper.SetEnvPrefix(strings.Trim(strings.ToUpper(name), "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

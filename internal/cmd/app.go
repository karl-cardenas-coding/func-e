// Copyright 2019 Tetrate
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
	"fmt"
	"net/url"
	"os/user"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
)

// NewApp create a new root command. The globals.GlobalOpts parameter allows tests to scope overrides, which avoids
// having to define a flag for everything needed in tests.
func NewApp(o *globals.GlobalOpts) *cli.App {
	var homeDir, envoyVersionsURL string

	app := cli.NewApp()
	app.Name = "func-e"
	app.HelpName = "func-e"
	app.Usage = `Install and run Envoy`
	// Keep lines at 77 to address leading indent of 3 in help statements
	// NOTE: remove indenting ourselves after the first line after urfave/cli#1275.
	app.UsageText = `To run Envoy, execute ` + "`func-e run -c your_envoy_config.yaml`" + `. This
   downloads and installs the latest version of Envoy for you.

   To list versions of Envoy you can use, execute ` + "`func-e versions -a`" + `. To
   choose one, invoke ` + fmt.Sprintf("`func-e use %s`", version.LastKnownEnvoy) + `. This installs into
   ` + fmt.Sprintf("`$FUNC_E_HOME/versions/%s`", version.LastKnownEnvoy) + `, if not already present.

   You may want to override ` + "`$ENVOY_VERSIONS_URL`" + ` to supply custom builds or
   otherwise control the source of Envoy binaries. When overriding, validate
   your JSON first: https://archive.tetratelabs.io/release-versions-schema.json`
	app.Version = string(version.FuncE)
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "home-dir",
			Usage:       "func-e home directory (location of installed versions and run archives)",
			DefaultText: globals.DefaultHomeDir,
			Destination: &homeDir,
			EnvVars:     []string{"FUNC_E_HOME"},
		},
		&cli.StringFlag{
			Name:        "envoy-versions-url",
			Usage:       "URL of Envoy versions JSON",
			DefaultText: globals.DefaultEnvoyVersionsURL,
			Destination: &envoyVersionsURL,
			EnvVars:     []string{"ENVOY_VERSIONS_URL"},
		}}
	app.Before = func(c *cli.Context) error {
		if err := setHomeDir(o, homeDir); err != nil {
			return err
		}
		return setEnvoyVersionsURL(o, envoyVersionsURL)
	}

	app.HideHelp = true
	app.Commands = []*cli.Command{
		helpCommand,
		NewRunCmd(o),
		NewVersionsCmd(o),
		NewUseCmd(o),
	}
	return app
}

// helpCommand allows us to hide the global flags which cleans up help and markdown
var helpCommand = &cli.Command{
	Name:      "help",
	Usage:     "Shows how to use a [command]",
	ArgsUsage: "[command]",
	Action: func(c *cli.Context) error {
		args := c.Args()
		if args.Present() {
			return cli.ShowCommandHelp(c, args.First())
		}
		return cli.ShowAppHelp(c)
	},
}

func setEnvoyVersionsURL(o *globals.GlobalOpts, versionsURL string) error {
	if o.EnvoyVersionsURL != "" { // overridden for tests
		return nil
	}
	if versionsURL == "" {
		o.EnvoyVersionsURL = globals.DefaultEnvoyVersionsURL
	} else {
		otherURL, err := url.Parse(versionsURL)
		if err != nil || otherURL.Host == "" || otherURL.Scheme == "" {
			return NewValidationError("%q is not a valid Envoy versions URL", versionsURL)
		}
		o.EnvoyVersionsURL = versionsURL
	}
	return nil
}

func setHomeDir(o *globals.GlobalOpts, homeDir string) error {
	if o.HomeDir != "" { // overridden for tests
		return nil
	}
	if homeDir == "" {
		u, err := user.Current()
		if err != nil || u.HomeDir == "" {
			return NewValidationError("unable to determine home directory. Set FUNC_E_HOME instead: %v", err)
		}
		o.HomeDir = filepath.Join(u.HomeDir, ".func-e")
	} else {
		abs, err := filepath.Abs(homeDir)
		if err != nil {
			return NewValidationError(err.Error())
		}
		o.HomeDir = abs
	}
	return nil
}

// Copyright 2021 Tetrate
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

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
	"github.com/tetratelabs/getenvoy/internal/version"
)

// TestGetEnvoyUse needs to always execute, so we run it in a separate home directory
func TestGetEnvoyUse(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	t.Run("not yet installed", func(t *testing.T) {
		stdout, stderr, err := getEnvoyExec("--home-dir", homeDir, "use", version.LastKnownEnvoy)

		require.NoError(t, err)
		require.Regexp(t, `^downloading https:.*tar.*z\n$`, stdout)
		require.Empty(t, stderr)

		// The binary was installed
		require.FileExists(t, filepath.Join(homeDir, "versions", version.LastKnownEnvoy, "bin", "envoy"))

		// The current version was written
		f, err := os.ReadFile(filepath.Join(homeDir, "version"))
		require.NoError(t, err)
		require.Equal(t, version.LastKnownEnvoy, string(f))
	})

	t.Run("already installed", func(t *testing.T) {
		stdout, stderr, err := getEnvoyExec("--home-dir", homeDir, "use", version.LastKnownEnvoy)

		require.NoError(t, err)
		require.Equal(t, version.LastKnownEnvoy+" is already downloaded\n", stdout)
		require.Empty(t, stderr)
	})
}

func TestGetEnvoyUse_UnknownVersion(t *testing.T) {
	stdout, stderr, err := getEnvoyExec("use", "1.1.1")

	require.EqualError(t, err, "exit status 1")
	require.Empty(t, stdout)
	require.Equal(t, fmt.Sprintf(`error: couldn't find version "1.1.1" for platform "%s/%s"
`, runtime.GOOS, runtime.GOARCH), stderr)
}

package env_v0_4

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/stretchr/testify/require"
)

func TestFreshDefaultNamespace(t *testing.T) {
	t.Setenv("ZROK2_DEFAULT_NAMESPACE", "")
	root := &Root{meta: &env_core.Metadata{RootPath: t.TempDir()}, cfg: &env_core.Config{DefaultNamespace: "cached"}, env: &env_core.Environment{AccountToken: "original", ZitiIdentity: "identity"}}
	cfg, env := root.cfg, root.env
	path := filepath.Join(root.meta.RootPath, "config.json")
	check := func(want string) {
		t.Helper()
		got, err := FreshDefaultNamespace(root)
		require.NoError(t, err)
		require.Equal(t, want, got)
		require.Same(t, cfg, root.cfg)
		require.Same(t, env, root.env)
	}
	check("public")
	for _, value := range []string{"team-a", "team-b", ""} {
		require.NoError(t, os.WriteFile(path, []byte(`{"default_namespace":"`+value+`"}`), 0600))
		want := value
		if want == "" {
			want = "public"
		}
		check(want)
	}
	require.NoError(t, os.WriteFile(path, []byte("{"), 0600))
	_, err := FreshDefaultNamespace(root)
	require.Error(t, err)
	t.Setenv("ZROK2_DEFAULT_NAMESPACE", "override")
	check("override")
	t.Setenv("ZROK2_DEFAULT_NAMESPACE", "")
	require.NoError(t, os.Remove(path))
	require.NoError(t, os.Mkdir(path, 0700))
	_, err = FreshDefaultNamespace(root)
	require.Error(t, err)
}

func TestConfigBindingPreservesExistingFormat(t *testing.T) {
	previousRoot := rootDirName
	rootDirName = t.TempDir()
	t.Cleanup(func() { rootDirName = previousRoot })
	cfg := &env_core.Config{ApiEndpoint: "https://controller.example", DefaultNamespace: "team", Headless: true, SuperNetwork: true}
	require.NoError(t, saveConfig(cfg))
	path, err := configFile()
	require.NoError(t, err)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.JSONEq(t, `{"api_endpoint":"https://controller.example","default_namespace":"team","headless":true,"super_network":true}`, string(raw))
	loaded, err := loadConfig()
	require.NoError(t, err)
	require.Equal(t, cfg, loaded)
}

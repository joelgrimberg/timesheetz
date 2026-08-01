package config

import (
	"testing"
)

func benchSetupConfig(b *testing.B) func() {
	b.Helper()
	tmpDir := b.TempDir()
	SetConfigPathOverride(tmpDir + "/config.json")
	SaveConfig(Config{
		Name:           "Bench User",
		CompanyName:    "Bench Co",
		FreeSpeech:     "",
		APIPort:        8080,
		StartAPIServer: true,
	})
	return func() { SetConfigPathOverride("") }
}

func BenchmarkGetAPIPort(b *testing.B) {
	cleanup := benchSetupConfig(b)
	defer cleanup()
	SetRuntimePort(0)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetAPIPort()
	}
}

func BenchmarkGetUserConfig(b *testing.B) {
	cleanup := benchSetupConfig(b)
	defer cleanup()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _, err := GetUserConfig()
		if err != nil {
			b.Fatal(err)
		}
	}
}

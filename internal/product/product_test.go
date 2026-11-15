package product

import (
	"os"
	"testing"
)

func TestReadsTheBuildConfig(t *testing.T) {
	config, err := os.ReadFile("../../build/config.yml")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Read(config)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "CoolSSH" || got.Version == "" || got.Copyright == "" || got.Donate == "" {
		t.Fatalf("read %+v", got)
	}
}

func TestRefusesAConfigWithoutAVersion(t *testing.T) {
	for name, config := range map[string]string{
		"no info":    "version: '3'\n",
		"no version": "info:\n  productName: coolssh\n",
		"not yaml":   "info: [",
	} {
		if _, err := Read([]byte(config)); err == nil {
			t.Errorf("%s: read without an error", name)
		}
	}
}

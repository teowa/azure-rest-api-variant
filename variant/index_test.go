package variant_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/teowa/azure-rest-api-variants/variant"
)

func TestBuild(t *testing.T) {
	specDir := os.Getenv("AZURE_REST_REPO_DIR")

	if !strings.HasSuffix(specDir, "specification") {
		t.Fatalf("AZURE_REST_REPO_DIR must specify the specification folder, e.g., AZURE_REST_REPO_DIR=\"/home/test/go/src/github.com/azure/azure-rest-api-specs/specification\"")
	}

	index, err := variant.Build(specDir)
	if err != nil {
		t.Fatalf("build error: %+v", err)
	}

	jsonString, err := json.MarshalIndent(index, "", "\t")
	if err != nil {
		t.Fatalf("marshal error: %+v", err)
	}

	if err = os.WriteFile("../variants.json", jsonString, 0644); err != nil {
		t.Fatalf("write file error: %+v", err)
	}
}

func TestLoad(t *testing.T) {
	b, err := os.ReadFile("../variants.json")
	if err != nil {
		t.Fatalf("read file error: %+v", err)
	}

	index := variant.Index{}
	if err = json.Unmarshal(b, &index); err != nil {
		t.Fatalf("unmarshal error: %+v", err)
	}

	t.Logf("index hash: %v, count: %v", index.Commit, index.Count)
}

package variant

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-openapi/jsonreference"
	openapiSpec "github.com/go-openapi/spec"
)

type Index struct {
	Commit   string                            `json:"commit,omitempty"`
	Count    int                               `json:"count,omitempty"`
	Variants map[string]map[string]interface{} `json:"variants,omitempty"`
}

var (
	currentSpecDir string
	lock           = &sync.RWMutex{}
)

func Build(specDir string) (*Index, error) {
	currentSpecDir = specDir

	var commit string
	repo, err := git.PlainOpen(filepath.Dir(specDir))
	if err != nil {
		if err != git.ErrRepositoryNotExists {
			return nil, err
		}
	} else {
		ref, err := repo.Head()
		if err != nil {
			return nil, err
		}
		commit = ref.Hash().String()
	}

	log.Printf("[INFO] commit: %s", commit)

	index := &Index{
		Commit:   commit,
		Variants: make(map[string]map[string]interface{}),
		Count:    0,
	}

	log.Printf("[INFO] Collecting specs dir: %s", specDir)
	specPaths, err := CollectSpecs(specDir)
	if err != nil {
		return nil, fmt.Errorf("collecting specs: %v", err)
	}

	log.Printf("[INFO] scanning index from %d spec files", len(specPaths))

	specChan := make(chan string)

	var waitGroup sync.WaitGroup
	waitGroup.Add(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func(i int) {
			defer waitGroup.Done()
			for specPath := range specChan {
				//log.Printf("%d [INFO] scanning spec file: %s", i, specPath)
				doc, err := loadSwagger(specPath)
				if err != nil {
					log.Fatal(err)
				}

				spec := doc.Spec()
				for name, schema := range spec.Definitions {
					index.markVariant(name, specPath, schema, spec, map[string]interface{}{})
				}

			}
		}(i)
	}

	for _, specPath := range specPaths {
		specChan <- specPath
	}
	close(specChan)

	done := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Minute):
		return nil, fmt.Errorf("timeout while scanning specs")
	}

	log.Printf("[INFO] Count: %d variants", index.Count)

	return index, nil
}

func (index *Index) markVariant(schemaName, specPath string, schema openapiSpec.Schema, spec *openapiSpec.Swagger, resolvedVariant map[string]interface{}) bool {
	if _, ok := resolvedVariant[makeKey(specPath, schemaName)]; ok {
		log.Printf("[WARN] %s cycle detected", makeKey(specPath, schemaName))
		return false
	}
	resolvedVariant[makeKey(specPath, schemaName)] = nil
	defer delete(resolvedVariant, makeKey(specPath, schemaName))

	if schema.Discriminator != "" {
		return true
	}

	if index.exists(makeKey(specPath, schemaName)) {
		return true
	}

	for _, allOf := range schema.AllOf {
		if allOf.Ref.String() != "" {
			resolved, err := openapiSpec.ResolveRefWithBase(spec, &allOf.Ref, &openapiSpec.ExpandOptions{RelativeBase: specPath})
			if err != nil {
				log.Fatal(err)
			}

			refSpec := spec
			refSchemaName, refSpecPath := SchemaNamePathFromRef(specPath, allOf.Ref)
			if refSpecPath != specPath {
				doc, err := loadSwagger(refSpecPath)
				if err != nil {
					log.Fatalf("[ERROR] load swagger %s: %+v", refSpecPath, err)
				}
				refSpec = doc.Spec()
			}
			if index.markVariant(refSchemaName, refSpecPath, *resolved, refSpec, resolvedVariant) {
				index.add(makeKey(refSpecPath, refSchemaName), makeKey(specPath, schemaName))
				return true
			}
		}
	}

	return false
}

func (index *Index) exists(key string) bool {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := index.Variants[key]; ok {
		return true
	}
	return false
}

func (index *Index) add(key, depKey string) {
	lock.Lock()
	defer lock.Unlock()

	//log.Printf("[INFO] %s depends on %s", key, depKey)
	if _, ok := index.Variants[key]; !ok {
		index.Variants[key] = map[string]interface{}{}
	}
	if _, ok := index.Variants[key][depKey]; !ok {
		index.Variants[key][depKey] = nil
		index.Count++
	}
}

func makeKey(specPath, schemaName string) string {
	relPath, err := filepath.Rel(currentSpecDir, specPath)
	if err != nil {
		log.Fatal(err)
	}

	ref := jsonreference.MustCreateRef(relPath + "#/definitions/" + schemaName)

	return ref.String()
}

func SchemaNamePathFromRef(swaggerPath string, ref openapiSpec.Ref) (name string, path string) {
	url := ref.GetURL()
	if url == nil {
		return "", ""
	}

	path = url.Path
	if path == "" {
		path = swaggerPath
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(swaggerPath), path)
	}

	fragments := strings.Split(url.Fragment, "/")
	return fragments[len(fragments)-1], path
}

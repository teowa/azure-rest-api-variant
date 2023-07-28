package variant

import (
	"github.com/go-openapi/loads"
	"github.com/hashicorp/golang-lru/v2"
)

var (

	// {swaggerPath: doc Object}
	swaggerCache, _ = lru.New[string, *loads.Document](20)
)

func loadSwagger(swaggerPath string) (*loads.Document, error) {
	if doc, ok := swaggerCache.Get(swaggerPath); ok {
		return doc, nil
	}

	doc, err := loads.JSONSpec(swaggerPath)
	if err != nil {
		return nil, err
	}
	swaggerCache.Add(swaggerPath, doc)
	return doc, nil
}

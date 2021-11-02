package pkg

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

type MockResourceProcessor struct {
}

func NewMockMapperFactory() *Mapper {
	return &Mapper{}
}

func NewMockFactory(mapper *Mapper) ArgsProcessor {
	return &MockResourceProcessor{}
}

func (m *MockResourceProcessor) MappingFor(resourceOrKindArg string) (*meta.RESTMapping, error) {
	lResourceOrKindArgs := strings.ToLower(resourceOrKindArg)
	tResourceOrKindArgs := strings.Title(lResourceOrKindArgs)
	return &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Resource: lResourceOrKindArgs},
		GroupVersionKind: schema.GroupVersionKind{Kind: tResourceOrKindArgs},
	}, nil
}

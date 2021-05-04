package pkg

import (
	"fmt"
	"testing"
)

func TestArgsProcessor_ResourceTypeOrNameArgs(t *testing.T) {
	mapper := NewMapperFactory()
	type fields struct {
		mapper *Mapper
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "multi resource",
			fields: fields{
				mapper: mapper,
			},
			args: args{
				args: []string{"pod"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewFactory(tt.fields.mapper)
			resource, err := a.MappingFor(tt.args.args[0])
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(resource)
			//a.ResourceTypeOrNameArgs(tt.args.args...)
			//fmt.Printf("size %d\n", len(a.resourceTuples))
			//fmt.Printf("err size %d\n", len(a.errs))
			//for _, rt := range a.ResourceTuples() {
			//	fmt.Printf("%s - %s\n", rt.Name, rt.Resource)
			//
			//}
		})
	}
}

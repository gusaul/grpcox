package core

import (
	"reflect"
	"testing"
)

func Test_prepareImport(t *testing.T) {
	type args struct {
		proto []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "sucess change import path to local",
			args: args{
				proto: []byte(`
				  package testing;

				  import "test.com/owner/repo/content.proto";`),
			},
			want: []byte(`
				  package testing;

				  import "content.proto";`),
		},
		{
			name: "sucess keep google import",
			args: args{
				proto: []byte(`
				  package testing;

				  import "google/proto/buf";
				  import "test.com/owner/repo/content.proto";`),
			},
			want: []byte(`
				  package testing;

				  import "google/proto/buf";
				  import "content.proto";`),
		},
		{
			name: "sucess keep local import",
			args: args{
				proto: []byte(`
				  package testing;

				  import "repo.proto";
				  import "test.com/owner/repo/content.proto";`),
			},
			want: []byte(`
				  package testing;

				  import "repo.proto";
				  import "content.proto";`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepareImport(tt.args.proto); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepareImport() = %v, want %v",
					string(got),
					string(tt.want))
			}
		})
	}
}

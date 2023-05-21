package workspace

import (
	"context"
	"reflect"
	"testing"
)

func TestQuery_genCompFilesByPrefix(t *testing.T) {
	type fields struct {
		ws    *Workspace
		files []*BuildFile
	}
	type args struct {
		ctx    context.Context
		prefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*BuildFile
		wantErr bool
	}{
		{
			name: "root directory",
			fields: fields{
				ws: New("/test/workspace"),
				files: []*BuildFile{
					{Path: "/test/workspace/Makefile", Description: "test workspace"},
					{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile"},
					{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile"},
				},
			},
			args: args{
				ctx:    context.Background(),
				prefix: "//",
			},
			want: []*BuildFile{{Path: "/test/workspace/Makefile", Description: "test workspace"}},
		},
		{
			name: "half completed path",
			fields: fields{
				ws: New("/test/workspace"),
				files: []*BuildFile{
					{Path: "/test/workspace/Makefile", Description: "test workspace"},
					{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile"},
					{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile"},
				},
			},
			args: args{
				ctx:    context.Background(),
				prefix: "//pkg/mmak",
			},
			want: []*BuildFile{
				{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile"},
			},
		},
		{
			name: "multiple matches",
			fields: fields{
				ws: New("/test/workspace"),
				files: []*BuildFile{
					{Path: "/test/workspace/Makefile", Description: "test workspace"},
					{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile"},
					{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile"},
				},
			},
			args: args{
				ctx:    context.Background(),
				prefix: "//pkg/",
			},
			want: []*BuildFile{
				{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile"},
				{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile"},
			},
		},
		{
			name: "match targets",
			fields: fields{
				ws: New("/test/workspace"),
				files: []*BuildFile{
					{Path: "/test/workspace/Makefile", Description: "test workspace"},
					{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile"},
					{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile"},
				},
			},
			args: args{
				ctx:    context.Background(),
				prefix: "//pkg/mmake:",
			},
			want: []*BuildFile{
				{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile"},
				{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Query{
				ws:    tt.fields.ws,
				files: tt.fields.files,
			}
			got, err := w.genCompFiles(tt.args.ctx, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("Query.QueryFilesByPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Query.QueryFilesByPrefix() = %v, want %v", got, tt.want)
				return
			}
			// check each file
			for i, bf := range got {
				if !reflect.DeepEqual(bf, tt.want[i]) {
					t.Errorf("Query.QueryFilesByPrefix() = %v, want %v", bf, tt.want[i])
				}
			}
		})
	}
}

func TestQuery_GetPackageFromFile(t *testing.T) {
	type fields struct {
		ws    *Workspace
		files []*BuildFile
	}
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "root directory",
			fields: fields{
				ws: New("/test/workspace"),
				files: []*BuildFile{
					{Path: "/test/workspace/Makefile", Description: "test workspace"},
					{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile"},
					{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile"},
				},
			},
			args: args{
				filePath: "/test/workspace/Makefile",
			},
			want: "//",
		},
		{
			name: "single slash doesn't",
			fields: fields{
				ws: New("/test/workspace"),
				files: []*BuildFile{
					{Path: "/test/workspace/Makefile", Description: "test workspace"},
					{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile"},
					{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile"},
				},
			},
			args: args{
				filePath: "/test/workspace/Makefile",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Query{
				ws:    tt.fields.ws,
				files: tt.fields.files,
			}
			got, err := q.GetPackageFromFile(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Query.GetPackageFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Query.GetPackageFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

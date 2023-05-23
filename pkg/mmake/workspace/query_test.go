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
		want    []string
		wantErr bool
	}{
		{
			name: "incomplete prefixes complete to nearest path",
			fields: fields{
				ws: New("/test/workspace"),
				files: []*BuildFile{
					{Path: "/test/workspace/Makefile", Description: "test workspace", Label: "//"},
					{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile", Label: "//pkg/mmake"},
					{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile", Label: "//pkg/ffake"},
					{Path: "/test/workspace/pkg/mmake/mmake2/Makefile", Description: "test makefile", Label: "//pkg/mmake/mmake2"},
					{Path: "/test/workspace/pkg/mmake/mmake3/Makefile", Description: "test makefile", Label: "//pkg/mmake/mmake3"},
				},
			},
			args: args{
				ctx:    context.Background(),
				prefix: "//pk",
			},
			want: []string{
				"//pkg/ffake",
				"//pkg/mmake",
			},
		},
		{
			name: "root directory",
			fields: fields{
				ws: New("/test/workspace"),
				files: []*BuildFile{
					{Path: "/test/workspace/Makefile", Description: "test workspace", Label: "//"},
					{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile", Label: "//pkg/mmake"},
					{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile", Label: "//pkg/ffake"},
				},
			},
			args: args{
				ctx:    context.Background(),
				prefix: "//",
			},
			want: []string{
				"//",
				"//pkg/",
			},
		},
		{
			name: "half completed path",
			fields: fields{
				ws: New("/test/workspace"),
				files: []*BuildFile{
					{Path: "/test/workspace/Makefile", Description: "test workspace", Label: "//"},
					{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile", Label: "//pkg/mmake"},
					{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile", Label: "//pkg/ffake"},
				},
			},
			args: args{
				ctx:    context.Background(),
				prefix: "//pkg/mmak",
			},
			want: []string{
				"//pkg/mmake",
			},
		},
		{
			name: "multiple matches",
			fields: fields{
				ws: New("/test/workspace"),
				files: []*BuildFile{
					{Path: "/test/workspace/Makefile", Description: "test workspace", Label: "//"},
					{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile", Label: "//pkg/mmake"},
					{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile", Label: "//pkg/ffake"},
				},
			},
			args: args{
				ctx:    context.Background(),
				prefix: "//pkg/",
			},
			want: []string{
				"//pkg/ffake",
				"//pkg/mmake",
			},
		},
		{
			name: "if a full label, return the subsequent slash",
			fields: fields{
				ws: New("/test/workspace"),
				files: []*BuildFile{
					{Path: "/test/workspace/Makefile", Description: "test workspace", Label: "//"},
					{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile", Label: "//pkg/mmake"},
					{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile", Label: "//pkg/ffake"},
					{Path: "/test/workspace/pkg/mmake/mmake2/Makefile", Description: "test makefile", Label: "//pkg/mmake/mmake2"},
				},
			},
			args: args{
				ctx:    context.Background(),
				prefix: "//pkg/mmake",
			},
			want: []string{
				"//pkg/mmake",
				"//pkg/mmake/",
			},
		},
		{
			name: "single slash doesn't crash",
			fields: fields{
				ws: New("/test/workspace"),
				files: []*BuildFile{
					{Path: "/test/workspace/Makefile", Description: "test workspace", Label: "//"},
					{Path: "/test/workspace/pkg/mmake/Makefile", Description: "test makefile", Label: "//pkg/mmake"},
					{Path: "/test/workspace/pkg/ffake/Makefile", Description: "test makefile", Label: "//pkg/ffake"},
					{Path: "/test/workspace/pkg/mmake/mmake2/Makefile", Description: "test makefile", Label: "//pkg/mmake/mmake2"},
				},
			},
			args: args{
				ctx:    context.Background(),
				prefix: "/",
			},
			want:    []string{},
			wantErr: true,
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
				t.Errorf("Query.genCompFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Query.genCompFiles() mismatching lengths = %v, want %v", got, tt.want)
				return
			}
			// check each file
			for i, bf := range got {
				if !reflect.DeepEqual(bf, tt.want[i]) {
					t.Errorf("Query.genCompFiles() = %v, want %v; complete set: %v, want %v", bf, tt.want[i], got, tt.want)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Query{
				ws:    tt.fields.ws,
				files: tt.fields.files,
			}
			got, err := GetPackageFromFile(tt.args.filePath, q.ws.rootPath)
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

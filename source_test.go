package specter

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestLocalFileSourceLoader_Load(t *testing.T) {
	existingFile := "./source_test.spec.hcl"
	absPath, _ := filepath.Abs(existingFile)

	type args struct {
		target string
	}
	tests := []struct {
		name    string
		args    args
		want    Source
		wantErr bool
	}{
		{
			name: "non existing file should return error",
			args: args{
				target: "does-not-exist",
			},
			want:    Source{},
			wantErr: true,
		},
		{
			name: "existing file should return valid source",
			args: args{
				target: existingFile,
			},
			want: Source{
				Location: absPath,
				Data:     []byte{},
				Format:   HCLSourceFormat,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LocalFileSourceLoader{}
			got, err := l.Load(tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() got = %v, want %v", got, tt.want)
			}
		})
	}
}

package utils

import (
	"reflect"
	"testing"
)

func TestUpgradeablelist(t *testing.T) {
	type args struct {
		version     string
		versionlist []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "mix",
			args: args{
				version:     "v1.10.0",
				versionlist: []string{"v1.10.0", "v1.9.1", "v1.9.0"},
			},
			want: make([]string, 0),
		},
		{
			name: "min",
			args: args{
				version:     "v1.9.0",
				versionlist: []string{"v1.10.0", "v1.9.1", "v1.9.0"},
			},
			want: []string{"v1.10.0", "v1.9.1"},
		},
		{
			name: "middle",
			args: args{
				version:     "v1.9.1",
				versionlist: []string{"v1.10.0", "v1.9.1", "v1.9.0"},
			},
			want: []string{"v1.10.0"},
		},
		{
			name: "two type min",
			args: args{
				version:     "v1.9.1",
				versionlist: []string{"v1.10.0", "v1.10.0-fedlcm-v0.3.0", "v1.9.1", "v1.9.1-fedlcm-v0.2.0"},
			},
			want: []string{"v1.10.0"},
		},
		{
			name: "two type max",
			args: args{
				version:     "v1.10.0",
				versionlist: []string{"v1.10.0", "v1.10.0-fedlcm-v0.3.0", "v1.9.1", "v1.9.1-fedlcm-v0.2.0"},
			},
			want: []string{},
		},
		{
			name: "two type max 1",
			args: args{
				version:     "v1.10.0-fedlcm-v0.3.0",
				versionlist: []string{"v1.10.0", "v1.10.0-fedlcm-v0.3.0", "v1.9.1", "v1.9.1-fedlcm-v0.2.0"},
			},
			want: []string{},
		},
		{
			name: "two type min 1",
			args: args{
				version:     "v1.9.1-fedlcm-v0.2.0",
				versionlist: []string{"v1.10.0", "v1.10.0-fedlcm-v0.3.0", "v1.9.1", "v1.9.1-fedlcm-v0.2.0"},
			},
			want: []string{"v1.10.0-fedlcm-v0.3.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Upgradeablelist(tt.args.version, tt.args.versionlist)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Upgradeablelist() = %v, want %v", got, tt.want)
			}
		})
	}
}

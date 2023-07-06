package changelog

import (
	"embed"
	"testing"
)

//go:embed testdata/*.txt
var testData embed.FS

func TestImage(t *testing.T) {
	k3sImageList, err := testData.ReadFile("testdata/image-list.txt")
	if err != nil {
		t.Fatal(err)
	}
	rke2BuildImages, err := testData.ReadFile("testdata/build-images.txt")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		file    []byte
		image   string
		want    string
		wantErr bool
	}{
		{
			name:    "k3s klipper-helm",
			file:    k3sImageList,
			image:   "klipper-helm",
			want:    "v0.8.0",
			wantErr: false,
		},
		{
			name:    "k3s klipper-lb",
			file:    k3sImageList,
			image:   "klipper-lb",
			want:    "v0.4.4",
			wantErr: false,
		},
		{
			name:    "rke2 hardened-coredns",
			file:    rke2BuildImages,
			image:   "hardened-coredns",
			want:    "v1.10.1",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := RepoFiles{ImageList: tt.file}
			got, err := files.Image(tt.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
			} else if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildVersion(t *testing.T) {
	k3sVersionScript, err := testData.ReadFile("testdata/k3s.version.sh.txt")
	if err != nil {
		t.Fatal(err)
	}
	rke2VersionScript, err := testData.ReadFile("testdata/rke2.version.sh.txt")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name     string
		file     []byte
		variable string
		want     string
		wantErr  bool
	}{
		{
			name:     "k3s containerd",
			file:     k3sVersionScript,
			variable: "VERSION_CONTAINERD",
			want:     "v0.0.0",
			wantErr:  false,
		},
		{
			name:     "k3s runc",
			file:     k3sVersionScript,
			variable: "VERSION_RUNC",
			want:     "v0.0.0",
			wantErr:  false,
		},
		{
			name:     "rke2 etcd",
			file:     rke2VersionScript,
			variable: "ETCD_VERSION",
			want:     "v3.5.7-k3s1",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := RepoFiles{VersionScript: tt.file}
			got, err := files.BuildVersion(tt.variable)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
			} else if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDockerfileLayer(t *testing.T) {
	file, err := testData.ReadFile("testdata/Dockerfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		dockerfile []byte
		component  string
		want       string
		wantErr    bool
	}{
		{
			name:       "v1.27.3+rke2r1 hardened-containerd",
			dockerfile: file,
			component:  "hardened-containerd",
			want:       "v1.7.1",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := RepoFiles{Dockerfile: tt.dockerfile}
			got, err := files.DockerfileLayer(tt.component)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
			} else if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDockerfileChart(t *testing.T) {
	file, err := testData.ReadFile("testdata/Dockerfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		dockerfile []byte
		chart      string
		want       string
		wantErr    bool
	}{
		{
			name:       "v1.27.3+rke2r1 rke2-cilium",
			dockerfile: file,
			chart:      "rke2-cilium",
			want:       "1.13.200",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 rke2-canal",
			dockerfile: file,
			chart:      "rke2-canal",
			want:       "v3.25.1",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 rke2-calico",
			dockerfile: file,
			chart:      "rke2-calico",
			want:       "v3.25.002",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 rke2-calico-crd",
			dockerfile: file,
			chart:      "rke2-calico-crd",
			want:       "v3.25.002",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 rke2-coredns",
			dockerfile: file,
			chart:      "rke2-coredns",
			want:       "1.24.004",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 rke2-ingress-nginx",
			dockerfile: file,
			chart:      "rke2-ingress-nginx",
			want:       "4.6.100",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 rke2-metrics-server",
			dockerfile: file,
			chart:      "rke2-metrics-server",
			want:       "2.11.100",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 rke2-multus",
			dockerfile: file,
			chart:      "rke2-multus",
			want:       "v3.9.3",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 rancher-vsphere-cpi",
			dockerfile: file,
			chart:      "rancher-vsphere-cpi",
			want:       "1.5.100",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 rancher-vsphere-csi",
			dockerfile: file,
			chart:      "rancher-vsphere-csi",
			want:       "3.0.1",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 harvester-cloud-provider",
			dockerfile: file,
			chart:      "harvester-cloud-provider",
			want:       "0.2.200",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 harvester-csi-driver",
			dockerfile: file,
			chart:      "harvester-csi-driver",
			want:       "0.1.1600",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 rke2-snapshot-controller",
			dockerfile: file,
			chart:      "rke2-snapshot-controller",
			want:       "1.7.202",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 rke2-snapshot-controller-crd",
			dockerfile: file,
			chart:      "rke2-snapshot-controller-crd",
			want:       "1.7.202",
			wantErr:    false,
		}, {
			name:       "v1.27.3+rke2r1 rke2-snapshot-validation-webhook",
			dockerfile: file,
			chart:      "rke2-snapshot-validation-webhook",
			want:       "1.7.101",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := RepoFiles{Dockerfile: tt.dockerfile}
			got, err := files.DockerfileChart(tt.chart)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
			} else if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoDependency(t *testing.T) {
	k3sModfile, err := testData.ReadFile("testdata/k3s.go.mod.txt")
	if err != nil {
		t.Fatal(err)
	}
	rke2Modfile, err := testData.ReadFile("testdata/rke2.go.mod.txt")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		modfile []byte
		module  string
		want    string
		wantErr bool
	}{
		{
			name:    "k3s kine",
			modfile: k3sModfile,
			module:  "kine",
			want:    "v0.10.1",
			wantErr: false,
		},
		{
			name:    "k3s etcd",
			modfile: k3sModfile,
			module:  "etcd/api/v3",
			want:    "v3.5.7-k3s1",
			wantErr: false,
		},
		{
			name:    "k3s flannel",
			modfile: k3sModfile,
			module:  "flannel",
			want:    "v0.22.0",
			wantErr: false,
		},
		{
			name:    "k3s helm-controller",
			modfile: k3sModfile,
			module:  "helm-controller",
			want:    "v0.15.2",
			wantErr: false,
		},
		{
			name:    "rke2 helm-controller",
			modfile: rke2Modfile,
			module:  "helm-controller",
			want:    "v0.15.0",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := RepoFiles{ModFile: tt.modfile}
			got, err := files.GoDependency(tt.module)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
			} else if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLite(t *testing.T) {
	file, err := testData.ReadFile("testdata/sqlite-binding.h.txt")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		file    []byte
		want    string
		wantErr bool
	}{
		{
			name:    "k3s sqlite",
			file:    file,
			want:    "3.42.0",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := RepoFiles{SQLiteBinding: tt.file}
			got, err := files.SQLite()
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
			} else if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

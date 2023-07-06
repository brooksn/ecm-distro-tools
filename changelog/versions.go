package changelog

import (
	"bufio"
	"errors"
	"regexp"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

var buildScriptExp = regexp.MustCompile(`(?P<version>v[\d\.]+(-k3s.\w*)?)`)
var chartExp = regexp.MustCompile(`CHART_VERSION="([\d\w\.]*)-?[\S]*"`)
var layerExp = regexp.MustCompile(`FROM\s+[\w\d-_\./]+:([\d\w\.]*)`)
var imageListExp = regexp.MustCompile(`:([\w\d\.]+)`)
var sqliteBindingExp = regexp.MustCompile(`define\s*.*SQLITE_VERSION\s*"(.+)"`)

func trimPeriods(v string) string {
	return strings.Replace(v, ".", "", -1)
}

func majMin(v string) (string, error) {
	majMin := semver.MajorMinor(v)
	if majMin == "" {
		return "", errors.New("version is not valid")
	}
	return majMin, nil
}

func (f *RepoFiles) DockerfileChart(name string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(f.Dockerfile)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "/charts/"+name+".yaml") {
			if match := chartExp.FindStringSubmatch(line); len(match) > 1 {
				return match[1], nil
			}
		}
	}
	return "", errors.New("version not found in Dockerfile")
}

func (f *RepoFiles) DockerfileLayer(name string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(f.Dockerfile)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, name) {
			if match := layerExp.FindStringSubmatch(line); len(match) > 1 {
				return match[1], nil
			}
		}
	}
	return "", errors.New("version not found in Dockerfile")
}

func goDependency(modFile *modfile.File, name string) (string, error) {
	for _, replace := range modFile.Replace {
		if strings.Contains(replace.Old.Path, name) {
			return replace.New.Version, nil
		}
	}
	for _, require := range modFile.Require {
		if strings.Contains(require.Mod.Path, name) {
			return require.Mod.Version, nil
		}
	}
	return "", errors.New("library not found")
}

func (f *RepoFiles) GoDependency(name string) (string, error) {
	modFile, err := modfile.Parse("go.mod", f.ModFile, nil)
	if err != nil {
		return "", err
	}
	return goDependency(modFile, name)
}

func (f *RepoFiles) BuildVersion(name string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(f.VersionScript)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, name) {
			if match := buildScriptExp.FindStringSubmatch(line); len(match) > 1 {
				return match[1], nil
			}
		}
	}
	return "", errors.New("image not found in image list")
}

func (f *RepoFiles) Image(name string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(f.ImageList)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, name) {
			if match := imageListExp.FindStringSubmatch(line); len(match) > 1 {
				return match[1], nil
			}
		}
	}
	return "", errors.New("image not found in image list")
}

func (f *RepoFiles) SQLite() (string, error) {
	match := sqliteBindingExp.FindStringSubmatch(string(f.SQLiteBinding))
	if len(match) < 2 {
		return "", errors.New("sqlite version not found")
	}
	return match[1], nil
}

func (r *Repo) Calico() (*Component, error) {
	version, err := r.Files.Image("calico-node")
	if err != nil {
		return nil, err
	}
	shortVer, err := majMin(version)
	if err != nil {
		return nil, err
	}
	trimmed := trimPeriods(version)
	return &Component{
		Name:    "calico",
		Version: version,
		URL:     "[" + version + "](https://projectcalico.docs.tigera.io/archive/" + shortVer + "/release-notes/#" + trimmed + ")",
	}, nil
}

func (r *Repo) CanalCalico() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) Cilium() (*Component, error) {
	version, err := r.Files.Image("cilium/cilium")
	if err != nil {
		return nil, err
	}
	return &Component{
		Name:    "cilium",
		Version: version,
		URL:     "[" + version + "](https://github.com/cilium/cilium/releases/tag/" + version + ")",
	}, nil
}

func (r *Repo) Containerd() (*Component, error) {
	var version string
	if r.Name == "k3s" {
		v, err := r.Files.BuildVersion("VERSION_CONTAINERD")
		if err != nil {
			return nil, err
		}
		version = v
	} else if r.Name == "rke2" {
		v, err := r.Files.DockerfileLayer("hardened-containerd")
		if err != nil {
			return nil, err
		}
		version = v
	}
	return &Component{
		Name:    "containerd",
		Version: version,
		URL:     "[" + version + "](https://github.com/k3s-io/containerd/releases/tag/" + version + ")",
	}, nil
}

func (r *Repo) CoreDNS() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) Etcd() (*Component, error) {
	var version string
	if r.Name == "k3s" {
		v, err := r.Files.GoDependency("etcd/api/v3")
		if err != nil {
			return nil, err
		}
		version = v
	} else if r.Name == "rke2" {
		v, err := r.Files.BuildVersion("ETCD_VERSION")
		if err != nil {
			return nil, err
		}
		version = v
	}
	return &Component{
		Name:    "calico",
		Version: version,
		URL:     "[" + version + "](https://github.com/k3s-io/etcd/releases/tag/" + version + ")",
	}, nil
}

func (r *Repo) Flannel() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) HelmController() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) IngressNginx() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) Kine() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) Kubernetes() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) LocalPathProvisioner() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) MajorMinor() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) MetricsServer() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) Multus() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) Runc() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) SQLite() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) Traefik() (*Component, error) {
	return &Component{}, nil
}

func (r *Repo) CNIs() ([]*Component, error) {
	return []*Component{}, nil
}

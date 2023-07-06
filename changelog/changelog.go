package changelog

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/go-github/v39/github"
	"github.com/pkg/errors"
	"golang.org/x/mod/modfile"
)

type Component struct {
	Name          string
	Version       string
	URL           string
	FIPSCompliant bool
}

type Changelog interface {
	Calico() (*Component, error)
	CanalCalico() (*Component, error)
	Cilium() (*Component, error)
	Containerd() (*Component, error)
	CoreDNS() (*Component, error)
	Etcd() (*Component, error)
	Flannel() (*Component, error)
	HelmController() (*Component, error)
	IngressNginx() (*Component, error)
	Kine() (*Component, error)
	Kubernetes() (*Component, error)
	LocalPathProvisioner() (*Component, error)
	MajorMinor() (*Component, error)
	MetricsServer() (*Component, error)
	Multus() (*Component, error)
	Runc() (*Component, error)
	SQLite() (*Component, error)
	Traefik() (*Component, error)
	CNIs() ([]*Component, error)
	Issues(ctx context.Context, gh *github.Client) ([]*Issue, error)
}

type VersionParser interface {
	BuildVersion(variable string) (string, error)
	DockerfileChart(name string) (string, error)
	DockerfileLayer(name string) (string, error)
	GoDependency(name string) (string, error)
	Image(name string) (string, error)
	SQLite() (string, error)
}

type Repo struct {
	Name         string
	Organization string
	Version      string
	Commits      []*github.RepositoryCommit
	Files        *RepoFiles
}

type RepoFiles struct {
	Dockerfile    []byte
	ImageList     []byte
	ModFile       []byte
	SQLiteBinding []byte
	VersionScript []byte
}

func MakeRepo(ctx context.Context, client *github.Client, org, repo, prevMilestone, milestone string) (*Repo, error) {
	if milestone == "" {
		return nil, fmt.Errorf("milestone is empty")
	}
	if prevMilestone == "" {
		return nil, fmt.Errorf("prevMilestone is empty")
	}
	comp, _, err := client.Repositories.CompareCommits(ctx, org, repo, prevMilestone, milestone, &github.ListOptions{})
	if err != nil {
		return nil, err
	}
	files, err := MakeRepoFiles(ctx, repo, milestone)
	if err != nil {
		return nil, err
	}
	return &Repo{
		Name:         repo,
		Organization: org,
		Version:      milestone,
		Commits:      comp.Commits,
		Files:        files,
	}, nil
}

func downloadFile(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return b, nil
}

func MakeRepoFiles(ctx context.Context, repo, version string) (*RepoFiles, error) {
	var url string

	// Dockerfile
	switch repo {
	case "k3s":
		url = "https://raw.githubusercontent.com/k3-io/k3s/" + version + "/Dockerfile.dapper"
	case "rke2":
		url = "https://raw.githubusercontent.com/rancher/rke2/" + version + "/Dockerfile"
	default:
		return nil, errors.New("unknown repo " + repo)
	}
	dockerfile, err := downloadFile(ctx, url)
	if err != nil {
		return nil, err
	}

	// Image list
	switch repo {
	case "k3s":
		url = "https://raw.githubusercontent.com/k3s-io/k3s/" + version + "/scripts/airgap/image-list.txt"
	case "rke2":
		url = "https://raw.githubusercontent.com/rancher/rke2/" + version + "/scripts/build-images"
	default:
		return nil, errors.New("unknown repo " + repo)
	}
	imageList, err := downloadFile(ctx, url)
	if err != nil {
		return nil, err
	}

	// Build versions
	switch repo {
	case "k3s":
		url = "https://raw.githubusercontent.com/k3s-io/k3s/" + version + "/scripts/version.sh"
	case "rke2":
		url = "https://raw.githubusercontent.com/rancher/rke2/" + version + "/scripts/version.sh"
	default:
		return nil, errors.New("unknown repo " + repo)
	}
	versionScript, err := downloadFile(ctx, url)
	if err != nil {
		return nil, err
	}

	// Go mod file
	switch repo {
	case "k3s":
		url = "https://raw.githubusercontent.com/k3s-io/k3s/" + version + "/go.mod"
	case "rke2":
		url = "https://raw.githubusercontent.com/rancher/rke2/" + version + "/go.mod"
	default:
		return nil, errors.New("unknown repo " + repo)
	}
	modFileBytes, err := downloadFile(ctx, url)
	if err != nil {
		return nil, err
	}
	modFile, err := modfile.Parse("go.mod", modFileBytes, nil)
	if err != nil {
		return nil, err
	}

	// SQLite binding file
	var sqliteBinding []byte
	if repo == "k3s" {
		ver, err := goDependency(modFile, "go-sqlite3")
		if err != nil {
			return nil, err
		}
		sqliteBinding, err = downloadFile(ctx, "https://raw.githubusercontent.com/mattn/go-sqlite3/"+ver+"/sqlite3-binding.h")
		if err != nil {
			return nil, err
		}
	}

	return &RepoFiles{
		Dockerfile:    dockerfile,
		ImageList:     imageList,
		VersionScript: versionScript,
		ModFile:       modFileBytes,
		SQLiteBinding: sqliteBinding,
	}, nil
}

func DoSomething1(c Changelog) error {
	return nil
}

func DoSomething2(ctx context.Context, gh *github.Client) error {
	c, err := MakeRepo(ctx, gh, "k3s-io", "k3s", "milestone", "prevMilestone")
	if err != nil {
		return err
	}
	return DoSomething1(c)

}

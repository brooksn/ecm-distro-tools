package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/google/go-containerregistry/pkg/name"
	reg "github.com/rancher/ecm-distro-tools/registry"
	"github.com/rancher/ecm-distro-tools/release"
	"github.com/rancher/ecm-distro-tools/release/rke2"
	"github.com/rancher/ecm-distro-tools/repository"
	"github.com/spf13/cobra"
)

const (
	ossRegistry = "docker.io"
)

func formatImageRef(ref name.Reference) string {
	return ref.Context().RepositoryStr() + ":" + ref.Identifier()
}

func getStatus(expected, supported bool) string {
	const (
		yes     = "✓"
		no      = "✗"
		skipped = "-"
	)
	if !expected {
		return skipped
	}
	return map[bool]string{true: yes, false: no}[supported]
}

func getRegistryStatus(img reg.Image, expectsAmd64, expectsArm64 bool) string {
	const (
		yes  = "✓"
		no   = "✗"
		warn = "!"
	)
	if !img.Exists {
		return no
	}

	hasAllArch := true
	if expectsAmd64 {
		hasAllArch = hasAllArch && img.Platforms[reg.Platform{OS: "linux", Architecture: "amd64"}]
	}
	if expectsArm64 {
		hasAllArch = hasAllArch && img.Platforms[reg.Platform{OS: "linux", Architecture: "arm64"}]
	}
	return map[bool]string{true: yes, false: warn}[hasAllArch]
}

func table(w io.Writer, results []rke2.Image) {
	sort.Slice(results, func(i, j int) bool {
		return formatImageRef(results[i].Reference) < formatImageRef(results[j].Reference)
	})

	missingCount := 0
	for _, result := range results {
		if !result.OSSImage.Exists || !result.PrimeImage.Exists {
			missingCount++
		}
	}
	if missingCount > 0 {
		fmt.Fprintln(w, missingCount, "incomplete images")
	} else {
		fmt.Fprintln(w, "all images OK")
	}

	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	defer tw.Flush()

	fmt.Fprintln(tw, "image\toss\tprime\tsig\tamd64\tarm64\twin")
	fmt.Fprintln(tw, "-----\t---\t-----\t---\t-----\t-----\t---")

	for _, result := range results {
		tw.Write([]byte(strings.Join([]string{
			formatImageRef(result.Reference),
			result.OSSStatus().String(),
			result.PrimeStatus().String(),
			result.SigStatus().String(),
			result.AMD64Status().String(),
			result.ARM64Status().String(),
			result.WindowsStatus().String(),
			"",
		}, "\t") + "\n"))
	}
}

func csv(w io.Writer, results []rke2.Image) {
	sort.Slice(results, func(i, j int) bool {
		return formatImageRef(results[i].Reference) < formatImageRef(results[j].Reference)
	})

	fmt.Fprintln(w, "image,oss,prime,sig,amd64,arm64,win")

	for _, result := range results {
		values := []string{
			formatImageRef(result.Reference),
			result.OSSStatus().String(),
			result.PrimeStatus().String(),
			result.SigStatus().String(),
			result.AMD64Status().String(),
			result.ARM64Status().String(),
			result.WindowsStatus().String(),
		}
		fmt.Fprintln(w, strings.Join(values, ","))
	}
}

var inspectCmd = &cobra.Command{
	Use:   "inspect [version]",
	Short: "Inspect release artifacts",
	Long: `Inspect release artifacts for a given version.
Currently supports inspecting the image list for published rke2 releases.
- Availability in Docker Hub (OSS) and Prime registry
- Prime image signature status
- Image availability for linux/amd64, linux/arm64, and windows/amd64
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("expected at least one argument: [version]")
		}

		ctx := context.Background()
		gh := repository.NewGithub(ctx, rootConfig.Auth.GithubToken)
		filesystem, err := release.NewFS(ctx, gh, "rancher", "rke2", args[0])
		if err != nil {
			return err
		}

		ossClient := reg.NewClient(ossRegistry, debug)

		var primeClient *reg.Client
		if rootConfig.PrimeRegistry != "" {
			primeClient = reg.NewClient(rootConfig.PrimeRegistry, debug)
		}

		inspector := rke2.NewReleaseInspector(filesystem, ossClient, primeClient, debug)

		results, err := inspector.InspectRelease(ctx, args[0])
		if err != nil {
			return err
		}

		outputFormat, _ := cmd.Flags().GetString("output")
		switch outputFormat {
		case "csv":
			csv(os.Stdout, results)
		default:
			table(os.Stdout, results)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
	inspectCmd.Flags().StringP("output", "o", "table", "Output format (table|csv)")
}

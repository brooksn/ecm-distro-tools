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

func table(w io.Writer, results []rke2.Image, wide bool) {
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

	if wide {
		fmt.Fprintln(tw, "image\toss\tprime\tsig\tamd64\tarm64\twin")
		fmt.Fprintln(tw, "-----\t---\t-----\t---\t-----\t-----\t---")
	} else {
		fmt.Fprintln(tw, "image\toss\tprime")
		fmt.Fprintln(tw, "-----\t---\t-----")
	}

	for _, result := range results {
		columns := []string{
			formatImageRef(result.Reference),
			result.OSSStatus().String(),
			result.PrimeStatus().String(),
		}
		if wide {
			columns = append(columns,
				result.SigStatus().String(),
				result.AMD64Status().String(),
				result.ARM64Status().String(),
				result.WindowsStatus().String(),
			)
		}
		tw.Write([]byte(strings.Join(append(columns, ""), "\t") + "\n"))
	}
}

func csv(w io.Writer, results []rke2.Image, wide bool) {
	sort.Slice(results, func(i, j int) bool {
		return formatImageRef(results[i].Reference) < formatImageRef(results[j].Reference)
	})

	if wide {
		fmt.Fprintln(w, "image,oss,prime,sig,amd64,arm64,win")
	} else {
		fmt.Fprintln(w, "image,oss,prime")
	}

	for _, result := range results {
		values := []string{
			formatImageRef(result.Reference),
			result.OSSStatus().String(),
			result.PrimeStatus().String(),
		}
		if wide {
			values = append(values,
				result.SigStatus().String(),
				result.AMD64Status().String(),
				result.ARM64Status().String(),
				result.WindowsStatus().String(),
			)
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
		wide, _ := cmd.Flags().GetBool("wide")
		switch outputFormat {
		case "csv":
			csv(os.Stdout, results, wide)
		default:
			table(os.Stdout, results, wide)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
	inspectCmd.Flags().StringP("output", "o", "table", "Output format (table|csv)")
	inspectCmd.Flags().BoolP("wide", "w", false, "Show all columns including sig, amd64, arm64, and win")
}

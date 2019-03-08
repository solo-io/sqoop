package install

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/cliutil/install"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/sqoop/cli/pkg/flagutils"
	"github.com/solo-io/sqoop/version"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/helm/pkg/chartutil"
	helmhooks "k8s.io/helm/pkg/hooks"
	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/renderutil"

	"github.com/solo-io/sqoop/cli/pkg/options"
	"github.com/spf13/cobra"
)

func getSqoopVersion(opts *options.Options) (string, error) {
	if !version.IsReleaseVersion() && opts.Install.HelmChartOverride == "" {
		return "", errors.Errorf("you must provide a Sqoop Helm chart URI via the 'file' option " +
			"when running an unreleased version of Sqoop")
	}
	return version.Version, nil
}

func KubeCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kube",
		Short: "install sqoop on kubernetes",
		Long:  "requires kubectl to be installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			sqoopVersion, err := getSqoopVersion(opts)
			if err != nil {
				return err
			}

			// Get location of Gloo helm chart
			helmChartArchiveUri := fmt.Sprintf(sqoopTemplateUrl, sqoopVersion)
			if helmChartOverride := opts.Install.HelmChartOverride; helmChartOverride != "" {
				helmChartArchiveUri = helmChartOverride
			}

			if path.Ext(helmChartArchiveUri) != ".tgz" && !strings.HasSuffix(helmChartArchiveUri, ".tar.gz") {
				return errors.Errorf("unsupported file extension for Helm chart URI: [%s]. Extension must "+
					"either be .tgz or .tar.gz", helmChartArchiveUri)
			}

			chart, err := install.GetHelmArchive(helmChartArchiveUri)
			if err != nil {
				return errors.Wrapf(err, "retrieving sqoop helm chart archive")
			}

			values, err := install.GetValuesFromFile(chart, "")
			if err != nil {
				return errors.Wrapf(err, "retrieving value file: %s", "")
			}

			// These are the .Release.* variables used during rendering
			renderOpts := renderutil.Options{
				ReleaseOptions: chartutil.ReleaseOptions{
					Namespace: opts.Install.Namespace,
					Name:      "sqoop",
				},
			}

			// FILTER FUNCTION 2: Keep only CRDs and collect the names
			var crdNames []string
			filterCrds := func(input []manifest.Manifest) ([]manifest.Manifest, error) {

				var crdManifests []manifest.Manifest
				for _, man := range input {

					// Split manifest into individual YAML docs
					crdDocs := make([]string, 0)
					for _, doc := range strings.Split(man.Content, "---") {

						// We need to define this ourselves, because if we unmarshal into `apiextensions.CustomResourceDefinition`
						// we don't get the TypeMeta (in the yaml they are nested under `metadata`, but the k8s struct has
						// them as top level fields...)
						var resource struct {
							Metadata v1.ObjectMeta
							v1.TypeMeta
						}
						if err := yaml.Unmarshal([]byte(doc), &resource); err != nil {
							return nil, errors.Wrapf(err, "parsing resource: %s", doc)
						}

						// Skip non-CRD resources
						if resource.TypeMeta.Kind != install.CrdKindName {
							continue
						}

						// Check whether the CRD is a Helm "crd-install" hook.
						// If not, throw an error, because this will cause race conditions when installing with Helm (which is
						// not the case here, but we want to validate the manifests whenever we have the chance)
						helmCrdInstallHookAnnotation, ok := resource.Metadata.Annotations[helmhooks.HookAnno]
						if !ok || helmCrdInstallHookAnnotation != helmhooks.CRDInstall {
							return nil, errors.Errorf("CRD [%s] must be annotated as a Helm '%s' hook", resource.Metadata.Name, helmhooks.CRDInstall)
						}

						// Keep track of the CRD name
						crdNames = append(crdNames, resource.Metadata.Name)
						crdDocs = append(crdDocs, doc)
					}
					crdManifests = append(crdManifests, manifest.Manifest{
						Name:    man.Name,
						Head:    man.Head,
						Content: strings.Join(crdDocs, install.YamlDocumentSeparator),
					})
				}

				return crdManifests, nil
			}

			// Render and install CRD manifests
			crdManifestBytes, err := install.RenderChart(chart, values, renderOpts,
				install.ExcludeNotes,
				filterCrds,
				install.ExcludeEmptyManifests)
			if err != nil {
				return errors.Wrapf(err, "rendering crd manifests")
			}
			if err := install.InstallManifest(crdManifestBytes, opts.Install.DryRun); err != nil {
				return errors.Wrapf(err, "installing crd manifests")
			}

			// Only run if this is not a dry run
			if !opts.Install.DryRun {
				if err := install.WaitForCrdsToBeRegistered(crdNames, time.Second*5, time.Millisecond*500); err != nil {
					return errors.Wrapf(err, "waiting for crds to be registered")
				}
			}

			// Render and install Gloo manifest
			manifestBytes, err := install.RenderChart(chart, values, renderOpts,
				install.ExcludeNotes,
				install.ExcludeCrds,
				install.ExcludeEmptyManifests)
			if err != nil {
				return err
			}
			return install.InstallManifest(manifestBytes, opts.Install.DryRun)
		},
	}
	pflags := cmd.PersistentFlags()
	flagutils.AddInstallFlags(pflags, &opts.Install)

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

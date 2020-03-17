package resolvermap

import (
	"fmt"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/rest"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/sqoop/cli/pkg/common"
	"github.com/solo-io/sqoop/cli/pkg/flagutils"
	"github.com/solo-io/sqoop/cli/pkg/helpers"
	"github.com/solo-io/sqoop/cli/pkg/options"
	v1 "github.com/solo-io/sqoop/pkg/api/v1"
	"github.com/spf13/cobra"
)

func Register(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register TypeName FieldName -f resolver.yaml [-s schema-name]",
		Short: "Register a resolver for a field in your Schema",
		Long: `Sets the resolver for a field in your schema. TypeName.FieldName will always be resolved using this resolver 
Resolvers must be defined in yaml format. 
See the documentation at https://sqoop.solo.io/v1/resolver_map/#sqoop.api.v1.Resolver for the API specification for Sqoop Resolvers`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.SetResourceName(&opts.Metadata, args); err != nil {
				return err
			}
			if err := ensureResolverMapParams(&opts.ResolverMap, args); err != nil {
				return err
			}
			if err := registerResolver(opts); err != nil {
				return err
			}
			fmt.Println("resolver map registered successfully")
			return nil
		},
		Args: ensureResolverMapArgs,
	}

	flagutils.AddResolverMapFlags(cmd.PersistentFlags(), &opts.ResolverMap)

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func registerResolver(opts *options.Options) error {
	client, err := helpers.ResolverMapClient()
	if err != nil {
		return err
	}

	var responseTemplate *v1.ResponseTemplate
	if opts.ResolverMap.ResponseTemplate != "" {
		responseTemplate = &v1.ResponseTemplate{
			Body: opts.ResolverMap.ResponseTemplate,
		}
	}
	var requestTemplate *v1.RequestTemplate
	if opts.ResolverMap.RequestTemplate != "" {
		requestTemplate = &v1.RequestTemplate{
			Body: opts.ResolverMap.RequestTemplate,
		}
	}

	resolver := &v1.FieldResolver{
		Resolver: &v1.FieldResolver_GlooResolver{
			GlooResolver: &v1.GlooResolver{
				ResponseTemplate: responseTemplate,
				RequestTemplate:  requestTemplate,
				Action: &gloov1.RouteAction{
					Destination: &gloov1.RouteAction_Single{
						Single: &gloov1.Destination{
							DestinationType: &gloov1.Destination_Upstream{
								Upstream: &core.ResourceRef{
									Name:      opts.ResolverMap.Upstream,
									Namespace: opts.Metadata.Namespace,
								},
							},
							DestinationSpec: &gloov1.DestinationSpec{
								DestinationType: &gloov1.DestinationSpec_Rest{
									Rest: &rest.DestinationSpec{
										FunctionName: opts.ResolverMap.Function,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	existingResolverMap, err := resolverMapForSchema(opts, client)
	if err != nil {
		return err
	}
	existingResolverMap.Types[opts.ResolverMap.TypeName].Fields[opts.ResolverMap.FieldName] = resolver
	_, err = client.Write(existingResolverMap, clients.WriteOpts{OverwriteExisting: true})
	if err != nil {
		return err
	}
	return nil
}

func resolverMapForSchema(opts *options.Options, client v1.ResolverMapClient) (*v1.ResolverMap, error) {
	resolverMaps, err := client.List(opts.Metadata.Namespace, clients.ListOpts{})
	if err != nil {
		return nil, err
	}

	for _, rm := range resolverMaps {
		typResolver, ok := rm.Types[opts.ResolverMap.TypeName]
		if !ok {
			continue
		}
		if _, ok := typResolver.Fields[opts.ResolverMap.FieldName]; !ok {
			continue
		}
		return rm, nil
	}
	return nil, fmt.Errorf("cannot find a resolver map for type %v with field %v",
		opts.ResolverMap.TypeName,
		opts.ResolverMap.FieldName)
}

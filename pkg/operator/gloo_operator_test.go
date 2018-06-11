package operator_test

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/api/types/v1"
	"github.com/solo-io/gloo/pkg/storage"
	"github.com/solo-io/gloo/pkg/storage/file"
	. "github.com/solo-io/qloo/pkg/operator"
	"github.com/solo-io/qloo/test"
)

var _ = Describe("GlooOperator", func() {
	var (
		tmpDir                 string
		gloo                   storage.Interface
		vServiceName, roleName = "test-virtualservice", "test-role"
		operator               *GlooOperator
	)
	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		gloo, err = file.NewStorage(tmpDir, time.Millisecond)
		Expect(err).NotTo(HaveOccurred())
		err = gloo.V1().Register()
		Expect(err).NotTo(HaveOccurred())
		operator = NewGlooOperator(gloo, vServiceName, roleName)
	})
	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})
	It("creates the virtualservice with all the required routes", func() {
		operator.ApplyResolvers(test.StarWarsResolverMap())
		err := operator.ConfigureGloo()
		Expect(err).NotTo(HaveOccurred())
		virtualService, err := gloo.V1().VirtualServices().Get(vServiceName)
		Expect(err).NotTo(HaveOccurred())
		Expect(virtualService.Routes).To(HaveLen(5))
		Expect(virtualService.Routes[0]).To(Equal(&v1.Route{
			Matcher: &v1.Route_RequestMatcher{
				RequestMatcher: &v1.RequestMatcher{
					Path: &v1.RequestMatcher_PathExact{
						PathExact: "/Droid.friends",
					},
					Verbs: []string{
						"POST",
					},
				},
			},
			SingleDestination: &v1.Destination{
				DestinationType: &v1.Destination_Function{
					Function: &v1.FunctionDestination{
						UpstreamName: "starwars-rest",
						FunctionName: "GetCharacters",
					},
				},
			},
		}))
		Expect(virtualService.Routes[1]).To(Equal(&v1.Route{
			Matcher: &v1.Route_RequestMatcher{
				RequestMatcher: &v1.RequestMatcher{
					Path: &v1.RequestMatcher_PathExact{
						PathExact: "/Human.friends",
					},
					Verbs: []string{
						"POST",
					},
				},
			},
			SingleDestination: &v1.Destination{
				DestinationType: &v1.Destination_Function{
					Function: &v1.FunctionDestination{
						UpstreamName: "starwars-rest",
						FunctionName: "GetCharacters",
					},
				},
			},
		}))
		Expect(virtualService.Routes[2]).To(Equal(&v1.Route{
			Matcher: &v1.Route_RequestMatcher{
				RequestMatcher: &v1.RequestMatcher{
					Path: &v1.RequestMatcher_PathExact{
						PathExact: "/Query.droid",
					},
					Verbs: []string{
						"POST",
					},
				},
			},
			SingleDestination: &v1.Destination{
				DestinationType: &v1.Destination_Function{
					Function: &v1.FunctionDestination{
						UpstreamName: "starwars-rest",
						FunctionName: "GetCharacter",
					},
				},
			},
			PrefixRewrite: "",
			Extensions:    (*types.Struct)(nil),
		}))
		Expect(virtualService.Routes[3]).To(Equal(&v1.Route{
			Matcher: &v1.Route_RequestMatcher{
				RequestMatcher: &v1.RequestMatcher{
					Path: &v1.RequestMatcher_PathExact{
						PathExact: "/Query.hero",
					},
					Verbs: []string{
						"POST",
					},
				},
			},
			SingleDestination: &v1.Destination{
				DestinationType: &v1.Destination_Function{
					Function: &v1.FunctionDestination{
						UpstreamName: "starwars-rest",
						FunctionName: "GetHero",
					},
				},
			},
			PrefixRewrite: "",
			Extensions:    (*types.Struct)(nil),
		}))
		Expect(virtualService.Routes[4]).To(Equal(&v1.Route{
			Matcher: &v1.Route_RequestMatcher{
				RequestMatcher: &v1.RequestMatcher{
					Path: &v1.RequestMatcher_PathExact{
						PathExact: "/Query.human",
					},
					Verbs: []string{
						"POST",
					},
				},
			},
			SingleDestination: &v1.Destination{
				DestinationType: &v1.Destination_Function{
					Function: &v1.FunctionDestination{
						UpstreamName: "starwars-rest",
						FunctionName: "GetCharacter",
					},
				},
			},
			PrefixRewrite: "",
			Extensions:    (*types.Struct)(nil),
		}))
		Expect(virtualService.Domains).To(HaveLen(1))
		Expect(virtualService.Domains[0]).To(Equal("*"))
		Expect(virtualService.Roles).To(HaveLen(1))
		Expect(virtualService.Roles[0]).To(Equal(roleName))
	})
})

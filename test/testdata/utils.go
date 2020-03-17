package testdata

import (
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/options/rest"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/sqoop/pkg/api/v1"
	"github.com/solo-io/sqoop/pkg/engine/exec"
	"github.com/solo-io/sqoop/pkg/engine/resolvers"
	"github.com/solo-io/sqoop/pkg/translator"
	"github.com/vektah/gqlgen/graphql"
	"github.com/vektah/gqlgen/neelance/schema"
)

const resolversName = "starwars-resolvers"
const namespace = "default"

var resolverMetadata = core.Metadata{
	Name:      resolversName,
	Namespace: namespace,
}

var StarWarsSchema = schema.MustParse(starWarsSchemaString)
var x rest.DestinationSpec

func StarWarsResolverMap() *v1.ResolverMap {
	resolverMap := translator.GenerateResolverMapSkeleton(resolverMetadata, StarWarsSchema)
	resolverMap.Types["Query"].Fields["hero"].Resolver = &v1.FieldResolver_GlooResolver{
		GlooResolver: &v1.GlooResolver{
			Action: &glooV1.RouteAction{
				Destination: &gloov1.RouteAction_Single{
					Single: &gloov1.Destination{
						Upstream: core.ResourceRef{
							Name: "starwars-rest",
						},
						DestinationSpec: &gloov1.DestinationSpec{
							DestinationType: &gloov1.DestinationSpec_Rest{
								Rest: &gloov1.DestinationSpec{
									FunctionName: "GetHero",
								},
							},
						},
					},
				},
			},
		},
	}
	resolverMap.Types["Query"].Fields["human"].Resolver = &v1.FieldResolver_GlooResolver{
		GlooResolver: &v1.GlooResolver{
			RequestTemplate: &v1.RequestTemplate{
				Body: `{"id": {{ index .Args "id" }}}`,
			},
			Action: &gloov1.RouteAction{
				Destination: &gloov1.RouteAction_Single{
					Single: &gloov1.Destination{
						Upstream: core.ResourceRef{
							Name: "starwars-rest",
						},
						DestinationSpec: &gloov1.DestinationSpec{
							DestinationType: &gloov1.DestinationSpec_Rest{
								Rest: &rest.DestinationSpec{
									FunctionName: "GetCharacter",
								},
							},
						},
					},
				},
			},
		},
	}
	resolverMap.Types["Query"].Fields["droid"].Resolver = &v1.FieldResolver_GlooResolver{
		GlooResolver: &v1.GlooResolver{
			RequestTemplate: &v1.RequestTemplate{
				Body: `{"id": {{ index .Args "id" }}}`,
			},
			Action: &glooV1.RouteAction{
				Destination: &glooV1.RouteAction_Single{
					Single: &glooV1.Destination{
						Upstream: core.ResourceRef{
							Name: "starwars-rest",
						},
						DestinationSpec: &glooV1.DestinationSpec{
							DestinationType: &glooV1.DestinationSpec_Rest{
								Rest: &rest.DestinationSpec{
									FunctionName: "GetCharacter",
								},
							},
						},
					},
				},
			},
		},
	}
	resolverMap.Types["Droid"].Fields["friends"].Resolver = &v1.FieldResolver_GlooResolver{
		GlooResolver: &v1.GlooResolver{
			RequestTemplate: &v1.RequestTemplate{
				Body: `{{ marshal (index .Parent "friend_ids") }}`,
			},
			Action: &glooV1.RouteAction{
				Destination: &glooV1.RouteAction_Single{
					Single: &glooV1.Destination{
						Upstream: core.ResourceRef{
							Name: "starwars-rest",
						},
						DestinationSpec: &glooV1.DestinationSpec{
							DestinationType: &glooV1.DestinationSpec_Rest{
								Rest: &rest.DestinationSpec{
									FunctionName: "GetCharacters",
								},
							},
						},
					},
				},
			},
		},
	}
	resolverMap.Types["Human"].Fields["appearsIn"].Resolver = &v1.FieldResolver_TemplateResolver{
		TemplateResolver: &v1.TemplateResolver{
			InlineTemplate: `{{ index .Parent "appears_in" }}}`,
		},
	}
	resolverMap.Types["Droid"].Fields["appearsIn"].Resolver = &v1.FieldResolver_TemplateResolver{
		TemplateResolver: &v1.TemplateResolver{
			InlineTemplate: `{{ index .Parent "appears_in" }}}`,
		},
	}
	return resolverMap
}

func StarWarsV1Schema() *v1.Schema {
	return &v1.Schema{
		Metadata:     resolverMetadata,
		InlineSchema: starWarsSchemaString,
	}
}

func StarWarsExecutableSchema(proxyAddr string) graphql.ExecutableSchema {
	execResolvers := StarWarsExecutableResolvers(proxyAddr)
	return exec.NewExecutableSchema(StarWarsSchema, execResolvers)
}

func StarWarsExecutableResolvers(proxyAddr string) *exec.ExecutableResolverMap {
	factory := StarWarsResolverFactory(proxyAddr)
	execResolvers, err := exec.NewExecutableResolvers(StarWarsSchema, factory.CreateResolver)
	if err != nil {
		panic(err)
	}
	return execResolvers
}

func StarWarsResolverFactory(proxyAddr string) *resolvers.ResolverFactory {
	return resolvers.NewResolverFactory(proxyAddr, StarWarsResolverMap())
}

var starWarsSchemaString = `# The query type, represents all of the entry points into our object graph
type Query {
    hero(episode: Episode = NEWHOPE): Character
    reviews(episode: Episode!, since: Time): [Review]!
    search(text: String!): [SearchResult]!
    character(id: ID!): Character
    droid(id: ID!): Droid
    human(id: ID!): Human
    starship(id: ID!): Starship
}
# The mutation type, represents all updates we can make to our data
type Mutation {
    createReview(episode: Episode!, review: ReviewInput!): Review
}
# The episodes in the Star Wars trilogy
enum Episode {
    # Star Wars Episode IV: A New Hope, released in 1977.
    NEWHOPE
    # Star Wars Episode V: The Empire Strikes Back, released in 1980.
    EMPIRE
    # Star Wars Episode VI: Return of the Jedi, released in 1983.
    JEDI
}
# A character from the Star Wars universe
interface Character {
    # The ID of the character
    id: ID!
    # The name of the character
    name: String!
    # The friends of the character, or an empty list if they have none
    friends: [Character]
    # The friends of the character exposed as a connection with edges
    friendsConnection(first: Int, after: ID): FriendsConnection!
    # The movies this character appears in
    appearsIn: [Episode!]!
}
# Units of height
enum LengthUnit {
    # The standard unit around the world
    METER
    # Primarily used in the United States
    FOOT
}
# A humanoid creature from the Star Wars universe
type Human implements Character {
    # The ID of the human
    id: ID!
    # What this human calls themselves
    name: String!
    # Height in the preferred unit, default is meters
    height(unit: LengthUnit = METER): Float!
    # Mass in kilograms, or null if unknown
    mass: Float
    # This human` + "`" + `s friends, or an empty list if they have none
    friends: [Character]
    # The friends of the human exposed as a connection with edges
    friendsConnection(first: Int, after: ID): FriendsConnection!
    # The movies this human appears in
    appearsIn: [Episode!]!
    # A list of starships this person has piloted, or an empty list if none
    starships: [Starship]
}
# An autonomous mechanical character in the Star Wars universe
type Droid implements Character {
    # The ID of the droid
    id: ID!
    # What others call this droid
    name: String!
    # This droid` + "`" + `s friends, or an empty list if they have none
    friends: [Character]
    # The friends of the droid exposed as a connection with edges
    friendsConnection(first: Int, after: ID): FriendsConnection!
    # The movies this droid appears in
    appearsIn: [Episode!]!
    # This droid` + "`" + `s primary function
    primaryFunction: String
}
# A connection object for a character` + "`" + `s friends
type FriendsConnection {
    # The total number of friends
    totalCount: Int!
    # The edges for each of the character` + "`" + `s friends.
    edges: [FriendsEdge]
    # A list of the friends, as a convenience when edges are not needed.
    friends: [Character]
    # Information for paginating this connection
    pageInfo: PageInfo!
}
# An edge object for a character` + "`" + `s friends
type FriendsEdge {
    # A cursor used for pagination
    cursor: ID!
    # The character represented by this friendship edge
    node: Character
}
# Information for paginating this connection
type PageInfo {
    startCursor: ID!
    endCursor: ID!
    hasNextPage: Boolean!
}
# Represents a review for a movie
type Review {
    # The number of stars this review gave, 1-5
    stars: Int!
    # Comment about the movie
    commentary: String
    # when the review was posted
    time: Time
}
# The input object sent when someone is creating a new review
input ReviewInput {
    # 0-5 stars
    stars: Int!
    # Comment about the movie, optional
    commentary: String
    # when the review was posted
    time: Time
}
type Starship {
    # The ID of the starship
    id: ID!
    # The name of the starship
    name: String!
    # Length of the starship, along the longest axis
    length(unit: LengthUnit = METER): Float!
    # coordinates tracking this ship
    history: [[Int]]
}
union SearchResult = Human | Droid | Starship
scalar Time
`

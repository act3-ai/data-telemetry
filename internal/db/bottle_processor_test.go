package db

import (
	"net/url"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	latest "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io/v1"
	"gitlab.com/act3-ai/asce/data/schema/pkg/selectors"
)

type BottleProcessorTestSuite struct {
	suite.Suite
}

func (s *BottleProcessorTestSuite) SetupSuite() {
}

func (s *BottleProcessorTestSuite) TestSourceConvertWithPartSelectors() {
	partSelectors, err := selectors.Parse([]string{"partkey!=value1,mykey=value2", "partkey2=45"})
	s.NoError(err)
	bottleDigest256Str := "beefefd3483c60a4364d3f6f328ee1897facdbffb043b51941424a34121bbbe9"
	bottleDigest, err := digest.Parse("sha256:" + bottleDigest256Str)
	s.NoError(err)

	uri1QueryParams := url.Values{}
	uri1QueryParams.Add("selector", "partkey!=value1,mykey=value2")
	uri1QueryParams.Add("selector", "partkey2=45")
	dtoSource1 := &latest.Source{
		Name: "test-source",
		URI:  "bottle:sha256:" + bottleDigest256Str + "?" + uri1QueryParams.Encode(),
	}
	oldSource1 := &Source{
		BottleMemberLocated: BottleMemberLocated{
			Model: Model{
				gorm.Model{
					ID: 1,
				},
			},
		},
	}

	expectedSource1 := &Source{
		BottleMemberLocated: BottleMemberLocated{
			Model: Model{
				gorm.Model{
					ID: 1,
				},
			},
			Location: 1,
		},
		Name:          dtoSource1.Name,
		URI:           dtoSource1.URI,
		BottleDigest:  bottleDigest,
		PartSelectors: partSelectors,
	}

	uri2QueryParams := url.Values{}
	uri2QueryParams.Add("type", "application/vnd.act3-ace.bottle.config.v1+json")
	uri2QueryParams.Add("selector", "partkey!=value1,mykey=value2")
	uri2QueryParams.Add("selector", "partkey2=45")
	dtoSource2 := &latest.Source{
		Name: "test-source2",
		URI:  "hash://sha256/" + bottleDigest256Str + "?" + uri2QueryParams.Encode(),
	}
	oldSource2 := &Source{
		BottleMemberLocated: BottleMemberLocated{
			Model: Model{
				gorm.Model{
					ID: 2,
				},
			},
		},
	}
	expectedSource2 := &Source{
		BottleMemberLocated: BottleMemberLocated{
			Model: Model{
				gorm.Model{
					ID: 2,
				},
			},
			Location: 2,
		},
		Name:          dtoSource2.Name,
		URI:           dtoSource2.URI,
		BottleDigest:  bottleDigest,
		PartSelectors: partSelectors,
	}

	outputSource1, err := convertSource(*oldSource1, 1, *dtoSource1)
	s.NoError(err)
	outputSource2, err := convertSource(*oldSource2, 2, *dtoSource2)
	s.NoError(err)
	checkFunc := func(srcTest Source, srcExpected Source) {
		s.Equal(srcExpected.Name, srcTest.Name)
		s.Equal(srcExpected.URI, srcTest.URI)
		s.Equal(srcExpected.PartSelectors, srcTest.PartSelectors)
	}
	checkFunc(*outputSource1, *expectedSource1)
	checkFunc(*outputSource2, *expectedSource2)
}

func TestBottleProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(BottleProcessorTestSuite))
}

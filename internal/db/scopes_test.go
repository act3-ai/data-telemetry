package db

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/act3-ai/go-common/pkg/redact"
	bottle "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io"
	"gitlab.com/act3-ai/asce/data/schema/pkg/selectors"

	"github.com/act3-ai/data-telemetry/v3/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
)

type ScopesTestSuite struct {
	suite.Suite
	ctx     context.Context
	con     *gorm.DB
	dataDir string
}

func (s *ScopesTestSuite) SetupTest() {
	s.ctx = context.Background()
	// set up db connection
	scheme := runtime.NewScheme()
	s.NoError(bottle.AddToScheme(scheme))
	dsn := "file::memory:"
	myDB, err := Open(s.ctx, v1alpha2.Database{
		DSN: redact.SecretURL(dsn),
	}, scheme)
	s.NoError(err)
	s.con = myDB

	s.dataDir = filepath.Join("..", "..", "testdata")
}

func (s *ScopesTestSuite) TestPartFilter() {
	partDigestABC := digest.FromBytes([]byte("abc"))
	partDigestJKL := digest.FromBytes([]byte("jkl"))
	partDigestXYZ := digest.FromBytes([]byte("xyz"))

	testPart1 := Part{
		BottleMemberLocated: BottleMemberLocated{
			BottleID: 1,
		},
		Name:   "TestPart1",
		Size:   1000,
		Digest: partDigestABC,
	}

	testPart2 := Part{
		BottleMemberLocated: BottleMemberLocated{
			BottleID: 2,
		},
		Name:   "TestPart2",
		Size:   1000,
		Digest: partDigestJKL,
	}

	testPart3 := Part{
		Name:   "TestPart3",
		Size:   1000,
		Digest: partDigestXYZ,
	}

	bottle1 := &Bottle{
		Base: Base{
			DataID: 1,
		},
		Description: "test bottle 1",
		Parts: []Part{
			testPart1,
			testPart2,
		},
	}

	dgst1 := digest.FromBytes([]byte("dgst1"))

	bottleDigest1 := &Digest{
		DataID: 1,
		Digest: dgst1,
	}

	s.NoError(s.con.Create(bottleDigest1).Error)
	s.NoError(s.con.Create(bottle1).Error)

	bottle2 := &Bottle{
		Base: Base{
			DataID: 2,
		},
		Description: "test bottle 2",
		Parts: []Part{
			testPart1,
			testPart2,
		},
	}

	dgst2 := digest.FromBytes([]byte("dgst2"))

	bottleDigest2 := &Digest{
		DataID: 2,
		Digest: dgst2,
	}

	s.NoError(s.con.Create(bottleDigest2).Error)
	s.NoError(s.con.Create(bottle2).Error)

	bottle3 := &Bottle{
		Base: Base{
			DataID: 3,
		},
		Description: "test bottle 3",
		Parts: []Part{
			testPart1,
			testPart3,
		},
	}

	dgst3 := digest.FromBytes([]byte("dgst3"))

	bottleDigest3 := &Digest{
		DataID: 3,
		Digest: dgst3,
	}

	s.NoError(s.con.Create(bottleDigest3).Error)
	s.NoError(s.con.Create(bottle3).Error)

	var partDigests []digest.Digest
	partDigests = append(partDigests, partDigestABC)
	partDigests = append(partDigests, partDigestJKL)
	tx := s.con.Table("bottles").
		Scopes(FilterByParts(partDigests)).
		Distinct("bottles.data_id").
		Order("bottles.data_id DESC")

	var entries []Bottle
	s.NoError(tx.Find(&entries).Error)

	s.EqualValues(len(entries), 2)
	s.EqualValues(int(entries[0].DataID), 2)
	s.EqualValues(int(entries[1].DataID), 1)
}

func (s *ScopesTestSuite) TestSelectorFilter() {
	bottle1 := &Bottle{
		Base: Base{
			DataID: 1,
		},
		Description: "test bottle 1",
		Labels: []Label{
			{
				Key:   "testKey",
				Value: "testValue",
			},
			{
				Key:   "otherTestKey",
				Value: "otherTestValue",
			},
		},
	}
	s.commitBottle(bottle1)

	bottle2 := &Bottle{
		Base: Base{
			DataID: 2,
		},
		Description: "test bottle 2",
		Labels: []Label{
			{
				Key:   "foo",
				Value: "bar",
			},
			{
				Key:          "numeric",
				NumericValue: 0,
			},
		},
	}

	s.commitBottle(bottle2)

	bottle3 := &Bottle{
		Base: Base{
			DataID: 3,
		},
		Description: "test bottle 3",
		Labels: []Label{
			{
				Key:   "foo",
				Value: "baz",
			},
			{
				Key:          "numeric",
				NumericValue: 2,
			},
		},
	}

	s.commitBottle(bottle3)

	bottle4 := &Bottle{
		Base: Base{
			DataID: 4,
		},
		Description: "test bottle 4",
		Labels:      []Label{},
	}

	s.commitBottle(bottle4)

	testBottles := []*Bottle{
		bottle1,
		bottle2,
		bottle3,
		bottle4,
	}

	type selectorTest struct {
		Name              string
		Selectors         []string
		ExpectedBottleIDs []int
		CheckFn           func([]Bottle)
	}

	selectorTests := []selectorTest{
		{
			Name:      "equalsSelector",
			Selectors: []string{"foo=bar"},
		}, {
			Name:      "notEqualsSelector",
			Selectors: []string{"foo!=bar"},
		}, {
			Name:      "inSelector",
			Selectors: []string{"foo in (bar, baz)"},
		}, {
			Name:      "notInSelector",
			Selectors: []string{"foo notin (bar, zab)"},
		}, {
			Name:      "existsSelector",
			Selectors: []string{"testKey"},
		}, {
			Name:      "notExistsSelector",
			Selectors: []string{"!testKey"},
		}, {
			Name:      "greaterThanSelector",
			Selectors: []string{"numeric > 1"},
		}, {
			Name:      "lessThanSelector",
			Selectors: []string{"numeric < 1"},
		}, {
			Name:      "multipleSelector",
			Selectors: []string{"!testKey", "foo!=bar"},
		},
	}

	for _, selectorTest := range selectorTests {
		tx := s.con.Table("bottles").
			Scopes(FilterBySelectors(selectorTest.Selectors))

		var entries []Bottle
		s.NoError(tx.Find(&entries).Error, "an error occurred when using selector \"%s\"", selectorTest.Name)

		// use the selectors to determine the bottles we expect to get back
		sel, err := selectors.Parse(selectorTest.Selectors)
		s.NoError(err)

		for _, btl := range testBottles {
			labelSet := make(map[string]string, len(btl.Labels))
			for _, lbl := range btl.Labels {
				if len(lbl.Value) == 0 {
					labelSet[lbl.Key] = fmt.Sprintf("%f", lbl.NumericValue)
				} else {
					labelSet[lbl.Key] = lbl.Value
				}
			}
			if sel.Matches(labels.Set(labelSet)) {
				selectorTest.ExpectedBottleIDs = append(selectorTest.ExpectedBottleIDs, int(btl.DataID))
			}
		}

		// Make sure we got back the bottles we expected to and none that we don't expect
		entryDataIDs := []int{}
		for _, entry := range entries {
			entryDataIDs = append(entryDataIDs, int(entry.DataID))
		}
		// Make sure every expected data ID is present in the entries we got back
		for _, expectedDataID := range selectorTest.ExpectedBottleIDs {
			s.Contains(entryDataIDs, expectedDataID, "expected bottle %d was not returned for selector test \"%s\"", expectedDataID, selectorTest.Name)
		}
		// Make sure every entry we got back is one of the expected data IDs
		for _, entryDataID := range entryDataIDs {
			s.Contains(selectorTest.ExpectedBottleIDs, entryDataID, "unexpected bottle %d returned for selector test \"%s\"", entryDataID, selectorTest.Name)
		}
	}
}

func (s *ScopesTestSuite) commitBottle(b *Bottle) {
	dgst := digest.FromString(fmt.Sprintf("%d", b.DataID))

	bottleDigest := &Digest{
		DataID: b.DataID,
		Digest: dgst,
	}

	s.NoError(s.con.Create(bottleDigest).Error)
	s.NoError(s.con.Create(b).Error)
}

func TestScopesTestSuite(t *testing.T) {
	suite.Run(t, new(ScopesTestSuite))
}

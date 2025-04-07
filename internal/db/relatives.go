package db

import (
	"strings"

	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"

	"github.com/act3-ai/bottle-schema/pkg/selectors"
)

/*
URI examples

bottle:sha256:05a8efd3483c60a4364d3f6f328ee1897facdbffb043b51941424a34121bbbe9#partkey!=value1,mykey=value2|partkey2=45
bottle:sha512:95a8efd3483c60a4364d3f6f328ee1897facdbffb043b51941424a34121bbbe905a8efd3483c60a4364d3f6f328ee1897facdbffb043b51941424a34121bbbe9#partkey!=value1,mykey=value2|partkey2=45
hash://sha256/05a8efd3483c60a4364d3f6f328ee1897facdbffb043b51941424a34121bbbe9?type=application/vnd.act3-ace.bottle.config.v1+json#partkey!=value1,mykey=value2|partkey2=45
bottle:sha256:05a8efd3483c60a4364d3f6f328ee1897facdbffb043b51941424a34121bbbe9
https://www.google.com

Bottle 1-4 are the same bottle.
*/

// Relative is a type constraint for any relative of a bottle.
type Relative interface {
	BottleRelative | NonBottleRelative | UnknownBottleRelative
}

// BottleRelative is a bottle relative to a bottle (parent, child, grand parent, etc.).
type BottleRelative struct { // vertex
	Bottle
	Digested
	PartSelectors selectors.LabelSelectorSet `gorm:"-"` // Optional if the relationship is partial
}

// MatchesSource returns true if the source matches this relative.
func (r BottleRelative) MatchesSource(source Source) bool {
	for _, d := range r.Digests {
		if d == source.BottleDigest {
			return true
		}
	}
	return false
}

// GetSourceByURI returns a pointer to a source if the given URI matches one of the sources on the BottleRelative.
// If no source is found, nil is returned.
func (r BottleRelative) GetSourceByURI(uri string) *Source {
	for _, s := range r.Sources {
		if s.URI == uri {
			return &s
		} else if strings.HasPrefix(s.URI, "hash") {
			if strings.Contains(s.URI, uri) {
				return &s
			}
		}
	}
	return nil
}

// GetSourceByRelative returns a pointer to a source if one of the given relative's digests matches one of the sources on the BottleRelative.
// If no source is found, nil is returned.
func (r BottleRelative) GetSourceByRelative(relative *BottleRelative) *Source {
	for _, d := range relative.Digests {
		source := r.GetSourceByURI(d.String())
		if source != nil {
			return source
		}
	}
	return nil
}

// NonBottleRelative represents a relative that is not a bottle but instead a URI of some other kind.
type NonBottleRelative struct {
	URI string
}

// MatchesSource returns true if the source matches this relative.
func (u NonBottleRelative) MatchesSource(source Source) bool {
	return string(u.URI) == source.URI
}

// UnknownBottleRelative represents a bottle that is not known to the telemetry server (so we know nothing more than its digest).
type UnknownBottleRelative struct {
	Digest        digest.Digest
	PartSelectors selectors.LabelSelectorSet
}

// MatchesSource returns true if the source matches this relative.
func (b UnknownBottleRelative) MatchesSource(source Source) bool {
	return digest.Digest(b.Digest) == source.BottleDigest
}

// Generation represents a generation of ancestors (e.g., parents, grandparents, children).
type Generation []BottleRelative

// GetDigests returns all the bottle digests for this generation (the result may contain duplicate digests).
func (g Generation) GetDigests() []digest.Digest {
	digests := make([]digest.Digest, 0, len(g))
	for _, relative := range g {
		digests = append(digests, relative.Digests...)
	}
	return digests
}

// getDigestSet returns a set of digests for fast lookup.
func (g Generation) getDigestSet() map[digest.Digest]bool {
	// make a set out of a the digests for fast lookup
	digests := g.GetDigests()
	digestSet := make(map[digest.Digest]bool, len(digests))
	for _, d := range digests {
		digestSet[d] = true
	}
	return digestSet
}

// GetUniqueDigests returns all the bottle digests for this generation (the result will not contain duplicate digests).
func (g Generation) GetUniqueDigests() []digest.Digest {
	digestSet := g.getDigestSet()
	digests := make([]digest.Digest, 0, len(digestSet))

	for d := range digestSet {
		digests = append(digests, d)
	}
	return digests
}

// GetNonBottleParents extracts the non-bottle URIs.
func (g Generation) GetNonBottleParents() []NonBottleRelative {
	nonBottles := []NonBottleRelative{}

	for _, relative := range g {
		for _, source := range relative.Sources {
			if len(source.BottleDigest) == 0 {
				// This is not a bottle so it must be a non-bottle
				nonBottles = append(nonBottles, NonBottleRelative(NonBottleRelative{URI: source.URI}))
			}
		}
	}

	return nonBottles
}

// GetUnknownBottleParents extracts the unknown bottle digests from a generation.
func (g Generation) GetUnknownBottleParents(knownParents Generation) []UnknownBottleRelative {
	unknownBottles := []UnknownBottleRelative{}

	// make a set out of the digests for fast lookup
	knownDigestSet := knownParents.getDigestSet()

	for _, relative := range g {
		for _, source := range relative.Sources {
			if len(source.BottleDigest) != 0 {
				// Test to see if the bottle is known
				if _, exists := knownDigestSet[source.BottleDigest]; !exists {
					unknownBottles = append(unknownBottles, UnknownBottleRelative(UnknownBottleRelative{Digest: source.BottleDigest, PartSelectors: source.PartSelectors}))
				}
			}
		}
	}

	return unknownBottles
}

// FindRelativeIdx finds the relative that matches the source in this generation.
func (g Generation) FindRelativeIdx(source Source) int {
	for i, r := range g {
		if r.MatchesSource(source) {
			// We found a bottle match
			return i
		}
	}
	return -1
}

// GetAncestors returns all ancestors
// Input: a bottle digest and depth (1 = parents, 2 = grandparents, etc.)
// Output: a double-slice where each outer slice index indicates the generation, i.e.
// 0 = parents, etc.
func GetAncestors(con *gorm.DB, bottleDigest digest.Digest, depth uint) ([]Generation, error) {
	return FindGenerations(con, bottleDigest, depth, FindParents)
}

// GetDescendants returns all descendants
// Input: a bottle digest and depth (1 = children, 2 = grandchildren, etc.)
// Output: a double-slice where each outer slice index indicates the generation, i.e.
// 0 = children, etc.
func GetDescendants(con *gorm.DB, bottleDigest digest.Digest, depth uint) ([]Generation, error) {
	return FindGenerations(con, bottleDigest, depth, FindChildren)
}

// NextGenerationFinder is a function that finds the next generation from the given digests (either up or down the family tree).
type NextGenerationFinder func(con *gorm.DB, digests []digest.Digest) (Generation, error)

// FindGenerations is a generic function to find multiple generations.
func FindGenerations(con *gorm.DB, bottleDigest digest.Digest, depth uint, finder NextGenerationFinder) ([]Generation, error) {
	generations := make([]Generation, 0, depth)

	digests := []digest.Digest{bottleDigest}
	for i := uint(0); i < depth; i++ {
		gen, err := finder(con, digests)
		if err != nil {
			return nil, err
		}
		generations = append(generations, gen)
		digests = gen.GetUniqueDigests()
	}
	return generations, nil
}

// FindParents finds all parents of the provides bottle digest (any number of bottles).
func FindParents(con *gorm.DB, digests []digest.Digest) (Generation, error) {
	// Get the bottles associated with the given digests
	originalBottles := []BottleRelative{}

	tx := con.Select("bottles.*").
		Table("bottles").
		Preload("Sources").
		Scopes(IncludeDigests("bottles"), FilterByDigests(digests, "bottles"))

	if err := tx.Find(&originalBottles).Error; err != nil {
		return nil, err
	}

	var sources []Source

	for _, b := range originalBottles {
		sources = append(sources, b.Sources...)
	}

	// Get the parents of the given digests
	gen := Generation{}

	tx = con.Select("bottles.*").
		Table("bottles").
		Preload("Labels").
		Preload("Sources").
		Scopes(IncludeDigests("bottles"), ParentsOf(digests), RankByNumPulls())

	if err := tx.Find(&gen).Error; err != nil {
		return nil, err
	}

	// Match up original bottles sources and ancestors to add part selectors for partial relationship
	for i, g := range gen {
		for _, s := range sources {
			for _, d := range g.Digests {
				if s.BottleDigest == d {
					gen[i].PartSelectors = s.PartSelectors
					continue
				}
			}
		}
	}

	return gen, nil
}

// FindChildren finds all children of the provides bottle digest (any number of bottles).
func FindChildren(con *gorm.DB, digests []digest.Digest) (Generation, error) {
	gen := Generation{}

	// TODO this is redundant with FindParents
	tx := con.Select("bottles.*").
		Table("bottles").
		Preload("Labels").
		Preload("Sources").
		Scopes(IncludeDigests("bottles"), ChildrenOf(digests), RankByNumPulls())

	if err := tx.Find(&gen).Error; err != nil {
		return nil, err
	}

	return gen, nil
}

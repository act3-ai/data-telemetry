package db

import (
	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"
)

// FindDeprecatedBy finds all bottles that are deprecated by the bottle with the given digest.
func FindDeprecatedBy(con *gorm.DB, dgst digest.Digest) ([]digest.Digest, error) {
	var deprecatedDigests []digest.Digest
	tx := con.Select("digests.digest").
		Table("bottles").
		Scopes(DeprecatedBy(dgst))

	if err := tx.Find(&deprecatedDigests).Error; err != nil {
		return nil, err
	}

	return deprecatedDigests, nil
}

// FindDeprecates finds all bottles that deprecates the bottle with the given digest.
func FindDeprecates(con *gorm.DB, dgst digest.Digest) ([]digest.Digest, error) {
	var deprecatesDigests []digest.Digest
	tx := con.Select("digests.digest").
		Table("bottles").
		Scopes(DeprecatesThis(dgst))

	if err := tx.Find(&deprecatesDigests).Error; err != nil {
		return nil, err
	}

	return deprecatesDigests, nil
}

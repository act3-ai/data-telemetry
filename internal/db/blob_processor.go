package db

import (
	"gorm.io/gorm"
)

// Blobs can be anything.  They currently are used to only store bottle's public artifacts but more can be stored if the schema grows.
// On the API side we really don't need the concept of an artifact (webapp handle all artifact requests).

// BlobProcessorVersion is the current version of the processor code.  This is incremented after each measurable change to the BlobProcessor().
const BlobProcessorVersion = 3

// BlobProcessor handles blob processing.
type BlobProcessor struct{}

// Version returns the processor version.
func (p *BlobProcessor) Version() uint {
	return BlobProcessorVersion
}

// PrimaryTable returns primary table that this processor updates.
func (p *BlobProcessor) PrimaryTable() string {
	return "blobs"
}

// Process converts Bottle data to the DB model for a Bottle.
func (p *BlobProcessor) Process(con *gorm.DB, base Base) error {
	// save to DB
	blob := Blob{
		Base: base,
	}

	return con.Save(&blob).Error
}

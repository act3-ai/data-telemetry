package db

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// BaseType is the interface for the base DB model functionality

// Processor knows how to process it's type to conver the raw data to a DB model.
type Processor interface {
	PrimaryTable() string
	Process(con *gorm.DB, base Base) error
	Version() uint
}

// Reprocess reprocesses the given type that are out of date w.r.t. the provided processor.
func Reprocess(ctx context.Context, con *gorm.DB, processor Processor) error {
	tableName := processor.PrimaryTable()
	log := logger.FromContext(ctx).With("table", tableName)

	numRows := 0
	tx := con.Table(tableName).
		Where("processor_version < ?", processor.Version()).
		Preload("Data")
	for {
		// TODO batch this with gorm's FindInBatches()
		/*
			rows, err := tx.Rows()
			if err != nil {
				return err
			}
			defer rows.Close()

			base := Base{}
			for rows.Next() {
				if err := tx.ScanRows(rows, &base); err != nil {
					return err
				}
		*/

		base := Base{}
		if err := tx.Take(&base).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return err
		}
		log.DebugContext(ctx, "Reprocessing",
			"id", base.ID,
			"canonicalDigest", base.Data.CanonicalDigest,
			"oldVersion", base.ProcessorVersion, "newVersion", processor.Version())
		base.ProcessorVersion = processor.Version()
		if err := processor.Process(con, base); err != nil {
			return err
		}

		numRows++
	}
	log.InfoContext(ctx, "Reprocessed", "rows", numRows)
	return nil
}

/*
// Not used yet

// ReprocessBatch reprocesses the given type that are out of date w.r.t. the provided processor
func ReprocessBatch(con *gorm.DB, processor Processor) error {
	batchSize := 10
	tableName := processor.PrimaryTable()

	tx := con.Table(tableName).
		Where("processor_version < ?", processor.Version()).
		Preload("Data")
	results := []Base{}
	result := tx.FindInBatches(&results, batchSize, func(tx *gorm.DB, batch int) error {
		updated := make([]any, tx.RowsAffected)
		for i, base := range results {
			log.V(1).Info("Reprocessing",
				"id", base.ID,
				"canonical digest", base.Data.CanonicalDigest,
				"old version", base.ProcessorVersion, "new version", processor.Version())
			base.ProcessorVersion = processor.Version()
			new, err := processor.Process(con, base)
			if err != nil {
				return err
			}
			updated[i] = new
			if err := con.Session(&gorm.Session{FullSaveAssociations: true}).Save(new).Error; err != nil {
				return err
			}
		}
		// return con.Table(tableName).Session(&gorm.Session{FullSaveAssociations: true}).Updates(updated).Error
		return nil
	})
	if result.Error != nil {
		return result.Error
	}
	log.Info("Reprocessed", "rows", result.RowsAffected)
	return nil
}
*/

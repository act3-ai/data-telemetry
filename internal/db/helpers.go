package db

import (
	"fmt"

	"gorm.io/gorm"
)

type locatable[K comparable] interface {
	GetLocation() K
}

// No longer applicable but might be needed in the future
// for why *T is needed see
// https://stackoverflow.com/questions/69573113/how-can-i-instantiate-a-new-pointer-of-type-argument-with-generic-go

func processByIndex[
	DBItem locatable[uint], Item any](
	con *gorm.DB, model any, fieldName string, items []Item,
	convert func(DBItem, int, Item) (*DBItem, error),
) ([]DBItem, error) {

	// first we grab what we have (so we can use the primary keys and index fields)
	var dbItems []DBItem
	if err := con.Model(model).Association(fieldName).
		Find(&dbItems); err != nil {
		return nil, fmt.Errorf("retrieving items from database: %w", err)
	}

	// get the index -> primary key mapping
	ids := make(map[uint]DBItem, len(dbItems))
	for _, dbItem := range dbItems {
		ids[dbItem.GetLocation()] = dbItem
	}

	// (re-)create the records
	dbItemsNew := make([]DBItem, 0, len(items))
	for i, item := range items {
		ui := uint(i)
		dbItem, err := convert(ids[ui], i, item)
		if err != nil {
			return nil, err
		}
		delete(ids, ui)
		dbItemsNew = append(dbItemsNew, *dbItem)
	}

	// remove any extraneous records
	for _, dbItem := range ids {
		if err := con.Delete(&dbItem).Error; err != nil {
			return nil, err
		}
	}

	return dbItemsNew, nil
}

func processByKey[DBItem locatable[string]](
	con *gorm.DB, dbBottle *Bottle, fieldName string, items map[string]string,
	convert func(DBItem, string, string) (*DBItem, error),
) ([]DBItem, error) {

	// first we grab what we have (so we can use the primary keys and index fields)
	var dbItems []DBItem
	if err := con.Model(dbBottle).Association(fieldName).
		Find(&dbItems); err != nil {
		return nil, fmt.Errorf("retrieving items from database: %w", err)
	}

	// get the index -> primary key mapping
	ids := make(map[string]DBItem, len(dbItems))
	for _, dbItem := range dbItems {
		ids[dbItem.GetLocation()] = dbItem
	}

	// (re-)create the records
	dbItemsNew := make([]DBItem, 0, len(items))
	for k, item := range items {
		dbItem, err := convert(ids[k], k, item)
		if err != nil {
			return nil, err
		}
		delete(ids, k)
		dbItemsNew = append(dbItemsNew, *dbItem)
	}

	// remove any extraneous records
	for _, dbItem := range ids {
		if err := con.Delete(&dbItem).Error; err != nil {
			return nil, err
		}
	}

	return dbItemsNew, nil
}

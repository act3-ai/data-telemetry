package db

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"

	"gitlab.com/act3-ai/asce/data/telemetry/v3/pkg/types"
	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
)

// EventProcessorVersion is the current version of the processor code.  This is incremented after each measurable change to the EventProcessor().
const EventProcessorVersion = 3

// EventProcessor handles bottle processing.
type EventProcessor struct{}

// Version returns the processor version.
func (p *EventProcessor) Version() uint {
	return EventProcessorVersion
}

// PrimaryTable returns primary table that this processor updates.
func (p *EventProcessor) PrimaryTable() string {
	return "events"
}

// Process converts Event data to the DB model.
func (p *EventProcessor) Process(con *gorm.DB, base Base) error {
	// ctx := con.Statement.Context
	// log := logger.FromContext(ctx)

	var eventDto types.Event
	if err := json.Unmarshal(base.Data.RawData, &eventDto); err != nil {
		return httputil.NewHTTPError(err, http.StatusConflict, "Failed to parse event", "request data", string(base.Data.RawData))
	}

	// input validation
	if err := eventDto.Validate(); err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid event definition: "+err.Error(), "event", eventDto)
	}

	if err := eventDto.ManifestDigest.Validate(); err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid manifestDigest")
	}

	// Find the manifest to which this event corresponds
	tx := con.
		Preload("Bottle").
		Preload("Layers").
		Scopes(FilterByDigest(eventDto.ManifestDigest, "manifests"))
	dbManifest := Manifest{}
	if err := tx.First(&dbManifest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.NewMissingDigestsError("manifest", []digest.Digest{eventDto.ManifestDigest})
		}
		return err
	}

	dbEvent := Event{
		Base: base,

		Manifest:       dbManifest,
		ManifestDigest: eventDto.ManifestDigest,

		Bottle:       dbManifest.Bottle,
		BottleDigest: dbManifest.BottleDigest,

		Action:       string(eventDto.Action),
		Repository:   eventDto.Repository,
		Tag:          eventDto.Tag,
		AuthRequired: eventDto.AuthRequired,
		Bandwidth:    eventDto.Bandwidth,
		Timestamp:    eventDto.Timestamp,
		Username:     eventDto.Username,
	}

	return con.Save(&dbEvent).Error
}

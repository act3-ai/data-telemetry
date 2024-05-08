package types

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/opencontainers/go-digest"

	val "gitlab.com/act3-ai/asce/data/schema/pkg/validation"
)

// EventAction is string constant that indicates a type of event, such as push or pull.
type EventAction string

const (
	// EventPush indicates a push action.
	EventPush EventAction = "push"
	// EventPull indicates a pull action.
	EventPull EventAction = "pull"
)

// Event is the /event request body.
type Event struct {
	ManifestDigest digest.Digest `json:"manifestDigest,omitempty"`
	Action         EventAction   `json:"action,omitempty"`
	Repository     string        `json:"repository,omitempty"`
	Tag            string        `json:"tag,omitempty"`
	AuthRequired   bool          `json:"authRequired,omitempty"`

	// Bandwidth in bytes per second
	Bandwidth uint64 `json:"bandwidth,omitempty"`

	Timestamp time.Time `json:"timestamp,omitempty"`
	Username  string    `json:"username,omitempty"`
}

// Validate Events.
func (e Event) Validate() error {
	conservativeNow := time.Now().Add(time.Minute)

	return validation.ValidateStruct(&e, //nolint:wrapcheck
		validation.Field(&e.ManifestDigest, validation.Required, val.IsDigest),
		validation.Field(&e.Action, validation.In(EventPush, EventPull)),
		validation.Field(&e.Repository, validation.Required),
		// validation.Field(&e.Tag, validation.Required),
		validation.Field(&e.AuthRequired),
		// validation.Field(&e.Bandwidth, validation.Min(0.0)), // TODO uncomment this when ace-dt is ready to provide this
		validation.Field(&e.Timestamp, validation.Required, validation.Max(conservativeNow)),
		validation.Field(&e.Username, validation.Required),
	)
}

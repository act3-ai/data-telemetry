package webapp

import (
	"time"

	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/act3-ai/data-telemetry/v3/internal/db"
)

func toAge(t time.Time) string {
	return duration.ShortHumanDuration(time.Since(t))
}

// getCommonLabelsFromBotleEntries will return the slice of labels that are common between bottles given.
func getCommonLabelsFromBotleEntries(bottleResultEntries []bottleResultEntry) []db.Label {
	commonLabels := []db.Label{}
	for i, b := range bottleResultEntries {
		if i == 0 {
			commonLabels = append(commonLabels, b.Labels...)
			continue
		}
		foundCommonLabels := []db.Label{}
		for _, cl := range commonLabels {
			for _, l := range b.Labels {
				if l.Key == cl.Key && l.Value == cl.Value {
					foundCommonLabels = append(foundCommonLabels, cl)
					continue
				}
			}
		}
		commonLabels = foundCommonLabels
	}
	return commonLabels
}

// removeLabels will remove labels in labelsToRemove from labels and return the new set.
func removeLabels(labelsToRemove []db.Label, labels []db.Label) []db.Label {
	result := []db.Label{}
	for _, l := range labels {
		foundLabelToRemove := false
		for _, lr := range labelsToRemove {
			if l.Key == lr.Key && l.Value == lr.Value {
				foundLabelToRemove = true
				continue
			}
		}
		if !foundLabelToRemove {
			result = append(result, l)
		}
	}
	return result
}

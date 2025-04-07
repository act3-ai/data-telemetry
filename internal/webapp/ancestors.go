package webapp

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"strings"

	echarts "github.com/go-echarts/go-echarts/v2/charts"
	echartsOpts "github.com/go-echarts/go-echarts/v2/opts"
	echartsRender "github.com/go-echarts/go-echarts/v2/render"
	echartsTemplates "github.com/go-echarts/go-echarts/v2/templates"
	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	"github.com/act3-ai/data-telemetry/v3/internal/db"
)

// extendedAncestors represents a collection of ancestors that include all types.
type extendedAncestors struct {
	NonBottles     []db.NonBottleRelative
	UnknownBottles []db.UnknownBottleRelative
	Bottles        []db.BottleRelative // This is also known as a db.Generation which has a FindRelativeIdx(source) on it
}

func newExtendedAncestors(ancestors, current db.Generation) extendedAncestors {
	// Get the other parents
	return extendedAncestors{
		NonBottles:     current.GetNonBottleParents(),
		UnknownBottles: current.GetUnknownBottleParents(ancestors),
		Bottles:        ancestors,
	}
}

func (g extendedAncestors) Size() int {
	return len(g.NonBottles) + len(g.UnknownBottles) + len(g.Bottles)
}

func (g extendedAncestors) FindRelative(nonBottleFormat, unknownBottleFormat, knownBottleFormat string, source db.Source) string {
	if len(source.BottleDigest) == 0 {
		// only search non bottles
		for i, r := range g.NonBottles {
			if r.MatchesSource(source) {
				return fmt.Sprintf(nonBottleFormat, i)
			}
		}
	} else {
		for i, r := range g.UnknownBottles {
			if r.MatchesSource(source) {
				return fmt.Sprintf(unknownBottleFormat, i)
			}
		}

		for i, r := range g.Bottles {
			if r.MatchesSource(source) {
				return fmt.Sprintf(knownBottleFormat, i)
			}
		}
	}

	// not found
	return ""
}

var customChartTemplate = `
{{- define "chart" }}
    {{- template "header" . }}
    {{- template "base" . }}
<style>
    .container {margin-top:30px; display: flex;justify-content: center;align-items: center;}
    .item {margin: auto;}
</style>
<script type="text/javascript">
goecharts_{{ .ChartID | safeJS }}.on('click', function(params) {
	if (params.dataType != "node") { return; }
	if (params.data.category == 3) {
		window.location = params.data.name;
	} else if (params.data.category == 2) {
		window.location = window.location.origin + "/www/bottle.html?digest=" + encodeURIComponent(params.data.name);
	} else { return; }
});
</script>
{{ end }}
`

var customHeaderTpl = `
{{ define "header" }}
<head>
    <meta charset="utf-8">
{{- range .JSAssets.Values }}
    <script src="{{ . }}"></script>
{{- end }}
{{- range .CustomizedJSAssets.Values }}
    <script src="{{ . }}"></script>
{{- end }}
{{- range .CSSAssets.Values }}
    <link href="{{ . }}" rel="stylesheet">
{{- end }}
{{- range .CustomizedCSSAssets.Values }}
    <link href="{{ . }}" rel="stylesheet">
{{- end }}
</head>
{{ end }}
`

// SVGs that are used for graph nodes must be in path format
// https://www.w3.org/TR/SVG/paths.html#PathData
// https://echarts.apache.org/en/option.html#series-graph.symbol
var (
	encodedBottleSVG        = "path://m 22,0 c -1.652344,0 -3,1.347656 -3,3 v 2 c 0,1.160156 0.839844,2 2,2 0,0.640625 -0.359375,1.101563 -1,1.21875 -4.703125,1.03125 -8,5.09375 -8,9.875 V 20.125 c 0,1.304688 0.835938,2.429688 2,2.84375 V 37.0625 c -1.164062,0.414063 -2,1.539063 -2,2.84375 V 45 c 0,2.757813 2.242188,5 5,5 h 16 c 2.757813,0 5,-2.242187 5,-5 V 39.90625 C 38,38.601563 37.164063,37.476563 36,37.0625 V 37 H 16 V 35 H 36 V 25 H 16 v -2 h 20 v -0.03125 c 1.164063,-0.414062 2,-1.539062 2,-2.84375 V 18.09375 C 38,13.3125 34.707031,9.257813 29.96875,8.21875 29.363281,8.109375 29,7.640625 29,7 30.160156,7 31,6.160156 31,5 V 3 C 31,1.347656 29.652344,0 28,0 Z"
	encodedBottleCrossedSVG = "path://m 22,0 c -1.652342,0 -3,1.3476577 -3,3 v 2 c 0,1.1601548 0.839845,2 2,2 0,0.6406244 -0.359376,1.1015631 -1,1.21875 -1.953626,0.4283697 -3.658753,1.3853971 -4.991577,2.694092 L 5.0847168,1.2712402 1.9067383,4.6610107 12.514282,14.918335 C 12.183921,15.920069 12,16.98612 12,18.09375 V 20.125 c 0,1.304687 0.835939,2.429688 2,2.84375 V 37.0625 c -1.164061,0.414063 -2,1.539064 -2,2.84375 V 45 c 0,2.75781 2.242191,5 5,5 h 16 c 2.75781,0 5,-2.24219 5,-5 v -5.09375 c 0,-0.126609 -0.01438,-0.249697 -0.02966,-0.372437 l 8.851684,8.559449 2.966065,-3.389893 L 36,31.307373 V 25 H 29.508057 L 27.449463,23 H 36 v -0.03125 c 1.164062,-0.414062 2,-1.539063 2,-2.84375 V 18.09375 C 38,13.312505 34.707026,9.257812 29.96875,8.21875 29.363282,8.1093751 29,7.6406244 29,7 30.160155,7 31,6.1601548 31,5 V 3 C 31,1.3476577 29.652342,0 28,0 Z m -6,23 h 4.871948 l 2.068238,2 H 16 Z m 0,12 h 17.281738 l 2.068238,2 H 16 Z"
	encodedGlobeSVG         = "path://M12 22C6.47715 22 2 17.5228 2 12C2 6.47715 6.47715 2 12 2C17.5228 2 22 6.47715 22 12C22 17.5228 17.5228 22 12 22ZM9.71002 19.6674C8.74743 17.6259 8.15732 15.3742 8.02731 13H4.06189C4.458 16.1765 6.71639 18.7747 9.71002 19.6674ZM10.0307 13C10.1811 15.4388 10.8778 17.7297 12 19.752C13.1222 17.7297 13.8189 15.4388 13.9693 13H10.0307ZM19.9381 13H15.9727C15.8427 15.3742 15.2526 17.6259 14.29 19.6674C17.2836 18.7747 19.542 16.1765 19.9381 13ZM4.06189 11H8.02731C8.15732 8.62577 8.74743 6.37407 9.71002 4.33256C6.71639 5.22533 4.458 7.8235 4.06189 11ZM10.0307 11H13.9693C13.8189 8.56122 13.1222 6.27025 12 4.24799C10.8778 6.27025 10.1811 8.56122 10.0307 11ZM14.29 4.33256C15.2526 6.37407 15.8427 8.62577 15.9727 11H19.9381C19.542 7.8235 17.2836 5.22533 14.29 4.33256Z"
)

type graphCategoryID int

const (
	graphCategoryMainBottle graphCategoryID = iota
	graphCategoryExternalBottle
	graphCategoryBottleRelative
	graphCategoryExternalURL
)

var graphCategories = map[graphCategoryID]*echartsOpts.GraphCategory{
	graphCategoryMainBottle: {
		Name: "Main Bottle",
		Label: &echartsOpts.Label{
			Show:      false,
			Color:     "green",
			Formatter: " ",
		},
	},
	graphCategoryBottleRelative: {
		Name: "Bottle Relative",
		Label: &echartsOpts.Label{
			Show:      false,
			Color:     "blue",
			Formatter: " ",
		},
	},
	graphCategoryExternalBottle: {
		Name: "External Bottle Relative",
		Label: &echartsOpts.Label{
			Show:      false,
			Color:     "yellow",
			Formatter: " ",
		},
	},
	graphCategoryExternalURL: {
		Name: "External URL",
		Label: &echartsOpts.Label{
			Show:      false,
			Color:     "white",
			Formatter: " ",
		},
	},
}

var graphSymbols = map[graphCategoryID]string{
	graphCategoryMainBottle:     encodedBottleSVG,
	graphCategoryBottleRelative: encodedBottleSVG,
	graphCategoryExternalBottle: encodedBottleCrossedSVG,
	graphCategoryExternalURL:    encodedGlobeSVG,
}

// graphNodeData contains all of the data needed to display a graph node.
type graphNodeData struct {
	displayData
	// Other graph nodes that this links to
	linksTo map[string]graphNodeData
	// icon size multiplier
	sizeMultiplier float32
	// encoded svg string for icon
	symbolSVG string
	// category for coloring, filtering and symbol
	category graphCategoryID
	// vertical offset factor relative to center
	verticalOffsetFactor float32
	// horizontal offset factor relative to center
	horizontalOffsetFactor float32
}

type displayData struct {
	// main identifier for graph node
	Name         string
	Description  string
	Authors      []db.Author
	PullScore    int
	Metrics      []db.Metric
	Labels       []db.Label
	IsDeprecated bool
	Note         string
}

func getGraphNodeDataByName(gnd *[]graphNodeData, names []string) *graphNodeData {
	for _, g := range *gnd {
		for _, n := range names {
			if g.Name == n {
				return &g
			}
		}
	}
	return nil
}

// GetAncestryGraphHTML renders an HTML graph of ancestry and returns it as a string.
// numGenAncestors and numGenDescendents determine the number of generations in each direction to render.
// templates is a pointer to the templates that include the needed html templates to be used in the graph rendering.
func GetAncestryGraphHTML(ctx context.Context, con *gorm.DB, bottle *db.BottleRelative, numGenAncestors, numGenDescendents uint, templates *template.Template) (template.HTML, error) {
	log := logger.FromContext(ctx).WithGroup("lineage-graph-gen").With("bottle", string(bottle.Digests[0]))

	log.DebugContext(ctx, "generating lineage graph HTML")

	// Get descendents and ancestors
	ancestors, err := db.GetAncestors(con, bottle.Digests[0], numGenAncestors)
	if err != nil {
		return "", err
	}

	descendants, err := db.GetDescendants(con, bottle.Digests[0], numGenDescendents)
	if err != nil {
		return "", err
	}

	// Convert ancestors and descendents into easier to graph datasctructure
	allGraphNodeData, err := getGraphData(bottle, ancestors, descendants)
	if err != nil {
		return "", err
	}

	// Generate echarts graph nodes and links
	graphNodes, graphLinks, err := getGraphNodesAndLinks(ctx, allGraphNodeData, templates, log)
	if err != nil {
		return "", err
	}

	// initialize graph
	graph := echarts.NewGraph()
	// To get a minimal js file use the online custom package builder
	// https://echarts.apache.org/en/builder.html
	graph.JSAssets.Values = []string{"/www/static/js/echarts.min.js"}
	graph.SetGlobalOptions(
		echarts.WithLegendOpts(echartsOpts.Legend{
			Show:      true,
			Type:      "plain",
			Top:       "top",
			TextStyle: &echartsOpts.TextStyle{Color: "white"},
		}),
		echarts.WithColorsOpts(echartsOpts.Colors{}),
		echarts.WithInitializationOpts(echartsOpts.Initialization{
			Width:  "1000px",
			Height: "700px",
		}))

	categories := make([]*echartsOpts.GraphCategory, len(graphCategories))
	for categoryIndex, c := range graphCategories {
		categories[categoryIndex] = c
	}

	// add nodes and links to graph
	graph.AddSeries("lineage", graphNodes, graphLinks, echarts.WithGraphChartOpts(
		echartsOpts.GraphChart{
			Layout:           "none",
			Roam:             false,
			EdgeSymbol:       []string{"none", "arrow"},
			EdgeSymbolSize:   20,
			Categories:       categories,
			EdgeLabel:        &echartsOpts.EdgeLabel{Show: true, Color: "white"},
			SymbolKeepAspect: true,
		},
	))
	log.DebugContext(ctx, "graph config generated", "json", graph.JSON())

	contents := []string{echartsTemplates.BaseTpl, customChartTemplate, customHeaderTpl}
	tpl := echartsRender.MustTemplate("chart", contents)

	graphHTML := new(bytes.Buffer)
	if err := tpl.ExecuteTemplate(graphHTML, "chart", graph); err != nil {
		return "", fmt.Errorf("could not execute chart generation template: %w", err)
	}
	return template.HTML(graphHTML.String()), nil
}

func appendDescendentGraphData(graphData *[]graphNodeData, descendents []db.Generation) {
	// keep track of what to link to
	childrenBottleGraphData := make(map[string]graphNodeData, 0)

	// create descendents graph data
	for genIndex := len(descendents) - 1; genIndex >= 0; genIndex-- {
		for relIndex := 0; relIndex < len(descendents[genIndex]); relIndex++ {

			// for each relative, find all of the bottles in the next generation that has this as its source
			digestStrings := digestsToStrings(descendents[genIndex][relIndex].Digests)
			if genIndex != len(descendents)-1 {
				childrenBottleGraphData = getChildrenGraphDataFromGeneration(digestStrings, descendents[genIndex+1], *graphData)
				existingGraphNode := getGraphNodeDataByName(graphData, digestStrings)
				if existingGraphNode != nil {
					existingGraphNode.linksTo = childrenBottleGraphData
					continue
				}
			}

			currentDescendent := descendents[genIndex][relIndex]
			*graphData = append(*graphData, graphNodeData{
				displayData: displayData{
					Name:        currentDescendent.Digests[0].String(),
					Description: currentDescendent.Description,
					Authors:     currentDescendent.Authors,
					// TODO
					// pullScore:   currentDescendent.PullScore,
					PullScore: -1,
					Metrics:   currentDescendent.Metrics,
					Labels:    currentDescendent.Labels,
					// TODO
					IsDeprecated: false,
				},
				linksTo:                childrenBottleGraphData,
				sizeMultiplier:         1,
				symbolSVG:              graphSymbols[graphCategoryBottleRelative],
				category:               graphCategoryBottleRelative,
				verticalOffsetFactor:   float32(relIndex),
				horizontalOffsetFactor: float32(genIndex + 1),
			})
		}
	}
}

func getLargestGeneration(generations []db.Generation) float64 {
	largestGenNum := 0.0
	for _, g := range generations {
		largestGenNum = math.Max(float64(largestGenNum), float64(len(g)))
	}
	return largestGenNum
}

func appendRootNodeGraphData(graphData *[]graphNodeData, rootBottle *db.BottleRelative, children map[string]graphNodeData) {
	*graphData = append(*graphData, graphNodeData{
		displayData: displayData{
			Name:        string(rootBottle.Digests[0]),
			Description: rootBottle.Description,
			Authors:     rootBottle.Authors,
			// TODO
			// pullScore:   rootBottle.PullScore,
			PullScore: -1,
			Metrics:   rootBottle.Metrics,
			Labels:    rootBottle.Labels,
			// TODO
			IsDeprecated: false,
			Note:         "This is the bottle you are currently viewing",
		},
		linksTo:                children,
		sizeMultiplier:         2,
		symbolSVG:              graphSymbols[graphCategoryMainBottle],
		category:               graphCategoryMainBottle,
		verticalOffsetFactor:   0,
		horizontalOffsetFactor: 0,
	})
}

func appendAncestorGraphData(graphData *[]graphNodeData, ancestors []db.Generation, startingGeneration db.Generation) {
	priorGen := startingGeneration
	// keep track of how much to shift each generation vertically
	totalVerticalShiftAmount := 0
	for genIndex, gen := range ancestors {
		// use extended ancestor for more detailed graph data
		currentGenExtendedAncestors := newExtendedAncestors(gen, priorGen)

		thisGenGraphData := make([]graphNodeData, 0)

		currentGenRelativeCount := 0

		for _, bottleRelative := range currentGenExtendedAncestors.Bottles {
			// for each relative, find all of the bottles in the next generation that has this as its source
			// note every ancestor should point to something
			digestStrings := digestsToStrings(bottleRelative.Digests)
			childrenBottleGraphData := getChildrenGraphDataFromGeneration(digestStrings, priorGen, *graphData)
			existingGraphNode := getGraphNodeDataByName(graphData, digestStrings)
			if existingGraphNode != nil {
				existingGraphNode.linksTo = childrenBottleGraphData
				continue
			}
			thisGenGraphData = append(thisGenGraphData, graphNodeData{
				displayData: displayData{
					Name:        bottleRelative.Digests[0].String(),
					Description: bottleRelative.Description,
					Authors:     bottleRelative.Authors,
					// TODO
					// pullScore:   bottleRelative.PullScore,
					PullScore: -1,
					Metrics:   bottleRelative.Metrics,
					Labels:    bottleRelative.Labels,
					// TODO
					IsDeprecated: false,
				},
				linksTo:                childrenBottleGraphData,
				sizeMultiplier:         1,
				symbolSVG:              graphSymbols[graphCategoryBottleRelative],
				category:               graphCategoryBottleRelative,
				verticalOffsetFactor:   float32(currentGenRelativeCount),
				horizontalOffsetFactor: float32((genIndex * -1) - 1),
			})
			currentGenRelativeCount++
		}

		for _, unknownBottleRelative := range currentGenExtendedAncestors.UnknownBottles {
			// for each relative, find all of the bottles in the next generation that has this as its source
			// note every ancestor should point to something
			digestStrings := []string{unknownBottleRelative.Digest.String()}
			childrenBottleGraphData := getChildrenGraphDataFromGeneration(digestStrings, priorGen, *graphData)
			existingGraphNode := getGraphNodeDataByName(graphData, digestStrings)
			if existingGraphNode != nil {
				existingGraphNode.linksTo = childrenBottleGraphData
				continue
			}
			thisGenGraphData = append(thisGenGraphData, graphNodeData{
				displayData: displayData{
					Name:         unknownBottleRelative.Digest.String(),
					Description:  "",
					Authors:      []db.Author{},
					PullScore:    -1,
					Metrics:      []db.Metric{},
					Labels:       []db.Label{},
					IsDeprecated: false,
					Note:         "This bottle's info is not stored on this Telemetry server",
				},
				linksTo:                childrenBottleGraphData,
				sizeMultiplier:         1,
				symbolSVG:              graphSymbols[graphCategoryExternalBottle],
				category:               graphCategoryExternalBottle,
				verticalOffsetFactor:   float32(currentGenRelativeCount),
				horizontalOffsetFactor: float32((genIndex * -1) - 1),
			})
			currentGenRelativeCount++
		}

		for _, nonBottleRelative := range currentGenExtendedAncestors.NonBottles {
			// for each relative, find all of the bottles in the next generation that has this as its source
			// note every ancestor should point to something
			uriStringSlice := []string{nonBottleRelative.URI}
			childrenBottleGraphData := getChildrenGraphDataFromGeneration(uriStringSlice, priorGen, *graphData)
			existingGraphNode := getGraphNodeDataByName(graphData, uriStringSlice)
			if existingGraphNode != nil {
				existingGraphNode.linksTo = childrenBottleGraphData
				continue
			}
			thisGenGraphData = append(thisGenGraphData, graphNodeData{
				displayData: displayData{
					Name:         nonBottleRelative.URI,
					Description:  "",
					Authors:      []db.Author{},
					PullScore:    -1,
					Metrics:      []db.Metric{},
					Labels:       []db.Label{},
					IsDeprecated: false,
					Note:         "This is an external source.",
				},
				linksTo:                childrenBottleGraphData,
				sizeMultiplier:         1,
				symbolSVG:              graphSymbols[graphCategoryExternalURL],
				category:               graphCategoryExternalURL,
				verticalOffsetFactor:   float32(currentGenRelativeCount),
				horizontalOffsetFactor: float32((genIndex * -1) - 1),
			})
			currentGenRelativeCount++
		}

		// at the end of each generation shift the vertical offset
		totalVerticalShiftAmount += (len(thisGenGraphData) - len(priorGen)) / 2
		for tggdIndex := range thisGenGraphData {
			thisGenGraphData[tggdIndex].verticalOffsetFactor -= float32(totalVerticalShiftAmount)
		}

		*graphData = append(*graphData, thisGenGraphData...)
		priorGen = gen
	}
}

func getGraphData(rootBottle *db.BottleRelative, ancestors, descendents []db.Generation) ([]graphNodeData, error) {
	graphData := make([]graphNodeData, 0)
	// populate the graph data right to left (child to parent) to generate links properly

	appendDescendentGraphData(&graphData, descendents)

	rootNodeChildren := make(map[string]graphNodeData, 0)

	if len(descendents) > 0 {
		rootNodeChildren = getChildrenGraphDataFromGeneration(digestsToStrings(rootBottle.Digests), descendents[0], graphData)
	}
	appendRootNodeGraphData(&graphData, rootBottle, rootNodeChildren)

	appendAncestorGraphData(&graphData, ancestors, db.Generation{*rootBottle})

	// get the largest generation so we may move items accordingly
	allGenerations := make([]db.Generation, 0)
	allGenerations = append(allGenerations, ancestors...)
	allGenerations = append(allGenerations, descendents...)
	largestGenNum := getLargestGeneration(allGenerations)

	// hack to get some space between the title and the real graph content
	graphData = append(graphData, graphNodeData{
		displayData: displayData{
			Name: "this is just for space",
		},
		linksTo:                make(map[string]graphNodeData, 0),
		sizeMultiplier:         0,
		symbolSVG:              graphSymbols[graphCategoryExternalURL],
		category:               graphCategoryExternalURL,
		verticalOffsetFactor:   ((float32(largestGenNum) - 1) / 2) * -1.25,
		horizontalOffsetFactor: 0,
	})

	return graphData, nil
}

func getGraphNodesAndLinks(ctx context.Context, bgData []graphNodeData, templates *template.Template, log *slog.Logger) ([]echartsOpts.GraphNode, []echartsOpts.GraphLink, error) {
	graphNodes := make([]echartsOpts.GraphNode, 0)
	graphLinks := make([]echartsOpts.GraphLink, 0)

	var horizontalOffsetUnit float32 = 20
	var verticalOffsetUnit float32 = 25
	var symbolSizeScaleUnit float32 = 50

	tooltipMaxFieldLen := 50

	tooltipFormatter := func(data displayData) string {
		truncatedDisplayData := displayData{
			Name:        truncateStringField(tooltipMaxFieldLen, data.Name),
			Description: truncateStringField(tooltipMaxFieldLen, data.Description),
			Authors: truncateKVField[db.Author](tooltipMaxFieldLen, data.Authors, func(a db.Author) int {
				return len(a.Name)
			}),
			PullScore: data.PullScore,
			Metrics: truncateKVField[db.Metric](tooltipMaxFieldLen, data.Metrics, func(m db.Metric) int {
				return len(m.Name) + len(fmt.Sprintf("%g", m.Value))
			}),
			Labels: truncateKVField[db.Label](tooltipMaxFieldLen, data.Labels, func(l db.Label) int {
				return len(l.Key) + len(l.Value)
			}),
			IsDeprecated: data.IsDeprecated,
			Note:         data.Note,
		}

		templateData := struct {
			Data displayData
			Top  string
		}{
			Data: truncatedDisplayData,
			Top:  "../",
		}
		fmtText := bytes.NewBuffer(make([]byte, 0))
		err := templates.ExecuteTemplate(fmtText, "lineage-tooltip.html", templateData)
		if err != nil {
			log.ErrorContext(ctx, "could not generate lineage graph tooltip", "error", err)
			return fmt.Sprintf("<b>ERROR: could not generate tooltip</b><br /><small>ID: %s</small>", data.Name)
		}

		return fmtText.String()
	}

	for _, bgd := range bgData {
		graphNodes = append(graphNodes, echartsOpts.GraphNode{
			Name:       bgd.Name,
			X:          bgd.horizontalOffsetFactor*horizontalOffsetUnit + 250,
			Y:          bgd.verticalOffsetFactor*verticalOffsetUnit + 250,
			Value:      0,
			Fixed:      false,
			Category:   bgd.category,
			Symbol:     bgd.symbolSVG,
			SymbolSize: []int{int(bgd.sizeMultiplier * symbolSizeScaleUnit), int(bgd.sizeMultiplier * symbolSizeScaleUnit)},
			ItemStyle:  &echartsOpts.ItemStyle{},
			Tooltip: &echartsOpts.Tooltip{
				Show:      true,
				Formatter: tooltipFormatter(bgd.displayData),
			},
		})

		for linkName, bgdLink := range bgd.linksTo {
			graphLinks = append(graphLinks, echartsOpts.GraphLink{
				Source: bgd.Name,
				Target: bgdLink.Name,
				Value:  0,
				Label: &echartsOpts.EdgeLabel{
					Show:      true,
					Position:  "middle",
					Formatter: " " + truncateStringField(15, linkName),
					FontSize:  14,
					Color:     "white",
				},
			})
		}
	}

	return graphNodes, graphLinks, nil
}

func appendGraphDataBySourceID(graphDataMap map[string]graphNodeData, relative db.BottleRelative, relativeSource db.Source, targetID string, graphDataPool []graphNodeData) {
	// is there a match for the relative sources and the target id?
	if relativeSource.BottleDigest.String() == targetID || relativeSource.URI == targetID {
		// now that we found a match, find the corresponding graphData
		for _, gd := range graphDataPool {
			for _, dgst := range relative.Digests {
				if gd.Name == dgst.String() {
					graphDataMap[relativeSource.Name] = gd
					break
				}
			}
		}
	}
}

func getChildrenGraphDataFromGeneration(originIdentifiers []string, nextGeneration db.Generation, graphData []graphNodeData) map[string]graphNodeData {
	childrenBottleGraphData := make(map[string]graphNodeData, 0)
	// foreach relative in the next generation,
	for _, nextGenRel := range nextGeneration {
		// foreach source on the relative
		for _, nextGenRelSource := range nextGenRel.Sources {
			// foreach origin node ID
			for _, originID := range originIdentifiers {
				appendGraphDataBySourceID(childrenBottleGraphData, nextGenRel, nextGenRelSource, originID, graphData)
			}
		}
	}
	return childrenBottleGraphData
}

func digestsToStrings(digests []digest.Digest) []string {
	stringSlice := make([]string, 0)
	for _, d := range digests {
		stringSlice = append(stringSlice, d.String())
	}
	return stringSlice
}

type kvFieldConstraint interface {
	db.Author | db.Metric | db.Label
}

func truncateKVField[T kvFieldConstraint](maxTextLength int, items []T, getItemLen func(T) int) []T {
	truncatedItems := make([]T, 0)
	totalItemTextLen := 0
	for _, i := range items {
		// authorListTextLen += len(a.Name)
		totalItemTextLen += getItemLen(i)
		if totalItemTextLen > maxTextLength {
			break
		}

		truncatedItems = append(truncatedItems, i)
	}
	return truncatedItems
}

func truncateStringField(maxTextLength int, field string) string {
	if len(field) < maxTextLength {
		return field
	}

	truncatedField := field[:maxTextLength]
	if strings.Contains(truncatedField, ",") {
		truncatedField = truncatedField[:strings.LastIndex(truncatedField, ",")]
	}
	return fmt.Sprintf("%s ...", truncatedField)
}

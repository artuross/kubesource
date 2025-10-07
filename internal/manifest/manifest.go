package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"

	yaml "github.com/goccy/go-yaml"

	"github.com/artuross/kubesource/pkg/config"
)

// ParsedDocument represents a rendered Kubernetes manifest (as an ordered map)
// accompanied by extracted metadata in the shape of config.Selector.
// Only fields present in the underlying document are populated.
type ParsedDocument struct {
	Metadata config.Selector
	Document yaml.MapSlice
}

func (d ParsedDocument) Matches(filter *config.Filter) bool {
	if filter == nil {
		return true
	}

	included := slices.ContainsFunc(filter.Include, d.matchesSelector)
	if len(filter.Include) == 0 {
		included = true
	}

	if !included {
		return false
	}

	if len(filter.Exclude) == 0 {
		return true
	}

	return !slices.ContainsFunc(filter.Exclude, d.matchesSelector)
}

func (d ParsedDocument) matchesSelector(selector config.Selector) bool {
	if selector.Kind != "" && selector.Kind != d.Metadata.Kind {
		return false
	}

	if selector.APIVersion != "" && selector.APIVersion != d.Metadata.APIVersion {
		return false
	}

	if selector.Metadata == nil {
		return true
	}

	if selector.Metadata.Name != "" && selector.Metadata.Name != d.Metadata.Metadata.Name {
		return false
	}

	if selector.Metadata.Namespace != "" && selector.Metadata.Namespace != d.Metadata.Metadata.Namespace {
		return false
	}

	for selectorLabel, selectorLabelValue := range selector.Metadata.Labels {
		documentLabel, ok := d.Metadata.Metadata.Labels[selectorLabel]
		if !ok {
			return false
		}

		if documentLabel != selectorLabelValue {
			return false
		}
	}

	return true
}

// ParseDocuments parses a multi-document or single-document YAML payload and
// returns a slice of ParsedDocument. Non-mapping (non-object) documents or empty
// documents are skipped, since Kubernetes manifests are expected to be mappings.
//
// Order of keys in each document is preserved via yaml.MapSlice.
func ParseDocuments(content []byte) ([]ParsedDocument, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(content))

	var documents []ParsedDocument
	for {
		var document yaml.MapSlice
		err := decoder.Decode(&document)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("decoding YAML document: %w", err)
		}

		if len(document) == 0 {
			continue
		}

		documents = append(documents, ParsedDocument{
			Metadata: extractSelector(document),
			Document: document,
		})
	}

	return documents, nil
}

func extractMapSliceString(document yaml.MapSlice, key string) string {
	value, ok := getMapSliceValue(document, key)
	if !ok {
		return ""
	}

	str, ok := value.(string)
	if !ok {
		return ""
	}

	return str
}

func extractMetadata(document yaml.MapSlice) *config.MetadataSelector {
	metadata, ok := getMapSliceValue(document, "metadata")
	if !ok {
		return nil
	}

	name := getNestedString(metadata, "name")
	namespace := getNestedString(metadata, "namespace")
	labels := getNestedLabels(metadata)

	if name == "" && namespace == "" && len(labels) == 0 {
		return nil
	}

	return &config.MetadataSelector{
		Name:      name,
		Namespace: namespace,
		Labels:    labels,
	}
}

func extractSelector(document yaml.MapSlice) config.Selector {
	return config.Selector{
		Kind:       extractMapSliceString(document, "kind"),
		APIVersion: extractMapSliceString(document, "apiVersion"),
		Metadata:   extractMetadata(document),
	}
}

func getMapSliceValue(document yaml.MapSlice, key string) (any, bool) {
	for _, item := range document {
		itemKey, ok := item.Key.(string)
		if !ok {
			continue
		}

		if itemKey == key {
			return item.Value, true
		}
	}

	return nil, false
}

func getNestedLabels(metadata any) map[string]string {
	labels := map[string]string{}

	var labelsNode any
	switch metadata := metadata.(type) {
	case yaml.MapSlice:
		value, ok := getMapSliceValue(metadata, "labels")
		if !ok {
			return map[string]string{}
		}

		labelsNode = value

	case map[string]any:
		labelsNode = metadata["labels"]
	}

	switch labelsNode := labelsNode.(type) {
	case yaml.MapSlice:
		for _, item := range labelsNode {
			itemKey, ok := item.Key.(string)
			if !ok {
				continue
			}

			value, ok := item.Value.(string)
			if !ok {
				continue
			}

			labels[itemKey] = value
		}
	case map[string]any:
		for key, value := range labelsNode {
			value, ok := value.(string)
			if !ok {
				continue
			}

			labels[key] = value
		}
	}

	return labels
}

func getNestedString(document any, key string) string {
	switch document := document.(type) {
	case yaml.MapSlice:
		value, ok := getMapSliceValue(document, key)
		if !ok {
			return ""
		}

		str, ok := value.(string)
		if !ok {
			return ""
		}

		return str

	case map[string]any:
		value, ok := document[key]
		if !ok {
			return ""
		}

		str, ok := value.(string)
		if !ok {
			return ""
		}

		return str
	}

	return ""
}

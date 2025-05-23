package main

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"gopkg.in/yaml.v3"

	"github.com/mirecl/golangci-lint/v2/pkg/config"
	"github.com/mirecl/golangci-lint/v2/scripts/website/types"
)

const listItemPrefix = "list-item-"

const (
	keyLinters    = "linters"
	keyFormatters = "formatters"
	keySettings   = "settings"
)

func getLintersListMarkdown(enabled bool, src string) string {
	linters, err := readJSONFile[[]*types.LinterWrapper](src)
	if err != nil {
		panic(err)
	}

	var neededLcs []*types.LinterWrapper
	for _, lc := range linters {
		if lc.Internal {
			continue
		}

		if slices.Contains(slices.Collect(maps.Keys(lc.Groups)), config.GroupStandard) == enabled {
			neededLcs = append(neededLcs, lc)
		}
	}

	sort.Slice(neededLcs, func(i, j int) bool {
		return neededLcs[i].Name < neededLcs[j].Name
	})

	slices.SortFunc(neededLcs, func(a, b *types.LinterWrapper) int {
		if a.IsDeprecated() && b.IsDeprecated() {
			return strings.Compare(a.Name, b.Name)
		}

		if a.IsDeprecated() {
			return 1
		}

		if b.IsDeprecated() {
			return -1
		}

		return strings.Compare(a.Name, b.Name)
	})

	lines := []string{
		"|Name|Description|AutoFix|Since|",
		"|----|-----------|-------|-----|",
	}

	for _, lc := range neededLcs {
		line := fmt.Sprintf("|%s|%s|%v|%s|",
			getName(lc),
			getDesc(lc),
			check(lc.CanAutoFix, "Auto fix supported"),
			lc.Since,
		)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func getName(lc *types.LinterWrapper) string {
	name := spanWithID(listItemPrefix+lc.Name, "", "")

	if hasSettings(lc.Name) && lc.Deprecation == nil {
		name += fmt.Sprintf("[%[1]s&nbsp;%[2]s](#%[1]s \"%[1]s configuration\")", lc.Name, "<FaCog size={'0.8rem'} />")
	} else {
		name += fmt.Sprintf("%[1]s[%[2]s](#%[2]s \"%[2]s has no configuration\")", spanWithID(lc.Name, "", ""), lc.Name)
	}

	if lc.OriginalURL != "" {
		icon := "<FaGithub size={'0.8rem'} />"
		if strings.Contains(lc.OriginalURL, "gitlab") {
			icon = "<FaGitlab size={'0.8rem'} />"
		}

		name += fmt.Sprintf("&nbsp;[%s](%s)", span(lc.Name+" repository", icon), lc.OriginalURL)
	}

	if lc.Deprecation == nil {
		return name
	}

	title := "deprecated"
	if lc.Deprecation.Replacement != "" {
		title += fmt.Sprintf(" since %s", lc.Deprecation.Since)
	}

	return name + "&nbsp;" + span(title, "⚠")
}

func check(b bool, title string) string {
	if b {
		return span(title, "✔")
	}
	return ""
}

func getDesc(lc *types.LinterWrapper) string {
	desc := lc.Desc
	if lc.Deprecation != nil {
		desc = lc.Deprecation.Message
		if lc.Deprecation.Replacement != "" {
			desc += fmt.Sprintf(" Replaced by %s.", lc.Deprecation.Replacement)
		}
	}

	return formatDesc(desc)
}

func formatDesc(desc string) string {
	runes := []rune(desc)

	r, _ := utf8.DecodeRuneInString(desc)
	runes[0] = unicode.ToUpper(r)

	if runes[len(runes)-1] != '.' {
		runes = append(runes, '.')
	}

	return strings.ReplaceAll(string(runes), "\n", "<br/>")
}

func hasSettings(name string) bool {
	tp := reflect.TypeOf(config.LintersSettings{})

	for i := range tp.NumField() {
		if strings.EqualFold(name, tp.Field(i).Name) {
			return true
		}
	}

	tp = reflect.TypeOf(config.FormatterSettings{})

	for i := range tp.NumField() {
		if strings.EqualFold(name, tp.Field(i).Name) {
			return true
		}
	}

	return false
}

func span(title, icon string) string {
	return fmt.Sprintf(`<span title=%q>%s</span>`, title, icon)
}

func spanWithID(id, title, icon string) string {
	return fmt.Sprintf(`<span id=%q title=%q>%s</span>`, id, title, icon)
}

type SettingSnippets struct {
	ConfigurationFile  string
	LintersSettings    string
	FormattersSettings string
}

func marshallSnippet(node *yaml.Node) (string, error) {
	builder := &strings.Builder{}

	if node.Value != "" {
		_, _ = fmt.Fprintf(builder, "### %s\n\n", node.Value)
	}
	_, _ = fmt.Fprintln(builder, "```yaml")

	encoder := yaml.NewEncoder(builder)
	encoder.SetIndent(2)

	err := encoder.Encode(node)
	if err != nil {
		return "", err
	}

	_, _ = fmt.Fprintln(builder, "```")
	_, _ = fmt.Fprintln(builder)

	return builder.String(), nil
}

type ExampleSnippetsExtractor struct {
	referencePath string
	assetsPath    string
}

func NewExampleSnippetsExtractor() *ExampleSnippetsExtractor {
	return &ExampleSnippetsExtractor{
		referencePath: ".golangci.reference.yml",
		assetsPath:    "assets",
	}
}

func (e *ExampleSnippetsExtractor) GetExampleSnippets() (*SettingSnippets, error) {
	reference, err := os.ReadFile(e.referencePath)
	if err != nil {
		return nil, fmt.Errorf("can't read .golangci.reference.yml: %w", err)
	}

	snippets, err := e.extractExampleSnippets(reference)
	if err != nil {
		return nil, fmt.Errorf("can't extract example snippets from .golangci.reference.yml: %w", err)
	}

	return snippets, nil
}

func (e *ExampleSnippetsExtractor) extractExampleSnippets(example []byte) (*SettingSnippets, error) {
	var data yaml.Node
	if err := yaml.Unmarshal(example, &data); err != nil {
		return nil, err
	}

	root := data.Content[0]

	globalNode := &yaml.Node{
		Kind:        root.Kind,
		Style:       root.Style,
		Tag:         root.Tag,
		Value:       root.Value,
		Anchor:      root.Anchor,
		Alias:       root.Alias,
		HeadComment: root.HeadComment,
		LineComment: root.LineComment,
		FootComment: root.FootComment,
		Line:        root.Line,
		Column:      root.Column,
	}

	snippets := SettingSnippets{}

	builder := strings.Builder{}

	for j, node := range root.Content {
		switch node.Value {
		case "run", "output", keyLinters, keyFormatters, "issues", "severity", "version":
		default:
			continue
		}

		nextNode := root.Content[j+1]

		newNode := &yaml.Node{
			Kind: nextNode.Kind,
			Content: []*yaml.Node{
				{
					HeadComment: fmt.Sprintf("See the dedicated %q documentation section.", node.Value),
					Kind:        node.Kind,
					Style:       node.Style,
					Tag:         node.Tag,
					Value:       "option",
				},
				{
					Kind:  node.Kind,
					Style: node.Style,
					Tag:   node.Tag,
					Value: "value",
				},
			},
		}

		if node.Value == "version" {
			n := &yaml.Node{
				HeadComment: fmt.Sprintf("See the dedicated %q documentation section.", node.Value),
				Kind:        node.Kind,
				Style:       node.Style,
				Tag:         node.Tag,
				Value:       node.Value,
				Content:     node.Content,
			}

			globalNode.Content = append(globalNode.Content, n, nextNode)
		} else {
			globalNode.Content = append(globalNode.Content, node, newNode)
		}

		if node.Value == keyLinters || node.Value == keyFormatters {
			for i := 0; i < len(nextNode.Content); i++ {
				if nextNode.Content[i].Value != keySettings {
					continue
				}

				settingSections, err := e.getSettingSections(node, nextNode.Content[i+1])
				if err != nil {
					return nil, err
				}

				switch node.Value {
				case keyLinters:
					snippets.LintersSettings = settingSections

				case keyFormatters:
					snippets.FormattersSettings = settingSections
				}

				nextNode.Content[i+1].Content = []*yaml.Node{
					{
						HeadComment: fmt.Sprintf(`See the dedicated "%s.%s" documentation section.`, node.Value, nextNode.Content[i].Value),
						Kind:        node.Kind,
						Style:       node.Style,
						Tag:         node.Tag,
						Value:       "option",
					},
					{
						Kind:  node.Kind,
						Style: node.Style,
						Tag:   node.Tag,
						Value: "value",
					},
				}

				i++
			}
		}

		nodeSection := &yaml.Node{
			Kind:    root.Kind,
			Style:   root.Style,
			Tag:     root.Tag,
			Value:   root.Value,
			Content: []*yaml.Node{node, nextNode},
		}

		snippet, errSnip := marshallSnippet(nodeSection)
		if errSnip != nil {
			return nil, errSnip
		}

		_, _ = builder.WriteString(fmt.Sprintf("### `%s` configuration\n\n%s", node.Value, snippet))
	}

	overview, err := marshallSnippet(globalNode)
	if err != nil {
		return nil, err
	}

	snippets.ConfigurationFile = overview + builder.String()

	return &snippets, nil
}

func (e *ExampleSnippetsExtractor) getSettingSections(node, nextNode *yaml.Node) (string, error) {
	linters, err := readJSONFile[[]*types.LinterWrapper](filepath.Join(e.assetsPath, fmt.Sprintf("%s-info.json", node.Value)))
	if err != nil {
		return "", err
	}

	lintersDesc := make(map[string]string)
	for _, lc := range linters {
		if lc.Internal {
			continue
		}

		// it's important to use lc.Name() nor name because name can be alias
		lintersDesc[lc.Name] = getDesc(lc)
	}

	builder := &strings.Builder{}

	for i := 0; i < len(nextNode.Content); i += 2 {
		r := &yaml.Node{
			Kind:  yaml.MappingNode,
			Tag:   nextNode.Tag,
			Value: node.Value,
			Content: []*yaml.Node{
				{
					Kind:  yaml.ScalarNode,
					Value: node.Value,
					Tag:   node.Tag,
				},
				{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "settings",
							Tag:   node.Tag,
						},
						{
							Kind:    yaml.MappingNode,
							Tag:     nextNode.Tag,
							Content: []*yaml.Node{nextNode.Content[i], nextNode.Content[i+1]},
						},
					},
				},
			},
		}

		_, _ = fmt.Fprintf(builder, "### %s\n\n", nextNode.Content[i].Value)
		_, _ = fmt.Fprintf(builder, "%s\n\n", lintersDesc[nextNode.Content[i].Value])
		_, _ = fmt.Fprintln(builder, "```yaml")

		encoder := yaml.NewEncoder(builder)
		encoder.SetIndent(2)

		err := encoder.Encode(r)
		if err != nil {
			return "", err
		}

		_, _ = fmt.Fprintln(builder, "```")
		_, _ = fmt.Fprintln(builder)
		_, _ = fmt.Fprintf(builder, "[%s](#%s)\n\n", span("Back to the top", "<FaArrowUp />"), listItemPrefix+nextNode.Content[i].Value)
		_, _ = fmt.Fprintln(builder)
	}

	return builder.String(), nil
}

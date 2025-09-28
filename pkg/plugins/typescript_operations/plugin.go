package typescript_operations

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jzeiders/graphql-go-gen/pkg/documents"
	"github.com/jzeiders/graphql-go-gen/pkg/plugin"
	"github.com/jzeiders/graphql-go-gen/pkg/plugins/base"
	"github.com/vektah/gqlparser/v2/ast"
)

// Plugin generates TypeScript types for GraphQL operations
type Plugin struct{}

// New creates a new TypeScript operations plugin
func New() plugin.Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "typescript-operations"
}

// Description returns the plugin description
func (p *Plugin) Description() string {
	return "Generates TypeScript types for GraphQL operations (queries, mutations, subscriptions)"
}

// DefaultConfig returns the default configuration
func (p *Plugin) DefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"strictNulls":           false,
		"immutableTypes":        false,
		"noExport":              false,
		"preResolveTypes":       true,
		"skipTypename":          false,
		"dedupeOperationSuffix": false,
		"omitOperationSuffix":   false,
		"flattenGeneratedTypes": false,
		"avoidOptionals":        false,
	}
}

// ValidateConfig validates the plugin configuration
func (p *Plugin) ValidateConfig(config map[string]interface{}) error {
	return nil
}

// Generate generates TypeScript operation types
func (p *Plugin) Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	if req.Schema == nil || req.Schema.Raw() == nil {
		return nil, fmt.Errorf("schema is required")
	}

	astSchema := req.Schema.Raw()
	cfg := parseConfig(req.Config)

	allOps := documents.CollectAllOperations(req.Documents)
	operations := make([]*ast.OperationDefinition, 0, len(allOps))
	for _, op := range allOps {
		if op.Name != "" {
			operations = append(operations, op)
		}
	}

	allFrags := documents.CollectAllFragments(req.Documents)
	fragments := make([]*ast.FragmentDefinition, 0, len(allFrags))
	fragmentMap := make(map[string]*ast.FragmentDefinition, len(allFrags))
	for _, frag := range allFrags {
		fragments = append(fragments, frag)
		fragmentMap[frag.Name] = frag
	}

	if len(operations) == 0 && len(fragments) == 0 {
		return &plugin.GenerateResponse{
			Files: map[string][]byte{
				req.OutputPath: []byte("// No GraphQL operations found\n"),
			},
		}, nil
	}

	gen := newGenerator(astSchema, cfg, fragmentMap)

	var sections []string
	if cfg.FlattenGeneratedTypes {
		sections = append(sections, gen.renderFragments(fragments)...)
		sections = append(sections, gen.renderOperations(operations)...)
	} else {
		sections = append(sections, gen.renderOperations(operations)...)
		sections = append(sections, gen.renderFragments(fragments)...)
	}

	content := strings.Join(filterNonEmpty(sections), "\n\n")

	return &plugin.GenerateResponse{
		Files: map[string][]byte{
			req.OutputPath: []byte(content),
		},
	}, nil
}

func filterNonEmpty(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			out = append(out, part)
		}
	}
	return out
}

type operationsConfig struct {
	ImmutableTypes          bool
	SkipTypename            bool
	OmitOperationSuffix     bool
	FlattenGeneratedTypes   bool
	FlattenIncludeFragments bool
	AvoidOptionals          bool
}

func parseConfig(cfg map[string]interface{}) operationsConfig {
	return operationsConfig{
		ImmutableTypes:          base.GetBool(cfg, "immutableTypes", false),
		SkipTypename:            base.GetBool(cfg, "skipTypename", false),
		OmitOperationSuffix:     base.GetBool(cfg, "omitOperationSuffix", false),
		FlattenGeneratedTypes:   base.GetBool(cfg, "flattenGeneratedTypes", false),
		FlattenIncludeFragments: base.GetBool(cfg, "flattenGeneratedTypesIncludeFragments", false),
		AvoidOptionals:          base.GetBool(cfg, "avoidOptionals", false),
	}
}

type generator struct {
	schema    *ast.Schema
	config    operationsConfig
	fragments map[string]*ast.FragmentDefinition
	scalars   map[string]string
}

func newGenerator(schema *ast.Schema, cfg operationsConfig, fragments map[string]*ast.FragmentDefinition) *generator {
	scalars := map[string]string{
		"ID":      "string",
		"String":  "string",
		"Boolean": "boolean",
		"Int":     "number",
		"Float":   "number",
	}
	return &generator{
		schema:    schema,
		config:    cfg,
		fragments: fragments,
		scalars:   scalars,
	}
}

func (g *generator) renderOperations(ops []*ast.OperationDefinition) []string {
	sections := make([]string, 0, len(ops))
	for _, op := range ops {
		if op.Name == "" {
			continue
		}
		sections = append(sections, g.renderOperation(op))
	}
	return sections
}

func (g *generator) renderOperation(op *ast.OperationDefinition) string {
	baseName := base.ToPascalCase(op.Name)
	suffix := ""
	if !g.config.OmitOperationSuffix {
		switch op.Operation {
		case ast.Query:
			suffix = "Query"
		case ast.Mutation:
			suffix = "Mutation"
		case ast.Subscription:
			suffix = "Subscription"
		}
	}

	variablesName := baseName + suffix + "Variables"
	resultName := baseName + suffix

	variablesBlock := g.renderVariablesType(op)
	resultType := g.renderOperationResult(op)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("export type %s = %s;\n\n\n", variablesName, variablesBlock))
	sb.WriteString(fmt.Sprintf("export type %s = %s;", resultName, resultType.Render("")))
	return sb.String()
}

func (g *generator) renderFragments(frags []*ast.FragmentDefinition) []string {
	if len(frags) == 0 {
		return nil
	}
	fragments := make([]*ast.FragmentDefinition, len(frags))
	copy(fragments, frags)

	if g.config.FlattenGeneratedTypes {
		sort.Slice(fragments, func(i, j int) bool {
			return fragments[i].Name < fragments[j].Name
		})
	}

	sections := make([]string, 0, len(fragments))
	for _, frag := range fragments {
		if frag == nil {
			continue
		}
		typeName := base.ToPascalCase(frag.Name) + "Fragment"
		selection := g.renderSelection(frag.TypeCondition, frag.SelectionSet, !g.config.SkipTypename)
		sections = append(sections, fmt.Sprintf("export type %s = %s;", typeName, selection.Render("")))
	}
	return sections
}

func (g *generator) renderVariablesType(op *ast.OperationDefinition) string {
	if len(op.VariableDefinitions) == 0 {
		return "Exact<{ [key: string]: never; }>"
	}

	lines := make([]string, 0, len(op.VariableDefinitions))
	for _, v := range op.VariableDefinitions {
		name := v.Variable
		if name == "" {
			continue
		}
		typ := g.renderVariableType(v.Type)
		optional := !v.Type.NonNull && !g.config.AvoidOptionals
		suffix := ";"
		if optional {
			lines = append(lines, fmt.Sprintf("  %s?: %s%s", name, typ, suffix))
		} else {
			lines = append(lines, fmt.Sprintf("  %s: %s%s", name, typ, suffix))
		}
	}

	if len(lines) == 0 {
		return "Exact<{ [key: string]: never; }>"
	}

	return "Exact<{\n" + strings.Join(lines, "\n") + "\n}>"
}

func (g *generator) renderVariableType(t *ast.Type) string {
	if t == nil {
		return "any"
	}
	baseType := g.renderInputBaseType(t)
	if !t.NonNull {
		return fmt.Sprintf("InputMaybe<%s>", baseType)
	}
	return baseType
}

func (g *generator) renderInputBaseType(t *ast.Type) string {
	if t == nil {
		return "any"
	}
	if t.Elem != nil {
		inner := g.renderInputBaseType(t.Elem)
		if !t.Elem.NonNull {
			inner = fmt.Sprintf("InputMaybe<%s>", inner)
		}
		listType := "Array"
		if g.config.ImmutableTypes {
			listType = "ReadonlyArray"
		}
		return fmt.Sprintf("%s<%s>", listType, inner)
	}
	name := unwrapTypeName(t)
	if name == "" {
		return "any"
	}
	if def := g.schema.Types[name]; def != nil && def.Kind == ast.Scalar {
		return fmt.Sprintf("Scalars['%s']['input']", name)
	}
	return name
}

func (g *generator) renderOperationResult(op *ast.OperationDefinition) tsType {
	var rootType *ast.Definition
	switch op.Operation {
	case ast.Query:
		rootType = g.schema.Query
	case ast.Mutation:
		rootType = g.schema.Mutation
	case ast.Subscription:
		rootType = g.schema.Subscription
	}
	if rootType == nil {
		return &tsPrimitive{Code: "{}"}
	}
	return g.renderSelection(rootType.Name, op.SelectionSet, true)
}

func (g *generator) renderSelection(typeName string, selectionSet ast.SelectionSet, allowTypename bool) tsType {
	def := g.schema.Types[typeName]
	if def == nil {
		return &tsPrimitive{Code: "{}"}
	}

	if def.Kind == ast.Union {
		return g.renderUnionSelection(def, selectionSet)
	}

	collector := newFieldCollector(g.config.ImmutableTypes)
	g.applySelections(def, selectionSet, collector, make(map[string]bool))
	fields := collector.Finalize(g, def, allowTypename && !g.config.SkipTypename, def.Name, false)
	return &tsObject{Fields: fields}
}

func (g *generator) renderUnionSelection(def *ast.Definition, selectionSet ast.SelectionSet) tsType {
	options := make([]tsType, 0, len(def.Types))
	for _, typeName := range def.Types {
		typeDef := g.schema.Types[typeName]
		if typeDef == nil {
			continue
		}
		collector := newFieldCollector(g.config.ImmutableTypes)
		collector.AddTypenameLiteral(typeName, true)
		g.applyUnionSelections(typeDef, selectionSet, collector, make(map[string]bool), typeName)
		fields := collector.Finalize(g, typeDef, false, typeName, true)
		options = append(options, &tsObject{Fields: fields})
	}
	return &tsUnion{Options: options}
}

func (g *generator) applySelections(typeDef *ast.Definition, selectionSet ast.SelectionSet, collector *fieldCollector, visited map[string]bool) {
	for _, sel := range selectionSet {
		switch s := sel.(type) {
		case *ast.Field:
			responseName := s.Alias
			if responseName == "" {
				responseName = s.Name
			}
			if s.Name == "__typename" {
				collector.AddField(responseName, s.Name, nil, &ast.Type{NamedType: "String"}, nil)
				continue
			}
			fieldDef := findFieldDefinition(typeDef, s.Name)
			if fieldDef == nil {
				continue
			}
			collector.AddField(responseName, s.Name, fieldDef, fieldDef.Type, s.SelectionSet)
		case *ast.InlineFragment:
			typeCondition := s.TypeCondition
			if typeCondition == "" || typeCondition == typeDef.Name || typeImplements(typeDef, typeCondition) {
				g.applySelections(typeDef, s.SelectionSet, collector, visited)
			}
		case *ast.FragmentSpread:
			frag := g.fragments[s.Name]
			if frag == nil {
				continue
			}
			if visited[frag.Name] {
				continue
			}
			if frag.TypeCondition == typeDef.Name || typeImplements(typeDef, frag.TypeCondition) || frag.TypeCondition == "" {
				visited[frag.Name] = true
				g.applySelections(typeDef, frag.SelectionSet, collector, visited)
				delete(visited, frag.Name)
			}
		}
	}
}

func (g *generator) applyUnionSelections(typeDef *ast.Definition, selectionSet ast.SelectionSet, collector *fieldCollector, visited map[string]bool, typeName string) {
	for _, sel := range selectionSet {
		switch s := sel.(type) {
		case *ast.Field:
			if s.Name == "__typename" {
				continue
			}
			fieldDef := findFieldDefinition(typeDef, s.Name)
			if fieldDef == nil {
				continue
			}
			responseName := s.Alias
			if responseName == "" {
				responseName = s.Name
			}
			collector.AddField(responseName, s.Name, fieldDef, fieldDef.Type, s.SelectionSet)
		case *ast.InlineFragment:
			if s.TypeCondition == "" || s.TypeCondition == typeName || typeImplements(typeDef, s.TypeCondition) {
				g.applySelections(typeDef, s.SelectionSet, collector, visited)
			}
		case *ast.FragmentSpread:
			frag := g.fragments[s.Name]
			if frag == nil {
				continue
			}
			if visited[frag.Name] {
				continue
			}
			if frag.TypeCondition == typeName || typeImplements(typeDef, frag.TypeCondition) || frag.TypeCondition == "" {
				visited[frag.Name] = true
				g.applySelections(typeDef, frag.SelectionSet, collector, visited)
				delete(visited, frag.Name)
			}
		}
	}
}

func (g *generator) renderTypeForField(fieldType *ast.Type, selectionSets []ast.SelectionSet) tsType {
	if fieldType == nil {
		return &tsPrimitive{Code: "any"}
	}
	if fieldType.Elem != nil {
		elem := g.renderTypeForField(fieldType.Elem, selectionSets)
		if !fieldType.Elem.NonNull {
			elem = &tsNullable{Inner: elem}
		}
		return &tsArray{Elem: elem, Immutable: g.config.ImmutableTypes}
	}

	name := unwrapTypeName(fieldType)
	if name == "" {
		return &tsPrimitive{Code: "any"}
	}

	def := g.schema.Types[name]
	if def == nil {
		if g.isScalar(name) {
			return &tsPrimitive{Code: g.scalarOutput(name)}
		}
		return &tsPrimitive{Code: name}
	}

	switch def.Kind {
	case ast.Scalar:
		return &tsPrimitive{Code: g.scalarOutput(name)}
	case ast.Enum:
		return &tsPrimitive{Code: def.Name}
	case ast.Union:
		combined := combineSelectionSets(selectionSets)
		return g.renderUnionSelection(def, combined)
	case ast.Object, ast.Interface:
		combined := combineSelectionSets(selectionSets)
		if len(combined) == 0 {
			return &tsObject{Fields: []*tsField{}}
		}
		return g.renderSelection(def.Name, combined, true)
	default:
		return &tsPrimitive{Code: def.Name}
	}
}

func (g *generator) scalarOutput(name string) string {
	if v, ok := g.scalars[name]; ok {
		return v
	}
	return "any"
}

func (g *generator) isScalar(name string) bool {
	if _, ok := g.scalars[name]; ok {
		return true
	}
	if def := g.schema.Types[name]; def != nil && def.Kind == ast.Scalar {
		return true
	}
	return false
}

func (g *generator) isScalarOutputType(t *ast.Type) bool {
	if t == nil {
		return true
	}
	if t.Elem != nil {
		return g.isScalarOutputType(t.Elem)
	}
	name := unwrapTypeName(t)
	if name == "" {
		return true
	}
	if g.isScalar(name) {
		return true
	}
	if def := g.schema.Types[name]; def != nil {
		return def.Kind == ast.Scalar || def.Kind == ast.Enum
	}
	return false
}

func combineSelectionSets(sets []ast.SelectionSet) ast.SelectionSet {
	if len(sets) == 0 {
		return nil
	}
	var combined ast.SelectionSet
	for _, sel := range sets {
		combined = append(combined, sel...)
	}
	return combined
}

func unwrapTypeName(t *ast.Type) string {
	if t == nil {
		return ""
	}
	for t.Elem != nil {
		t = t.Elem
	}
	return t.NamedType
}

func findFieldDefinition(def *ast.Definition, name string) *ast.FieldDefinition {
	if def == nil {
		return nil
	}
	for _, field := range def.Fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

func typeImplements(def *ast.Definition, interfaceName string) bool {
	if def == nil {
		return false
	}
	for _, iface := range def.Interfaces {
		if iface == interfaceName {
			return true
		}
	}
	return false
}

type fieldCollector struct {
	immutable   bool
	order       []string
	fields      map[string]*collectedField
	hasTypename bool
}

type collectedField struct {
	ResponseName    string
	GraphQLName     string
	Definition      *ast.FieldDefinition
	Type            *ast.Type
	SelectionSets   []ast.SelectionSet
	IsTypename      bool
	TypenameLiteral string
	ForceRequired   bool
}

func newFieldCollector(immutable bool) *fieldCollector {
	return &fieldCollector{
		immutable: immutable,
		fields:    make(map[string]*collectedField),
	}
}

func (c *fieldCollector) AddField(responseName, graphQLName string, def *ast.FieldDefinition, typ *ast.Type, selection ast.SelectionSet) {
	if existing, ok := c.fields[responseName]; ok {
		if selection != nil && len(selection) > 0 {
			existing.SelectionSets = append(existing.SelectionSets, selection)
		}
		return
	}

	field := &collectedField{
		ResponseName: responseName,
		GraphQLName:  graphQLName,
		Definition:   def,
		Type:         typ,
	}
	if selection != nil && len(selection) > 0 {
		field.SelectionSets = append(field.SelectionSets, selection)
	}

	if graphQLName == "__typename" {
		field.IsTypename = true
		c.hasTypename = true
	}

	c.fields[responseName] = field
	c.order = append(c.order, responseName)
}

func (c *fieldCollector) AddTypenameLiteral(typeName string, required bool) {
	if c.hasTypename {
		return
	}
	field := &collectedField{
		ResponseName:    "__typename",
		GraphQLName:     "__typename",
		IsTypename:      true,
		TypenameLiteral: typeName,
		ForceRequired:   required,
	}
	if required {
		field.Type = &ast.Type{NamedType: "String", NonNull: true}
	} else {
		field.Type = &ast.Type{NamedType: "String"}
	}
	c.fields["__typename"] = field
	c.order = append([]string{"__typename"}, c.order...)
	c.hasTypename = true
}

func (c *fieldCollector) Finalize(g *generator, parentDef *ast.Definition, addTypename bool, typeName string, forceTypenameRequired bool) []*tsField {
	if addTypename && !c.hasTypename && typeName != "" {
		field := &collectedField{
			ResponseName:    "__typename",
			GraphQLName:     "__typename",
			IsTypename:      true,
			TypenameLiteral: typeName,
			Type:            &ast.Type{NamedType: "String"},
		}
		if forceTypenameRequired {
			field.Type.NonNull = true
			field.ForceRequired = true
		}
		c.fields["__typename"] = field
		c.order = append([]string{"__typename"}, c.order...)
		c.hasTypename = true
	}

	scalarFields := make([]*tsField, 0, len(c.order))
	objectFields := make([]*tsField, 0, len(c.order))
	for _, name := range c.order {
		cf := c.fields[name]
		if cf == nil {
			continue
		}
		field := g.buildTsField(cf)
		if cf.IsTypename || g.isScalarOutputType(cf.Type) {
			scalarFields = append(scalarFields, field)
		} else {
			objectFields = append(objectFields, field)
		}
	}
	return append(scalarFields, objectFields...)
}

func (g *generator) buildTsField(cf *collectedField) *tsField {
	readonly := g.config.ImmutableTypes

	if cf.IsTypename {
		literal := cf.TypenameLiteral
		optional := !cf.ForceRequired
		var typeExpr tsType = &tsPrimitive{Code: "string"}
		if literal != "" {
			typeExpr = &tsPrimitive{Code: fmt.Sprintf("'%s'", literal)}
		}
		return &tsField{
			Name:     cf.ResponseName,
			Optional: optional,
			Nullable: false,
			Readonly: readonly,
			Type:     typeExpr,
		}
	}

	typ := cf.Type
	selectionSets := cf.SelectionSets
	var tsType tsType
	if typ == nil {
		tsType = &tsPrimitive{Code: "any"}
	} else {
		tsType = g.renderTypeForField(typ, selectionSets)
	}

	optional := typ != nil && !typ.NonNull && !g.config.AvoidOptionals
	nullable := typ != nil && !typ.NonNull

	return &tsField{
		Name:     cf.ResponseName,
		Optional: optional,
		Nullable: nullable,
		Readonly: readonly,
		Type:     tsType,
	}
}

type tsType interface {
	Render(indent string) string
}

type tsPrimitive struct {
	Code string
}

func (p *tsPrimitive) Render(_ string) string {
	return p.Code
}

type tsNullable struct {
	Inner tsType
}

func (n *tsNullable) Render(indent string) string {
	return n.Inner.Render(indent) + " | null"
}

type tsArray struct {
	Elem      tsType
	Immutable bool
}

func (a *tsArray) Render(indent string) string {
	listType := "Array"
	if a.Immutable {
		listType = "ReadonlyArray"
	}
	if union, ok := a.Elem.(*tsUnion); ok {
		optionIndent := indent + "    "
		var sb strings.Builder
		sb.WriteString(listType + "<\n")
		sb.WriteString(union.renderOptions(optionIndent))
		sb.WriteString("\n" + indent + "  >")
		return sb.String()
	}
	return fmt.Sprintf("%s<%s>", listType, a.Elem.Render(indent))
}

type tsUnion struct {
	Options []tsType
}

func (u *tsUnion) Render(indent string) string {
	parts := make([]string, len(u.Options))
	for i, opt := range u.Options {
		parts[i] = opt.Render(indent)
	}
	return strings.Join(parts, " | ")
}

func (u *tsUnion) renderOptions(indent string) string {
	var sb strings.Builder
	for i, opt := range u.Options {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(indent + "| " + opt.Render(indent+"  "))
	}
	return sb.String()
}

type tsObject struct {
	Fields []*tsField
}

func (o *tsObject) Render(indent string) string {
	if len(o.Fields) == 0 {
		return "{}"
	}
	parts := make([]string, len(o.Fields))
	for i, field := range o.Fields {
		parts[i] = field.Render(indent)
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

type tsField struct {
	Name     string
	Optional bool
	Nullable bool
	Readonly bool
	Type     tsType
}

func (f *tsField) Render(indent string) string {
	var sb strings.Builder
	if f.Readonly {
		sb.WriteString("readonly ")
	}
	sb.WriteString(f.Name)
	if f.Optional {
		sb.WriteString("?")
	}
	sb.WriteString(": ")
	sb.WriteString(f.Type.Render(indent))
	if f.Nullable {
		sb.WriteString(" | null")
	}
	return sb.String()
}

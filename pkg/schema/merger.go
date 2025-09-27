package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// ConflictResolver is a function that resolves conflicts between two GraphQL types
type ConflictResolver func(left *ast.Definition, right *ast.Definition, conflictType string) (*ast.Definition, error)

// MergeOptions contains options for merging schemas
type MergeOptions struct {
	// OnTypeConflict is called when two types with the same name are found
	// If nil, conflicts will result in an error
	OnTypeConflict ConflictResolver

	// TrackSources tracks which source each type came from
	TrackSources bool

	// AllowEmptySchema allows merging to continue even if some schemas are empty
	AllowEmptySchema bool
}

// SchemaConflict represents a conflict between two schema definitions
type SchemaConflict struct {
	TypeName     string
	LeftSource   string
	RightSource  string
	ConflictType string // "type", "field", "argument", "directive"
	Details      string
}

func (c SchemaConflict) Error() string {
	return fmt.Sprintf("schema conflict on type %q between %s and %s: %s conflict - %s",
		c.TypeName, c.LeftSource, c.RightSource, c.ConflictType, c.Details)
}

// SchemaMerger handles merging multiple GraphQL schemas
type SchemaMerger struct {
	options   MergeOptions
	sources   map[string]string // tracks type name to source
	conflicts []SchemaConflict
}

// NewSchemaMerger creates a new schema merger
func NewSchemaMerger(options MergeOptions) *SchemaMerger {
	return &SchemaMerger{
		options:   options,
		sources:   make(map[string]string),
		conflicts: []SchemaConflict{},
	}
}

// MergeSchemas merges multiple schemas into a single schema
func MergeSchemas(ctx context.Context, schemas []*ast.Schema, sources []string, options MergeOptions) (*ast.Schema, error) {
	if len(schemas) == 0 {
		return nil, fmt.Errorf("no schemas provided")
	}

	if len(schemas) != len(sources) {
		return nil, fmt.Errorf("number of schemas (%d) must match number of sources (%d)", len(schemas), len(sources))
	}

	merger := NewSchemaMerger(options)
	return merger.Merge(ctx, schemas, sources)
}

// Merge performs the actual schema merging
func (m *SchemaMerger) Merge(ctx context.Context, schemas []*ast.Schema, sources []string) (*ast.Schema, error) {
	if len(schemas) == 0 {
		return nil, fmt.Errorf("no schemas to merge")
	}

	// Start with an empty merged schema
	merged := &ast.Schema{
		Types:      make(map[string]*ast.Definition),
		Directives: make(map[string]*ast.DirectiveDefinition),
	}

	// Process each schema
	for i, schema := range schemas {
		source := sources[i]

		if schema == nil {
			if !m.options.AllowEmptySchema {
				return nil, fmt.Errorf("schema from source %q is nil", source)
			}
			continue
		}

		// Merge types
		if err := m.mergeTypes(merged, schema, source); err != nil {
			return nil, fmt.Errorf("merging types from %s: %w", source, err)
		}

		// Merge directives
		if err := m.mergeDirectives(merged, schema, source); err != nil {
			return nil, fmt.Errorf("merging directives from %s: %w", source, err)
		}

		// Merge schema definition (Query, Mutation, Subscription)
		if err := m.mergeSchemaDefinition(merged, schema, source); err != nil {
			return nil, fmt.Errorf("merging schema definition from %s: %w", source, err)
		}
	}

	// Validate the merged schema has required types
	if err := m.validateMergedSchema(merged); err != nil {
		return nil, err
	}

	return merged, nil
}

// mergeTypes merges type definitions from source into target
func (m *SchemaMerger) mergeTypes(target, source *ast.Schema, sourceName string) error {
	for typeName, sourceType := range source.Types {
		// Skip built-in types
		if strings.HasPrefix(typeName, "__") {
			continue
		}

		// Skip Query, Mutation, Subscription as they're handled separately
		if typeName == "Query" || typeName == "Mutation" || typeName == "Subscription" {
			continue
		}

		existingType, exists := target.Types[typeName]
		if !exists {
			// No conflict, add the type
			target.Types[typeName] = sourceType
			if m.options.TrackSources {
				m.sources[typeName] = sourceName
			}
			continue
		}

		// Type already exists - check for conflicts
		conflict, err := m.detectTypeConflict(existingType, sourceType)
		if err != nil {
			return fmt.Errorf("error detecting conflict for type %s: %w", typeName, err)
		}

		if conflict != nil {
			// Handle the conflict
			existingSource := m.sources[typeName]
			if existingSource == "" {
				existingSource = "unknown"
			}

			if m.options.OnTypeConflict != nil {
				// Use custom resolver
				resolved, err := m.options.OnTypeConflict(existingType, sourceType, conflict.ConflictType)
				if err != nil {
					return fmt.Errorf("conflict resolution failed for type %s: %w", typeName, err)
				}
				target.Types[typeName] = resolved
				if m.options.TrackSources {
					m.sources[typeName] = fmt.Sprintf("%s+%s", existingSource, sourceName)
				}
			} else {
				// Default behavior: error on conflict
				conflict.TypeName = typeName
				conflict.LeftSource = existingSource
				conflict.RightSource = sourceName
				return conflict
			}
		}
	}

	return nil
}

// detectTypeConflict checks if two types have conflicts
func (m *SchemaMerger) detectTypeConflict(left, right *ast.Definition) (*SchemaConflict, error) {
	if left.Name != right.Name {
		return nil, fmt.Errorf("comparing different type names: %s vs %s", left.Name, right.Name)
	}

	// Check if kinds are different
	if left.Kind != right.Kind {
		return &SchemaConflict{
			TypeName:     left.Name,
			ConflictType: "type",
			Details:      fmt.Sprintf("different kinds: %s vs %s", left.Kind, right.Kind),
		}, nil
	}

	// For scalar types, they're compatible if they have the same name and kind
	if left.Kind == ast.Scalar {
		return nil, nil
	}

	// For enum types, check if values match
	if left.Kind == ast.Enum {
		return m.detectEnumConflict(left, right)
	}

	// For object/interface/input types, check fields
	if left.Kind == ast.Object || left.Kind == ast.Interface || left.Kind == ast.InputObject {
		return m.detectFieldConflicts(left, right)
	}

	// For union types, check member types
	if left.Kind == ast.Union {
		return m.detectUnionConflict(left, right)
	}

	return nil, nil
}

// detectFieldConflicts checks for conflicts in fields
func (m *SchemaMerger) detectFieldConflicts(left, right *ast.Definition) (*SchemaConflict, error) {
	// Build maps of fields for comparison
	leftFields := make(map[string]*ast.FieldDefinition)
	rightFields := make(map[string]*ast.FieldDefinition)

	for _, field := range left.Fields {
		leftFields[field.Name] = field
	}
	for _, field := range right.Fields {
		rightFields[field.Name] = field
	}

	// Check for fields that exist in both but might have conflicts
	for name, leftField := range leftFields {
		if rightField, exists := rightFields[name]; exists {
			// Check if field types match
			if !typesEqual(leftField.Type, rightField.Type) {
				return &SchemaConflict{
					TypeName:     left.Name,
					ConflictType: "field",
					Details:      fmt.Sprintf("field %q has different types: %s vs %s", leftField.Name, leftField.Type.String(), rightField.Type.String()),
				}, nil
			}

			// Check arguments
			if conflict := m.detectArgumentConflicts(left.Name, leftField, rightField); conflict != nil {
				return conflict, nil
			}
		}
	}

	// Check if the field sets are different (different fields indicate incompatible types)
	leftOnlyFields := []string{}
	rightOnlyFields := []string{}

	for name := range leftFields {
		if _, exists := rightFields[name]; !exists {
			leftOnlyFields = append(leftOnlyFields, name)
		}
	}

	for name := range rightFields {
		if _, exists := leftFields[name]; !exists {
			rightOnlyFields = append(rightOnlyFields, name)
		}
	}

	// If there are fields that exist only in one type, this is a conflict
	if len(leftOnlyFields) > 0 && len(rightOnlyFields) > 0 {
		return &SchemaConflict{
			TypeName:     left.Name,
			ConflictType: "field",
			Details:      fmt.Sprintf("types have different fields - left only: %v, right only: %v", leftOnlyFields, rightOnlyFields),
		}, nil
	}

	return nil, nil
}

// detectArgumentConflicts checks for conflicts in field arguments
func (m *SchemaMerger) detectArgumentConflicts(typeName string, leftField, rightField *ast.FieldDefinition) *SchemaConflict {
	for _, leftArg := range leftField.Arguments {
		rightArg := findArgument(rightField.Arguments, leftArg.Name)
		if rightArg == nil {
			continue
		}

		if !typesEqual(leftArg.Type, rightArg.Type) {
			return &SchemaConflict{
				TypeName:     typeName,
				ConflictType: "argument",
				Details:      fmt.Sprintf("field %q argument %q has different types: %s vs %s", leftField.Name, leftArg.Name, leftArg.Type.String(), rightArg.Type.String()),
			}
		}
	}
	return nil
}

// detectEnumConflict checks for conflicts in enum values
func (m *SchemaMerger) detectEnumConflict(left, right *ast.Definition) (*SchemaConflict, error) {
	leftValues := make(map[string]bool)
	for _, val := range left.EnumValues {
		leftValues[val.Name] = true
	}

	rightValues := make(map[string]bool)
	for _, val := range right.EnumValues {
		rightValues[val.Name] = true
	}

	// Check if enum values are exactly the same
	if len(leftValues) != len(rightValues) {
		return &SchemaConflict{
			TypeName:     left.Name,
			ConflictType: "enum",
			Details:      fmt.Sprintf("different number of enum values: %d vs %d", len(leftValues), len(rightValues)),
		}, nil
	}

	for val := range leftValues {
		if !rightValues[val] {
			return &SchemaConflict{
				TypeName:     left.Name,
				ConflictType: "enum",
				Details:      fmt.Sprintf("enum value %q exists in one schema but not the other", val),
			}, nil
		}
	}

	return nil, nil
}

// detectUnionConflict checks for conflicts in union member types
func (m *SchemaMerger) detectUnionConflict(left, right *ast.Definition) (*SchemaConflict, error) {
	leftTypes := make(map[string]bool)
	for _, typ := range left.Types {
		leftTypes[typ] = true
	}

	rightTypes := make(map[string]bool)
	for _, typ := range right.Types {
		rightTypes[typ] = true
	}

	// Check if types are exactly the same
	if len(leftTypes) != len(rightTypes) {
		return &SchemaConflict{
			TypeName:     left.Name,
			ConflictType: "union",
			Details:      fmt.Sprintf("different number of union types: %d vs %d", len(leftTypes), len(rightTypes)),
		}, nil
	}

	for typ := range leftTypes {
		if !rightTypes[typ] {
			return &SchemaConflict{
				TypeName:     left.Name,
				ConflictType: "union",
				Details:      fmt.Sprintf("union member type %q exists in one schema but not the other", typ),
			}, nil
		}
	}

	return nil, nil
}

// mergeDirectives merges directive definitions
func (m *SchemaMerger) mergeDirectives(target, source *ast.Schema, sourceName string) error {
	for name, sourceDir := range source.Directives {
		if existingDir, exists := target.Directives[name]; exists {
			// Check for conflicts in directive definitions
			if !directivesEqual(existingDir, sourceDir) {
				if m.options.OnTypeConflict != nil {
					// For now, just keep the first directive
					// TODO: Add specific directive conflict resolution
					continue
				} else {
					existingSource := "unknown"
					if m.options.TrackSources && m.sources[name] != "" {
						existingSource = m.sources[name]
					}
					return &SchemaConflict{
						TypeName:     name,
						LeftSource:   existingSource,
						RightSource:  sourceName,
						ConflictType: "directive",
						Details:      fmt.Sprintf("directive %q has conflicting definitions", name),
					}
				}
			}
		} else {
			target.Directives[name] = sourceDir
			if m.options.TrackSources {
				m.sources[name] = sourceName
			}
		}
	}
	return nil
}

// mergeSchemaDefinition merges Query, Mutation, and Subscription types
func (m *SchemaMerger) mergeSchemaDefinition(target, source *ast.Schema, sourceName string) error {
	// Merge Query type
	if source.Query != nil || source.Types["Query"] != nil {
		sourceQuery := source.Query
		if sourceQuery == nil {
			sourceQuery = source.Types["Query"]
		}

		if target.Query == nil && target.Types["Query"] == nil {
			target.Query = sourceQuery
			target.Types["Query"] = sourceQuery
		} else {
			targetQuery := target.Query
			if targetQuery == nil {
				targetQuery = target.Types["Query"]
			}

			// Check for conflicts before merging
			if err := m.checkObjectFieldConflicts(targetQuery, sourceQuery, "Query"); err != nil {
				return err
			}

			// Merge fields from source Query into target Query
			merged := m.mergeObjectFields(targetQuery, sourceQuery)
			target.Query = merged
			target.Types["Query"] = merged
		}
	}

	// Merge Mutation type
	if source.Mutation != nil || source.Types["Mutation"] != nil {
		sourceMutation := source.Mutation
		if sourceMutation == nil {
			sourceMutation = source.Types["Mutation"]
		}

		if target.Mutation == nil && target.Types["Mutation"] == nil {
			target.Mutation = sourceMutation
			target.Types["Mutation"] = sourceMutation
		} else {
			targetMutation := target.Mutation
			if targetMutation == nil {
				targetMutation = target.Types["Mutation"]
			}

			// Check for conflicts before merging
			if err := m.checkObjectFieldConflicts(targetMutation, sourceMutation, "Mutation"); err != nil {
				return err
			}

			// Merge fields from source Mutation into target Mutation
			merged := m.mergeObjectFields(targetMutation, sourceMutation)
			target.Mutation = merged
			target.Types["Mutation"] = merged
		}
	}

	// Merge Subscription type
	if source.Subscription != nil || source.Types["Subscription"] != nil {
		sourceSubscription := source.Subscription
		if sourceSubscription == nil {
			sourceSubscription = source.Types["Subscription"]
		}

		if target.Subscription == nil && target.Types["Subscription"] == nil {
			target.Subscription = sourceSubscription
			target.Types["Subscription"] = sourceSubscription
		} else {
			targetSubscription := target.Subscription
			if targetSubscription == nil {
				targetSubscription = target.Types["Subscription"]
			}

			// Check for conflicts before merging
			if err := m.checkObjectFieldConflicts(targetSubscription, sourceSubscription, "Subscription"); err != nil {
				return err
			}

			// Merge fields from source Subscription into target Subscription
			merged := m.mergeObjectFields(targetSubscription, sourceSubscription)
			target.Subscription = merged
			target.Types["Subscription"] = merged
		}
	}

	return nil
}

// checkObjectFieldConflicts checks for field conflicts when merging object types
func (m *SchemaMerger) checkObjectFieldConflicts(target, source *ast.Definition, typeName string) error {
	for _, sourceField := range source.Fields {
		targetField := findField(target.Fields, sourceField.Name)
		if targetField != nil {
			// Check if field types match
			if !typesEqual(targetField.Type, sourceField.Type) {
				if m.options.OnTypeConflict == nil {
					return &SchemaConflict{
						TypeName:     typeName,
						ConflictType: "field",
						Details:      fmt.Sprintf("field %q has different types: %s vs %s", sourceField.Name, targetField.Type.String(), sourceField.Type.String()),
					}
				}
			}

			// Check for argument conflicts
			if !argumentsEqual(targetField.Arguments, sourceField.Arguments) {
				if m.options.OnTypeConflict == nil {
					return &SchemaConflict{
						TypeName:     typeName,
						ConflictType: "argument",
						Details:      fmt.Sprintf("field %q has different arguments", sourceField.Name),
					}
				}
			}
		}
	}
	return nil
}

// mergeObjectFields merges fields from source into target, creating a new Definition
func (m *SchemaMerger) mergeObjectFields(target, source *ast.Definition) *ast.Definition {
	// Create a new definition with merged fields
	merged := &ast.Definition{
		Kind:        target.Kind,
		Name:        target.Name,
		Description: target.Description,
		Fields:      make(ast.FieldList, 0),
		Interfaces:  target.Interfaces,
		Directives:  target.Directives,
		Position:    target.Position,
		BuiltIn:     target.BuiltIn,
	}

	// Add all fields from target
	fieldMap := make(map[string]*ast.FieldDefinition)
	for _, field := range target.Fields {
		fieldMap[field.Name] = field
		merged.Fields = append(merged.Fields, field)
	}

	// Add new fields from source, checking for conflicts
	for _, sourceField := range source.Fields {
		if targetField, exists := fieldMap[sourceField.Name]; exists {
			// Field exists in both - check for conflicts
			if !typesEqual(targetField.Type, sourceField.Type) {
				// Type conflict - skip this field (it's already in merged from target)
				continue
			}

			// Check for argument conflicts
			if !argumentsEqual(targetField.Arguments, sourceField.Arguments) {
				// Argument conflict - skip this field
				continue
			}

			// Fields are compatible, already added from target
		} else {
			// Field only exists in source, add it
			merged.Fields = append(merged.Fields, sourceField)
		}
	}

	return merged
}

// argumentsEqual checks if two argument lists are equal
func argumentsEqual(left, right ast.ArgumentDefinitionList) bool {
	if len(left) != len(right) {
		return false
	}

	// Build map of left arguments
	leftArgs := make(map[string]*ast.ArgumentDefinition)
	for _, arg := range left {
		leftArgs[arg.Name] = arg
	}

	// Check all right arguments exist and match
	for _, rightArg := range right {
		leftArg, exists := leftArgs[rightArg.Name]
		if !exists {
			return false
		}
		if !typesEqual(leftArg.Type, rightArg.Type) {
			return false
		}
	}

	return true
}

// validateMergedSchema ensures the merged schema is valid
func (m *SchemaMerger) validateMergedSchema(schema *ast.Schema) error {
	// Ensure we have at least a Query type
	if schema.Query == nil {
		// Try to find a Query type in the types map
		if queryType, exists := schema.Types["Query"]; exists && queryType.Kind == ast.Object {
			schema.Query = queryType
		} else {
			return fmt.Errorf("merged schema has no Query type")
		}
	}

	// Check for Mutation type if it exists in types but not in schema definition
	if schema.Mutation == nil {
		if mutationType, exists := schema.Types["Mutation"]; exists && mutationType.Kind == ast.Object {
			schema.Mutation = mutationType
		}
	}

	// Check for Subscription type if it exists in types but not in schema definition
	if schema.Subscription == nil {
		if subType, exists := schema.Types["Subscription"]; exists && subType.Kind == ast.Object {
			schema.Subscription = subType
		}
	}

	return nil
}

// Helper functions

func findField(fields ast.FieldList, name string) *ast.FieldDefinition {
	for _, field := range fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

func findArgument(args ast.ArgumentDefinitionList, name string) *ast.ArgumentDefinition {
	for _, arg := range args {
		if arg.Name == name {
			return arg
		}
	}
	return nil
}

func typesEqual(a, b *ast.Type) bool {
	if a == nil || b == nil {
		return a == b
	}

	if a.NamedType != b.NamedType {
		return false
	}

	if a.NonNull != b.NonNull {
		return false
	}

	if (a.Elem == nil) != (b.Elem == nil) {
		return false
	}

	if a.Elem != nil {
		return typesEqual(a.Elem, b.Elem)
	}

	return true
}

func directivesEqual(a, b *ast.DirectiveDefinition) bool {
	if a.Name != b.Name {
		return false
	}

	// Check arguments match
	if !argumentDefinitionsEqual(a.Arguments, b.Arguments) {
		return false
	}

	// Check locations
	if len(a.Locations) != len(b.Locations) {
		return false
	}

	locMap := make(map[ast.DirectiveLocation]bool)
	for _, loc := range a.Locations {
		locMap[loc] = true
	}

	for _, loc := range b.Locations {
		if !locMap[loc] {
			return false
		}
	}

	return true
}

// argumentDefinitionsEqual checks if two argument definition lists are equal
func argumentDefinitionsEqual(left, right ast.ArgumentDefinitionList) bool {
	if len(left) != len(right) {
		return false
	}

	// Build map of left arguments
	leftArgs := make(map[string]*ast.ArgumentDefinition)
	for _, arg := range left {
		leftArgs[arg.Name] = arg
	}

	// Check all right arguments exist and match
	for _, rightArg := range right {
		leftArg, exists := leftArgs[rightArg.Name]
		if !exists {
			return false
		}
		if !typesEqual(leftArg.Type, rightArg.Type) {
			return false
		}
		// Check if default values match if both have defaults
		if (leftArg.DefaultValue != nil) != (rightArg.DefaultValue != nil) {
			return false
		}
	}

	return true
}
package definition

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/weaveworks/schemer/importer"
)

const (
	// DefPrefix is the JSON Schema prefix required in the definition map
	DefPrefix = "#/definitions/"
)

type RefNameFormatFunc func(pkg, name string) string

// Generator can create definitions from Exprs
type Generator struct {
	Strict      bool
	Definitions map[string]*Definition
	Importer    importer.Importer

	// TagNamespace to select which tag to use (e.g. json, yaml)
	TagNamespace string

	// FromatRefName allows customization of the ref name
	FromatRefName RefNameFormatFunc
}

// newStructDefinition handles saving definitions for refs in the map
func (dg *Generator) newStructDefinition(
	thisPkg, name string,
	typeSpec ast.Expr,
	structComment string,
) *Definition {
	def := Definition{}
	commentMeta, err := dg.handleComment(thisPkg, name, structComment, &def)
	if err != nil {
		panic(err)
	}
	if commentMeta.NoDerive {
		return &def
	}
	structType, ok := typeSpec.(*ast.StructType)
	if !ok {
		panic(fmt.Errorf("Cannot handle non-struct TypeSpec %s", name))
	}
	for _, field := range structType.Fields.List {
		tags := strings.Split(GetFieldTag(field).Get(dg.TagNamespace), ",")
		fieldName := tags[0]
		inline := len(field.Names) == 0 && fieldName == ""
		if !inline {
			for _, v := range tags {
				if v == "inline" {
					inline = true
					break
				}
			}
		}

		fieldDoc := field.Doc.Text()

		if def.Properties == nil {
			def.Properties = make(map[string]*Definition)
		}

		var required []string
		var preferredOrder []string
		var properties map[string]*Definition
		// If we are embedded and don't specify a JSON field name
		if inline {
			// We have to handle an embedded field, get its definition
			// and deconstruct it into this def
			ref, _ := dg.newPropertyRef(thisPkg, "", field.Type, fieldDoc, true)
			properties = ref.Properties
			preferredOrder = ref.PreferredOrder
			required = ref.Required
		} else {
			if fieldName == "" || fieldName == "-" {
				// private field
				continue
			}

			// For embedded types
			refName := ""
			// For non-embedded types
			if len(field.Names) > 0 {
				refName = field.Names[0].Name
			}

			field, isRequired := dg.newPropertyRef(thisPkg, refName, field.Type, fieldDoc, false)
			preferredOrder = []string{fieldName}
			properties = map[string]*Definition{
				fieldName: field,
			}

			required = []string{}
			if isRequired {
				required = []string{fieldName}
			}

			// Setting additional properties prevents oneOf from working
			if len(def.OneOf) == 0 {
				def.AdditionalProperties = false
			}
		}

		def.PreferredOrder = append(def.PreferredOrder, preferredOrder...)
		for k, v := range properties {
			def.Properties[k] = v
		}
		def.Required = append(def.Required, required...)
	}
	return &def
}

// newPropertyRef creates a new JSON schema Definition
func (dg *Generator) newPropertyRef(thisPkg, referenceName string, t ast.Expr, propertyComment string, inline bool) (*Definition, bool) {
	var def *Definition

	var refTypeName string
	var refTypeSpec *ast.TypeSpec

typeSwitch:
	switch tt := t.(type) {
	case *ast.Ident:
		typeName := tt.Name
		if obj, ok := dg.Importer.SearchEntryPackageForObj(typeName); ok {
			// If we have a declared type behind our ident, add it
			refTypeName, refTypeSpec = tt.Name, obj.Decl.(*ast.TypeSpec)
		} else if obj, ok = dg.Importer.SearchPackageForObj(thisPkg, typeName); ok {
			typeName = dg.FromatRefName(thisPkg, tt.Name)
			refTypeName, refTypeSpec = typeName, obj.Decl.(*ast.TypeSpec)
		} else {
			// error or primitive types
		}

		def = &Definition{}
		setTypeOrRef(def, typeName)
		setDefaultForNonPointerType(def, typeName)

	case *ast.StarExpr:
		def, _ = dg.newPropertyRef(thisPkg, referenceName, tt.X, propertyComment, inline)
		def.Default = nil

	case *ast.SelectorExpr:
		var (
			typeName string
			err      error
		)
		thisPkg, typeName, refTypeSpec, err = dg.Importer.FindImportedTypeSpec(tt)
		if err != nil {
			pkg := tt.X.(*ast.Ident).Name
			typeName := tt.Sel.Name

			switch {
			case pkg == "yaml" && typeName == "Node":
				def = &Definition{}
				break typeSwitch
			}

			panic(fmt.Errorf(
				"Couldn't import type %s.%s from identifier: %w",
				pkg, typeName, err,
			))
		}

		refTypeName = dg.FromatRefName(thisPkg, typeName)
		def = &Definition{}
		setTypeOrRef(def, refTypeName)

	case *ast.ArrayType:
		item, _ := dg.newPropertyRef(thisPkg, "", tt.Elt, "", inline)
		def = &Definition{
			Type:  "array",
			Items: item,
		}

	case *ast.MapType:
		additional, _ := dg.newPropertyRef(thisPkg, "", tt.Value, "", inline)
		def = &Definition{
			Type:                 "object",
			Default:              "{}",
			AdditionalProperties: additional,
		}

	case *ast.StructType:
		def = dg.newStructDefinition(
			thisPkg, referenceName, t, propertyComment,
		)

	case *ast.InterfaceType:
		// Only `interface{}` is supported
		def = &Definition{}

	default:
		panic(fmt.Errorf("Unexpected type %v for %s", t, referenceName))
	}

	// Add a new definition if necessary
	if refTypeSpec != nil {
		structDef, _ := dg.newPropertyRef(thisPkg, refTypeName, refTypeSpec.Type, refTypeSpec.Doc.Text(), inline)
		// If we're inlining this, we want the struct definition, not the ref
		// and we also don't need to save it in our definitions
		if inline {
			return structDef, false
		}
		dg.Definitions[refTypeName] = structDef
	}

	commentMeta, err := dg.handleComment(thisPkg, referenceName, propertyComment, def)
	if err != nil {
		panic(err)
	}

	return def, commentMeta.Required
}

// CollectDefinitionsFromStruct gets a complete definition for the root object
func (dg *Generator) CollectDefinitionsFromStruct(thisPkg, rootStruct string) {
	rootIdent := ast.Ident{
		NamePos: token.NoPos,
		Name:    rootStruct,
	}
	_, _ = dg.newPropertyRef(thisPkg, rootStruct, &rootIdent, "", false)
}

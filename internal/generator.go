package internal

import (
	"fmt"
	"github.com/vektah/gqlparser/v2/ast"
	"strings"
)

const (
	Header = `// Code generated by go generate; DO NOT EDIT.
// This file was generated from GraphQL schema

package %s
`

	StructTPL = `type %s struct {
%s
}`

	FieldTPL        = "  %s %s `json:\"%s\"`"
	ListFieldTPL    = "  %s []%s `json:\"%s\"`"
	EnumTypeDefTPL  = "type %s %s"
	EnumDefConstTPL = "const %s%s %s = \"%s\""
)

var GQLTypesToGoTypes = map[string]string{
	"Int":     "int64",
	"Float":   "float64",
	"String":  "string",
	"Boolean": "bool",
	"ID":      "string",
}

type GoGenerator struct {
	entities    []string
	enumMapType map[string]string

	packageName   string
	output        Outputer
	disableHeader bool
}

func NewGoGenerator(output Outputer, entities []string, packageName string) *GoGenerator {
	return &GoGenerator{output: output, entities: entities, packageName: packageName}
}

func (g *GoGenerator) Generate(doc *ast.SchemaDocument) error {
	if !g.disableHeader {
		if err := g.output.Writeln(fmt.Sprintf(Header, g.packageName)); err != nil {
			return err
		}
	}

	declaredKeywords := keywordMap{}

	//remap enums
	enumMap := buildEnumMap(doc)
	reqEntities := resolveEntityDependencies(doc, g.entities, enumMap)

	// Write enum const
	for _, e := range enumMap {
		if len(reqEntities) > 0 && !inArray(e.TypeName, reqEntities) {
			continue
		}
		if err := declaredKeywords.Set(e.TypeName); err != nil {
			return err
		}
		if err := g.output.Write(fmt.Sprintf(EnumTypeDefTPL, e.TypeName, "string")); err != nil {
			return err
		}
		if err := g.output.Write("\n"); err != nil {
			return err
		}
		for _, v := range e.Values {
			if err := declaredKeywords.Set(fmt.Sprintf("%s%s", e.TypeName, v)); err != nil {
				return err
			}
			if err := g.output.Write(fmt.Sprintf(EnumDefConstTPL, e.TypeName, v, e.TypeName, v)); err != nil {
				return err
			}
			if err := g.output.Write("\n"); err != nil {
				return err
			}
		}
	}

	for _, i := range doc.Definitions {
		if len(reqEntities) > 0 && !inArray(i.Name, reqEntities) {
			continue
		}
		if i.Name == "Query" || i.Name == "Mutation" {
			continue
		}
		if i.Kind == ast.Object || i.Kind == ast.InputObject {
			var fields []string
			for _, f := range i.Fields {
				typeName := resolveType(f.Type.Name(), enumMap, f.Type.NonNull)
				fieldName := strings.Title(f.Name)
				jsonFieldName := f.Name
				if f.Type.Elem != nil { // list type
					elemTypeName := resolveType(f.Type.Elem.Name(), enumMap, f.Type.Elem.NonNull)
					fields = append(fields, fmt.Sprintf(ListFieldTPL, fieldName, elemTypeName, jsonFieldName))
				} else {
					fields = append(fields, fmt.Sprintf(FieldTPL, fieldName, typeName, jsonFieldName))
				}
			}
			if err := declaredKeywords.Set(i.Name); err != nil {
				return err
			}
			if err := g.output.Writeln(fmt.Sprintf(StructTPL, i.Name, strings.Join(fields, "\n"))); err != nil {
				return err
			}
		} else if i.Kind == ast.Union {
			fields := []string{fmt.Sprintf(FieldTPL, "TypeName", "string", "__typeName")}
			fields = append(fields, i.Types...)
			if err := declaredKeywords.Set(i.Name); err != nil {
				return err
			}
			if err := g.output.Writeln(fmt.Sprintf(StructTPL, i.Name, strings.Join(fields, "\n"))); err != nil {
				return err
			}
		}
		if err := g.output.Writeln(""); err != nil {
			return err
		}
	}

	if missingEntities := declaredKeywords.GetMissingKeys(g.entities); len(missingEntities) > 0 {
		return fmt.Errorf("the following entites are not found in graphql schemas: %v", missingEntities)
	}

	return nil
}

func resolveType(typeName string, enumMap map[string]enum, notNull bool) string {
	if tName, hasType := GQLTypesToGoTypes[typeName]; hasType {
		typeName = tName
	}
	if tName, hasEnumType := enumMap[typeName]; hasEnumType {
		typeName = tName.TypeName
	}
	if !notNull { // if type can be nullable, use pointer
		typeName = strings.Join([]string{"*", typeName}, "")
	}
	return typeName
}


func resolveEntityDependencies(doc *ast.SchemaDocument, reqEntities []string, enumMap map[string]enum) []string {
	dependsOn := map[string][]string{}
	for _, i := range doc.Definitions {
		var depEntities []string
		if i.Kind == ast.Object || i.Kind == ast.InputObject {
			for _, f := range i.Fields {
				if en, hasEnum := enumMap[f.Type.Name()]; hasEnum {
					depEntities = append(depEntities, en.TypeName)
				} else {
					depEntities = append(depEntities, f.Type.Name())
				}
			}
		} else if i.Kind == ast.Union {
			depEntities = i.Types
		}
		dependsOn[i.Name] = append(dependsOn[i.Name], depEntities...)
	}

	entities := reqEntities
	resolvedEntities := map[string]bool{}
	allResolved := false
	for !allResolved {
		allResolved = true
		for _, e := range entities {
			if depEnt, has := dependsOn[e]; has && !resolvedEntities[e] {
				entities = append(entities, depEnt...)
				allResolved = false
				resolvedEntities[e] = true
			}
		}
	}

	return entities
}

type enum struct {
	TypeName string
	Values   []string
}

func buildEnumMap(doc *ast.SchemaDocument) map[string]enum {
	enumMap := map[string]enum{}
	for _, i := range doc.Definitions {
		if i.Kind == ast.Enum {
			enumTypeName := fmt.Sprintf("Enum%s", i.Name)
			var vals []string
			for _, e := range i.EnumValues {
				vals = append(vals, e.Name)
			}
			enumMap[i.Name] = enum{TypeName: enumTypeName, Values: vals}
		}
	}
	return enumMap
}


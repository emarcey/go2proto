package main

import (
	"flag"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/packages"

	gostringconverters "github.com/emarcey/go-string-converters"
)

type arrFlags []string

func (i *arrFlags) String() string {
	return ""
}

func (i *arrFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	filter             = flag.String("filter", "", "Filter struct names.")
	protoFolder        = flag.String("f", "", "Proto output path.")
	currProtoFileName  = flag.String("c", "", "Full filepath for existing version of proto, if applicable.")
	useSnakeFieldNames = flag.Bool("s", false, "Use to set proto structs names to snake_case instead of camelCase.")
	pkgFlags           arrFlags
)

func main() {
	flag.Var(&pkgFlags, "p", "Go source packages.")
	flag.Parse()

	if len(pkgFlags) == 0 || protoFolder == nil {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := checkOutFolder(*protoFolder); err != nil {
		log.Fatal(err)
	}

	currProtoMessages, err := BuildCurrentProtoMap(*currProtoFileName)
	if err != nil {
		log.Fatal(err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	pkgs, err := loadPackages(pwd, pkgFlags)
	if err != nil {
		log.Fatal(err)
	}

	msgs := getMessages(pkgs, *filter, currProtoMessages, *useSnakeFieldNames)

	if err := writeOutput(msgs, *protoFolder); err != nil {
		log.Fatal(err)
	}
}

func checkOutFolder(path string) error {
	_, err := os.Stat(path)
	return err
}

func loadPackages(pwd string, pkgs []string) ([]*packages.Package, error) {
	fset := token.NewFileSet()
	cfg := &packages.Config{
		Dir:  pwd,
		Mode: packages.LoadSyntax,
		Fset: fset,
	}
	return packages.Load(cfg, pkgs...)
}

type message struct {
	Name   string
	Fields []field
}

type field struct {
	Name       string
	TypeName   string
	Order      int
	IsRepeated bool
	Tags       string
	IsEmbedded bool
}

func getMessages(pkgs []*packages.Package, filter string, currProtoMessages ProtoMessageMap, useSnakeFieldNames bool) []message {
	seen := map[string]struct{}{}

	messageMap := make(map[string]message)
	for _, p := range pkgs {
		for _, t := range p.TypesInfo.Defs {
			if t == nil {
				continue
			}
			if !t.Exported() {
				continue
			}
			if _, ok := seen[t.Name()]; ok {
				continue
			}
			if s, ok := t.Type().Underlying().(*types.Struct); ok {
				seen[t.Name()] = struct{}{}
				if filter == "" || strings.Contains(t.Name(), filter) {
					messageMap[t.Name()] = getMessage(t, s, currProtoMessages, useSnakeFieldNames)
				}
			}
		}
	}

	var out []message

	for _, msg := range messageMap {
		msg.Fields = resolveEmbedded(msg.Fields, messageMap, currProtoMessages, msg.Name)
		out = append(out, msg)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func getMessage(t types.Object, s *types.Struct, currProtoMessages ProtoMessageMap, useSnakeFieldNames bool) message {
	msg := message{
		Name:   t.Name(),
		Fields: []field{},
	}

	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		if !f.Exported() || isElasticsearchNoSource(s.Tag(i)) {
			continue
		}
		fieldName := toProtoFieldName(f.Name(), useSnakeFieldNames)
		order := currProtoMessages.GetFieldNum(t.Name(), fieldName)
		newField := field{
			Name:       fieldName,
			TypeName:   toProtoFieldTypeName(f),
			IsRepeated: isRepeated(f),
			Order:      int(order),
			Tags:       s.Tag(i),
			IsEmbedded: f.Embedded(),
		}
		msg.Fields = append(msg.Fields, newField)
	}
	return msg
}

func resolveEmbedded(msgFields []field, messageMap map[string]message, currProtoMessages ProtoMessageMap, msgName string) []field {
	var newFields []field
	for _, field := range msgFields {
		if !field.IsEmbedded {
			order := currProtoMessages.GetFieldNum(msgName, field.Name)
			field.Order = int(order)
			newFields = append(newFields, field)
			continue
		}
		currProtoMessages.RemoveFieldNum(msgName, field.Name)

		embeddedMsg := messageMap[field.TypeName]
		embeddedFields := resolveEmbedded(embeddedMsg.Fields, messageMap, currProtoMessages, embeddedMsg.Name)

		for _, embeddedField := range embeddedFields {
			order := currProtoMessages.GetFieldNum(msgName, embeddedField.Name)
			embeddedField.Order = int(order)
			newFields = append(newFields, embeddedField)
		}

	}
	return newFields
}

func toProtoFieldTypeName(f *types.Var) string {
	switch f.Type().Underlying().(type) {
	case *types.Basic:
		name := f.Type().String()
		return normalizeType(name)
	case *types.Slice:
		name := splitNameHelper(f)
		return normalizeType(strings.TrimLeft(name, "[]"))

	case *types.Pointer, *types.Struct:
		name := splitNameHelper(f)
		return normalizeType(name)
	}
	return f.Type().String()
}

func splitNameHelper(f *types.Var) string {
	// TODO: this is ugly. Find another way of getting field type name.
	parts := strings.Split(f.Type().String(), ".")

	name := parts[len(parts)-1]

	if name[0] == '*' {
		name = name[1:]
	}
	return name
}

func normalizeType(name string) string {
	switch name {
	case "int":
		return "int64"
	case "float32":
		return "float"
	case "float64":
		return "double"
	default:
		return name
	}
}

func isElasticsearchNoSource(tagString string) bool {
	if tagString == "" {
		return false
	}

	tags := strings.Split(tagString, " ")
	for _, tag := range tags {
		tagSplit := strings.Split(tag, ":")
		if len(tagSplit) != 2 || tagSplit[0] != "elasticsearch" {
			continue
		}
		cleanTag := strings.Trim(tagSplit[1], "\"")
		for _, val := range strings.Split(cleanTag, ",") {
			if val == "no_source" {
				return true
			}
		}
	}
	return false
}

func isRepeated(f *types.Var) bool {
	_, ok := f.Type().Underlying().(*types.Slice)
	return ok
}

func toProtoFieldName(name string, useSnakeFieldNames bool) string {
	if len(name) == 2 {
		return strings.ToLower(name)
	}
	r, n := utf8.DecodeRuneInString(name)
	val := string(unicode.ToLower(r)) + name[n:]
	if useSnakeFieldNames {
		return gostringconverters.SnakeCase(val)
	}
	return val
}

func escapeQuotes(tag string) string {
	return strings.Replace(tag, `"`, `\"`, -1)
}

var FUNC_MAP = template.FuncMap{
	"escapeQuotes": escapeQuotes,
}

func writeOutput(msgs []message, path string) error {
	msgTemplate := `syntax = "proto3";
package proto;

import "tagger/tagger.proto";

{{range .}}
//easyjson:json
message {{.Name}} {
{{- range .Fields}}
{{- if .IsRepeated}}
  repeated {{.TypeName}} {{.Name}} = {{if ne .Tags "" }}{{.Order}} [(tagger.tags) = "{{escapeQuotes .Tags}}"]; {{ else }}{{.Order}};{{ end }}
{{- else}}
  {{.TypeName}} {{.Name}} = {{if ne .Tags "" }}{{.Order}} [(tagger.tags) = "{{escapeQuotes .Tags}}"]; {{ else }}{{.Order}};{{ end }}
{{- end}}
{{- end}}
}
{{end}}
`
	tmpl, err := template.New("test").Funcs(FUNC_MAP).Parse(msgTemplate)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(filepath.Join(path, "output.proto"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, msgs)
}

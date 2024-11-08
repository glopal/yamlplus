package yamlp

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	AddTagResolver("!tmpl", tmplResolver)
}

func tmplResolver(rc ResolveContext) (*yqlib.CandidateNode, error) {
	val, err := renderTemplate(rc.Target.Value, rc)
	if err != nil {
		return nil, err
	}

	node := &yqlib.CandidateNode{}
	node.Kind = yqlib.ScalarNode
	node.Tag = "!!str"
	node.Value = val

	return node, nil
}

func renderTemplate(tmpl string, rc ResolveContext) (string, error) {
	declTmpl, funcMap := processVariables(rc.Vars)

	t := template.New("")
	t.Funcs(funcMap)

	t, err := t.Parse(declTmpl + tmpl)
	if err != nil {
		return "", err
	}

	t.Option("missingkey=error")

	data, err := rc.Ctx.Interface()
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, data)

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func processVariables(vars map[string]*ContextNode) (string, template.FuncMap) {
	funcMap := map[string]any{}
	declTmpl := ""

	for name, node := range vars {
		declTmpl += fmt.Sprintf("{{$%s := %s}}", name, name)
		funcMap[name] = func() interface{} {
			val, _ := node.Interface()
			return val
		}
	}

	return declTmpl, funcMap
}

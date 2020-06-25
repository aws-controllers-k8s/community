package template

import (
	"io/ioutil"
	"path/filepath"
	ttpl "text/template"
)

type DockerTemplateVars struct {
	ServiceAlias string
}

func NewDockerfileTemplate(tplDir string) (*ttpl.Template, error) {
	tplPath := filepath.Join(tplDir, "Dockerfile.tpl")
	tplContents, err := ioutil.ReadFile(tplPath)
	if err != nil {
		return nil, err
	}
	t := ttpl.New("Dockerfile")
	if t, err = t.Parse(string(tplContents)); err != nil {
		return nil, err
	}
	includes := []string{
		"boilerplate_hash",
	}
	for _, include := range includes {
		if t, err = IncludeTemplate(t, tplDir, include); err != nil {
			return nil, err
		}
	}
	return t, nil
}


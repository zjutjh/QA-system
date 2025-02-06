// plugins/plugin1.go
package plugins

import (
	"QA-System/internal/pkg/extension"
	"fmt"
)

const (
	PluginName  = "plugin1"
	Version     = "1.0.0"
	Author      = "Author1"
	Description = "This is plugin 1"
)

type Plugin1 struct{}

func (p *Plugin1) GetMetadata() extension.PluginMetadata {
	return extension.PluginMetadata{
		Name:        PluginName,
		Version:     Version,
		Author:      Author,
		Description: Description,
	}
}

func (p *Plugin1) Execute(params map[string]interface{}) error {
	fmt.Println("Plugin1 executing with params:", params)
	params["processed_by"] = "plugin1"
	return nil
}

func init() {
	extension.RegisterPlugin(&Plugin1{})
}

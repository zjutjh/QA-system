// plugins/plugin2.go
package plugins

import (
	"QA-System/internal/pkg/extension"
	"fmt"
)

type Plugin2 struct{}

func (p *Plugin2) GetMetadata() extension.PluginMetadata {
	return extension.PluginMetadata{
		Name:        "plugin2",
		Version:     "1.0.0",
		Author:      "Author2",
		Description: "This is plugin 2",
	}
} //这是另一种写metaData的方式

func (p *Plugin2) Execute(params map[string]interface{}) error {
	fmt.Println("Plugin2 executing with params:", params)
	params["processed_by"] = "plugin2"
	return nil
}

func init() {
	extension.RegisterPlugin(&Plugin2{})
}

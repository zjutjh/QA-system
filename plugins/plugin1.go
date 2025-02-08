package plugins

import (
	"QA-System/internal/pkg/extension"
	"fmt"
	"time"
)

const (
	PluginName  = "plugin1"
	Version     = "0.0.1"
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

func (p *Plugin1) Execute() error {
	// 插件的主要逻辑
	fmt.Println("Plugin1 executing at", time.Now().Format(time.RFC3339))
	// 这里可以添加插件的具体功能代码
	return nil
}

func init() {
	extension.RegisterPlugin(&Plugin1{})
}

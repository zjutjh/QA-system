package plugins

import (
	"fmt"
	"time"

	"QA-System/internal/pkg/extension"
)

const (
	PluginName  = "plugin1"          // PluginName插件名
	Version     = "0.0.1"            // Version版本号（学习AWS的版本名谢谢）
	Author      = "Author1"          // Author作者
	Description = "This is plugin 1" // Descrtption 插件描述
)

// Plugin1 示例插件1的结构，要给manager的
type Plugin1 struct{}

func (p *Plugin1) GetMetadata() extension.PluginMetadata {
	_ = p
	return extension.PluginMetadata{
		Name:        PluginName,
		Version:     Version,
		Author:      Author,
		Description: Description,
	}
}

func (p *Plugin1) Execute() error {
	_ = p
	// 插件的主要逻辑
	fmt.Println("Plugin1 executing at", time.Now().Format(time.RFC3339))
	// 这里可以添加插件的具体功能代码
	return nil
}

func init() {
	extension.RegisterPlugin(&Plugin1{})
}

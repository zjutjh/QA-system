package extension

// PluginChain 插件链
type PluginChain struct {
	plugins []Plugin
}

// NewPluginChain 创建新的插件链
func NewPluginChain() *PluginChain {
	return &PluginChain{
		plugins: make([]Plugin, 0),
	}
}

// Add 添加插件到链中
func (pc *PluginChain) Add(plugin Plugin) *PluginChain {
	pc.plugins = append(pc.plugins, plugin)
	return pc
}

// Execute 按顺序执行插件链中的所有插件
func (pc *PluginChain) Execute(params map[string]any) error {
	for _, plugin := range pc.plugins {
		if err := plugin.Execute(params); err != nil {
			return err
		}
	}
	return nil
}

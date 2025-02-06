// extension/manager.go
package extension

import (
	"QA-System/internal/global/config"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

var (
	plugins = make(map[string]Plugin) // 插件名称 -> 插件实例
	mu      sync.Mutex
)

// extension包自己的init函数，用来看一眼extension是不是被导入了
func init() {
	fmt.Println("插件包加载模块初始化成功 黑暗森林威慑建立")
}

// 在 RegisterPlugin()
func RegisterPlugin(p Plugin) {
	mu.Lock()
	defer mu.Unlock()
	metadata := p.GetMetadata()
	plugins[metadata.Name] = p
	zap.L().Info("Plugin registered successfully",
		zap.String("name", metadata.Name),
		zap.String("version", metadata.Version))
}

// GetPlugin 获取插件
func GetPlugin(name string) (Plugin, bool) {
	mu.Lock()
	defer mu.Unlock()
	p, ok := plugins[name]
	return p, ok
}

// LoadPlugins 从配置文件中加载插件并返回插件实例列表
func LoadPlugins() ([]Plugin, error) {
	pluginNames := config.Config.GetStringSlice("plugins.order")
	zap.L().Info("Loading plugins from config",
		zap.Strings("plugin_names", pluginNames))
	var pluginList []Plugin

	for _, name := range pluginNames {
		p, ok := GetPlugin(name)
		if !ok {
			return nil, fmt.Errorf("plugin %s not found", name)
		}
		pluginList = append(pluginList, p)
	}

	return pluginList, nil
}

// ExecutePlugins 依次执行插件链
func ExecutePlugins(params map[string]interface{}) error {

	pluginList, err := LoadPlugins()
	if err != nil {
		return err
	}

	fmt.Println(pluginList)
	for _, p := range pluginList {
		err := p.Execute(params)
		if err != nil {
			return fmt.Errorf("plugin %s failed: %v", p.GetMetadata().Name, err)
		}
	}

	return nil
}

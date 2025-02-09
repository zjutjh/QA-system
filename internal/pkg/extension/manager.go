// extension/manager.go
package extension

import (
	"fmt"
	"sync"

	"QA-System/internal/global/config"
	"go.uber.org/zap"
)

var (
	plugins = make(map[string]Plugin) // 插件名称 -> 插件实例
	mu      sync.Mutex
)

// extension包自己的init函数，用来看一眼extension是不是被导入了
func init() {
	fmt.Println("插件包加载模块初始化成功 阶梯计划成功")
}

// RegisterPlugin 向插件管理器注册插件
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
func ExecutePlugins() error {
	pluginList, err := LoadPlugins()
	if err != nil {
		return err
	}

	// 启动所有插件的后台服务
	for _, p := range pluginList {
		metadata := p.GetMetadata()
		zap.L().Info("Starting plugin service",
			zap.String("name", metadata.Name),
			zap.String("version", metadata.Version))

		// 使用goroutine启动插件服务
		go func(plugin Plugin) {
			if err := plugin.Execute(); err != nil { // 移除了 nil 参数
				zap.L().Error("Plugin service failed",
					zap.String("name", plugin.GetMetadata().Name),
					zap.Error(err))
			}
		}(p)
	}

	return nil
}

package plugins

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var _ Plugin = (*PluginDebug)(nil)
var _ PluginConfig = (*PluginDebugConfig)(nil)

type PluginDebugConfig struct {
	// The prefix can be used to properly identify log messages
	Prefix string `mapstructure:"prefix"`

	// You can use the Args map to overwrite/add arguments passed to the command
	// for debugging purposes
	Args map[string]interface{} `mapstructure:"args"`
}

func (c *PluginDebugConfig) NewPlugin() (Plugin, error) {
	return NewPluginDebug(c), nil
}

func (c *PluginDebugConfig) IsUnique() bool {
	return false
}

type PluginDebug struct {
	config *PluginDebugConfig
}

func NewPluginDebug(config *PluginDebugConfig) *PluginDebug {
	plugin := &PluginDebug{
		config: config,
	}

	return plugin
}

func (p *PluginDebug) MountRoutes(engine *gin.Engine, listenerRoute string, listenerHandler func(args map[string]interface{}) (interface{}, error)) {
	// NOOP for this plugin
}

func (p *PluginDebug) HookPreExecute(args map[string]interface{}) (map[string]interface{}, error) {
	// Merge args with any provided ones
	for key, val := range p.config.Args {
		args[key] = val
	}

	logrus.WithField("args", args).Warnf("[%s] PRE-EXECUTE", p.config.Prefix)

	return args, nil
}

package pkg

import (
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const keyAuthDefaultHTTPBasicUser = "gte"
const keyAuthApiKeyQuery = "__gteApiKey"
const keyArgsHeadersKey = "__gteHeaders"

func MountRoutes(engine *gin.Engine, config *Config) {
	storageCache := new(sync.Map)

	for route, listenerConfig := range config.Listeners {
		log := logrus.WithField("listener", route)

		listener := compileListener(&config.Defaults, listenerConfig, route, false, storageCache)
		handler := getGinListenerHandler(listener)
		mountRoutesByMethod(engine, listener.config.Methods, route, handler)

		for _, plugin := range listener.plugins {
			if plugin, ok := plugin.(PluginHookMountRoutes); ok {
				plugin.HookMountRoutes(engine, listener)
			}
		}

		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			log.WithFields(logrus.Fields{
				"config": spew.Sdump(listener.config),
			}).Debug("added listener")
		} else {
			log.Info("added listener")
		}
	}
}

func mountRoutesByMethod(engine *gin.Engine, methods []string, route string, handler gin.HandlerFunc) {
	if len(methods) == 0 {
		engine.GET(route, handler)
		engine.POST(route, handler)
	} else {
		for _, method := range methods {
			engine.Handle(method, route, handler)
		}
	}
}

// @formatter:off
/// [listener-response]
type ListenerResponse struct {
	*ExecCommandResult
	Storage            *StorageEntry     `json:"storage,omitempty"`
	Error              *string           `json:"error,omitempty"`
	ErrorHandlerResult *ListenerResponse `json:"errorHandlerResult,omitempty"`
}

/// [listener-response]
// @formatter:on

var regexListenerRouteCleaner = regexp.MustCompile(`[\W]`)

func getGinListenerHandler(listener *CompiledListener) gin.HandlerFunc {
	return func(c *gin.Context) {
		handled, args := prepareListenerRequestHandling(c, listener.config.Auth)
		if handled {
			return
		}

		ctxHandled, response, err := listener.HandleRequest(c, args)
		if ctxHandled {
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, response)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func prepareListenerRequestHandling(
	c *gin.Context,
	authConfigs []*AuthConfig,
) (bool, map[string]interface{}) {
	if err := verifyAuth(c, authConfigs); err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return true, nil
	}

	args, err := extractArgsFromGinContext(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.WithMessage(err, "failed to extract args from request"))
		return true, nil
	}
	return false, args
}

func extractArgsFromGinContext(c *gin.Context) (map[string]interface{}, error) {
	args := make(map[string]interface{})

	// Use route params, if any
	for _, param := range c.Params {
		args[param.Key] = param.Value
	}

	// Add headers to args
	{
		headerMap := make(map[string]interface{})
		for k := range c.Request.Header {
			headerMap[strings.ToLower(k)] = c.GetHeader(k)
		}
		args[keyArgsHeadersKey] = headerMap
	}

	if c.Request.Method != http.MethodGet {
		b := binding.Default(c.Request.Method, c.ContentType())
		if b == binding.Form || b == binding.FormMultipart {
			queryMap := make(map[string][]string)
			if err := c.ShouldBindWith(&queryMap, b); err != nil {
				return nil, errors.WithMessage(err, "failed to parse request form body")
			}
			for key, vals := range queryMap {
				if len(vals) > 0 {
					args[key] = vals[len(vals)-1]
				} else {
					args[key] = true
				}

				args["_form_"+key] = vals
			}
		} else {
			if err := c.ShouldBindWith(&args, b); err != nil {
				return nil, errors.WithMessage(err, "failed to parse request body")
			}
		}
	}

	// Always bind query
	{
		queryMap := make(map[string][]string)
		if err := c.ShouldBindQuery(&queryMap); err != nil {
			return nil, errors.WithMessage(err, "failed to parse request query")
		}
		for key, vals := range queryMap {
			if len(vals) > 0 {
				args[key] = vals[len(vals)-1]
			} else {
				args[key] = true
			}

			args["_query_"+key] = vals
		}
	}

	return args, nil
}

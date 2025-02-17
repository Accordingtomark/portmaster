package core

import (
	"net/http"

	"github.com/safing/portbase/api"
	"github.com/safing/portbase/log"
	"github.com/safing/portbase/modules"
	"github.com/safing/portbase/utils/debug"
	"github.com/safing/portmaster/resolver"
	"github.com/safing/portmaster/status"
	"github.com/safing/portmaster/updates"
)

func registerAPIEndpoints() error {
	if err := api.RegisterEndpoint(api.Endpoint{
		Path:        "core/shutdown",
		Write:       api.PermitSelf,
		BelongsTo:   module,
		ActionFunc:  shutdown,
		Name:        "Shut Down Portmaster",
		Description: "Shut down the Portmaster Core Service and all UI components.",
	}); err != nil {
		return err
	}

	if err := api.RegisterEndpoint(api.Endpoint{
		Path:        "core/restart",
		Write:       api.PermitAdmin,
		BelongsTo:   module,
		ActionFunc:  restart,
		Name:        "Restart Portmaster",
		Description: "Restart the Portmaster Core Service.",
	}); err != nil {
		return err
	}

	if err := api.RegisterEndpoint(api.Endpoint{
		Path:        "debug/core",
		Read:        api.PermitAnyone,
		BelongsTo:   module,
		DataFunc:    debugInfo,
		Name:        "Get Debug Information",
		Description: "Returns network debugging information, similar to debug/info, but with system status data.",
		Parameters: []api.Parameter{{
			Method:      http.MethodGet,
			Field:       "style",
			Value:       "github",
			Description: "Specify the formatting style. The default is simple markdown formatting.",
		}},
	}); err != nil {
		return err
	}

	return nil
}

// shutdown shuts the Portmaster down.
func shutdown(_ *api.Request) (msg string, err error) {
	log.Warning("core: user requested shutdown via action")

	// Do not run in worker, as this would block itself here.
	go modules.Shutdown() //nolint:errcheck

	return "shutdown initiated", nil
}

// restart restarts the Portmaster.
func restart(_ *api.Request) (msg string, err error) {
	log.Info("core: user requested restart via action")

	// Let the updates module handle restarting.
	updates.RestartNow()

	return "restart initiated", nil
}

// debugInfo returns the debugging information for support requests.
func debugInfo(ar *api.Request) (data []byte, err error) {
	// Create debug information helper.
	di := new(debug.Info)
	di.Style = ar.Request.URL.Query().Get("style")

	// Add debug information.
	di.AddVersionInfo()
	di.AddPlatformInfo(ar.Context())
	status.AddToDebugInfo(di)
	resolver.AddToDebugInfo(di)
	di.AddLastReportedModuleError()
	di.AddLastUnexpectedLogs()
	di.AddGoroutineStack()

	// Return data.
	return di.Bytes(), nil
}

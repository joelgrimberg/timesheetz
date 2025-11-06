package datalayer

import (
	"timesheet/internal/api"
	"timesheet/internal/config"
	"timesheet/internal/db"
	"timesheet/internal/logging"
)

var dataLayerInstance db.DataLayer

// GetDataLayer returns the appropriate data layer based on configuration
// This is the main entry point for all data operations
func GetDataLayer() db.DataLayer {
	// Return cached instance if available
	if dataLayerInstance != nil {
		return dataLayerInstance
	}

	apiMode := config.GetAPIMode()

	switch apiMode {
	case "local":
		// Use local database only
		dataLayerInstance = &db.LocalDBLayer{}
		logging.Log("Using local database mode")

	case "remote":
		// Use remote API only
		apiClient, err := api.GetClient()
		if err != nil {
			logging.Log("Failed to create API client, falling back to local: %v", err)
			dataLayerInstance = &db.LocalDBLayer{}
		} else {
			dataLayerInstance = api.NewClientAdapter(apiClient)
			logging.Log("Using remote API mode")
		}

	case "dual":
		// Use both local DB and remote API
		localLayer := &db.LocalDBLayer{}
		apiClient, err := api.GetClient()
		if err != nil {
			logging.Log("Failed to create API client for dual mode, using local only: %v", err)
			dataLayerInstance = localLayer
		} else {
			remoteLayer := api.NewClientAdapter(apiClient)
			dataLayerInstance = db.NewDualLayer(localLayer, remoteLayer)
			logging.Log("Using dual mode (local DB + remote API)")
		}

	default:
		// Default to local mode
		logging.Log("Unknown apiMode '%s', defaulting to local", apiMode)
		dataLayerInstance = &db.LocalDBLayer{}
	}

	return dataLayerInstance
}

// ResetDataLayer resets the cached data layer instance (for testing)
func ResetDataLayer() {
	dataLayerInstance = nil
}


package HTML_Handler

import (
	"encoding/json"
	"net/http"
)

type activityResults struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ===========================
// API Server related activities
// ============================

func handleStartActivity(w http.ResponseWriter, r *http.Request) {
	// Implementation for handling start Activity
}

func handleStopActivity(w http.ResponseWriter, r *http.Request) {
	// Implementation for handling stop Activity
}

func handleRestartActivity(w http.ResponseWriter, r *http.Request) {
	// Implementation for handling restart Activity
}

func handleBackupActivity(w http.ResponseWriter, r *http.Request) {
	// Implementation for handling backup Activity
}

// ==========================
// Computer related activities
// ==========================

func handleShutdownActivity(w http.ResponseWriter, r *http.Request) {
	if !user_is_logged(r) {
		w.WriteHeader(401)
		res, _ := json.Marshal(activityResults{
			Status:  "fail",
			Message: "You are not logged in",
		})
		w.Write(res)
		return
	}
	if webHosterRunning != nil {
		webHosterRunning <- false
	}
	res, _ := json.Marshal(activityResults{
		Status:  "success",
		Message: "Computer is shutting down, the connection will be lost!!",
	})
	w.Write(res)
}
func handleRebootActivity(w http.ResponseWriter, r *http.Request) {
	if !user_is_logged(r) {
		w.WriteHeader(401)
		res, _ := json.Marshal(activityResults{
			Status:  "fail",
			Message: "You are not logged in",
		})
		w.Write(res)
		return
	}
	res, _ := json.Marshal(activityResults{
		Status:  "success",
		Message: "Computer is rebooting, the connection will be lost! In case that the server does't go back online, verify the computer startup programs",
	})
	w.Write(res)
}

// ==========================
// Unknown Activity handler
// ==========================
func handleUnknownActivity(w http.ResponseWriter, r *http.Request, activity string) {
	println("Unkown activity: ", activity)
	w.WriteHeader(400) // Bad request
	res, _ := json.Marshal(activityResults{
		Status:  "fail",
		Message: "Unkown command",
	})
	w.Write(res)
}

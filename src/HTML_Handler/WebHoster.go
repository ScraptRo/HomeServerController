package HTML_Handler

import (
	"ServerController/src/API_Handler"
	common "ServerController/src/Common"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var webHosterRunning = chan<- bool(nil)
var server *http.Server
var wg sync.WaitGroup
var Server_name string = "Home Server Controller"

type command struct {
	commandDescription    string
	commandHandler        func(http.ResponseWriter, *http.Request, []string)
	availableWithoutLogin bool
}

var activitiesMap = map[string]func(http.ResponseWriter, *http.Request){
	"start":    handleStartActivity,
	"stop":     handleStopActivity,
	"restart":  handleRestartActivity,
	"backup":   handleBackupActivity,
	"shutdown": handleShutdownActivity,
	"reboot":   handleRebootActivity,
}

var commandsMap = map[string]command{
	"login":           {"Let the user login based on credentials, and gives permisions based on user details, determined by the admin { login [username] [password] }", handleLoginCommand, true},
	"logout":          {"Logs out the user, giving him access to switch to other accounts", handleLogOutCommand, false},
	"whoami":          {"Specify the account you are connected", handleWhoAmICommand, false},
	"change_password": {"Changes the password of the user that you are logged in as { change_password [new_password]}", handleChangePassword, false},
	"add_user":        {"Creates a new user { add_user [username] [password] [is_admin](optional, default false) [admin_grade](optional, default 1)}", handleAddUserCommand, false},
}

func init() {
	commandsMap["help"] = command{"Show this help message. Also can describe other commands by tipping help [command]", handleHelpCommand, true}
}

func user_is_logged(r *http.Request) bool {
	cookie, err := r.Cookie("SVC_username")
	if err != nil {
		return false
	}
	if cookie.Value == "" {
		return false
	}
	return true
}

func GetFileContentsAsString(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func getHomePage(w http.ResponseWriter, r *http.Request) {
	file, err := GetFileContentsAsString("res/web_files/index.html")
	if err != nil {
		http.Error(w, "Could not open template", http.StatusInternalServerError)
		return
	}
	t, _ := template.New("setup").Parse(file)
	t.Execute(w, nil)
}

func getStyles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	file, err := GetFileContentsAsString("res/web_files/styles.css")
	if err != nil {
		http.Error(w, "Could not open CSS file", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(file))
}

func getScripts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	file, err := GetFileContentsAsString("res/web_files/scripts.js")
	if err != nil {
		http.Error(w, "Could not open JS file", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(file))
}

func handleCommands(w http.ResponseWriter, r *http.Request) {
	wg.Add(1)
	defer wg.Done()

	w.Header().Set("Content-Type", "application/json")
	const maxBodySize = 10 * 1024 // 10KB limit
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodySize))

	// Handle potential errors when reading the body
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		if err.Error() == "http: request body too large" {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
		}
		return
	}

	if len(body) == 0 {
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}

	type CommandRequest struct {
		Command    string   `json:"command"`
		Parameters []string `json:"parameters"`
	}

	var m CommandRequest
	if err := json.Unmarshal(body, &m); err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	if handler, exists := commandsMap[m.Command]; exists {
		handler.commandHandler(w, r, m.Parameters)
	} else {
		handleUnknownCommand(w, m.Command)
	}
}

func handleActivities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	const maxBodySize = 10 * 1024 // 10KB limit
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodySize))

	// Handle potential errors when reading the body
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		if err.Error() == "http: request body too large" {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
		}
		return
	}
	if len(body) == 0 {
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}
	type ActivityRequest struct {
		Activity string `json:"activity"`
	}
	var m ActivityRequest
	if err := json.Unmarshal(body, &m); err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	if handler, exists := activitiesMap[m.Activity]; exists {
		handler(w, r)
	} else {
		handleUnknownActivity(w, r, m.Activity)
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cookie, err := r.Cookie("SVC_username")
	if err != nil {
		// Cookie doesn't exist - user is not logged in
		w.WriteHeader(401) // Unauthorized
		res, _ := json.Marshal(commandResults{
			Status:  "fail",
			Message: "Not logged in",
		})
		w.Write(res)
		return
	}

	// Your JS expects these fields:
	_, totalMemory, _ := common.GetMemoryUsage()
	response := map[string]interface{}{
		"status":      "success",
		"username":    cookie.Value,
		"port":        "5050",
		"startTime":   API_Handler.StartTime,
		"cpu":         common.GetCPUUsage(),
		"memory":      totalMemory,
		"connections": API_Handler.GetNumberOfConnections(),
	}

	jsonResponse, _ := json.Marshal(response)
	w.Write(jsonResponse)
}

func handServerDetails(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":      "success",
		"server_name": Server_name,
		"server_uid":  common.GetOrCreateID(),
		"port":        API_Handler.GetServerPort(),
	}
	jsonResponse, _ := json.Marshal(response)
	w.Write(jsonResponse)
}

func StartWebHoster(serverRunning chan<- bool) {
	webHosterRunning = serverRunning
	http.HandleFunc("/", getHomePage)
	http.HandleFunc("/styles.css", getStyles)
	http.HandleFunc("/scripts.js", getScripts)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/api/command", handleCommands)
	http.HandleFunc("/api/activities", handleActivities)
	http.HandleFunc("/api/status", handleStatus)
	http.HandleFunc("/WebServerController/details", handServerDetails)

	ipAddress := common.GetOutboundIP()
	port := "8080"
	addr := ipAddress.String() + ":" + port
	fmt.Println("Web server starting at: http://" + addr)
	server = &http.Server{
		Addr:         addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	server.ListenAndServe()
}

func StopWebHoster() {
	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		} else {
			fmt.Println("Web server stopped.")
		}
	}
}

func WaitForWebHoster() {
	wg.Wait()
}

func handleHelpCommand(w http.ResponseWriter, r *http.Request, parameters []string) {
	type helpResult struct {
		Status  string   `json:"status"`
		Message []string `json:"message"`
	}
	var res helpResult
	res.Status = "success"
	if len(parameters) != 0 {
		for parameter := range parameters {
			res.Message = append(res.Message, parameters[parameter]+" : "+commandsMap[parameters[parameter]].commandDescription+"\n")
		}
	} else {
		for key, command := range commandsMap {
			res.Message = append(res.Message, key+" : "+command.commandDescription+"\n")
		}
	}
	data, _ := json.Marshal(res)
	w.Write([]byte(data))
}

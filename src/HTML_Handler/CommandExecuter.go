package HTML_Handler

import (
	"ServerController/src/User_Handler"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type commandResults struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	Username   string `json:"username,omitempty"`
	UpdateUser bool   `json:"update_user,omitempty"`
}

func handleLoginCommand(w http.ResponseWriter, r *http.Request, parameters []string) {
	w.Header().Set("Content-Type", "application/json")
	if len(parameters) != 2 {
		w.WriteHeader(400)
		res, _ := json.Marshal(commandResults{
			Status:  "fail",
			Message: "Invalid number of parameters, correct usage: login [username] [password]",
		})
		w.Write(res)
		return
	}
	if User_Handler.Authenticate_user(parameters[0], parameters[1]) {
		http.SetCookie(w, &http.Cookie{
			Name:     "SVC_username",
			Value:    parameters[0],
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(24 * time.Hour),
		})
		w.WriteHeader(200)
		res, _ := json.Marshal(commandResults{
			Status:  "success",
			Message: "User logged in successfully",
		})
		w.Write(res)
		return
	}
	w.WriteHeader(401)
	res, _ := json.Marshal(commandResults{
		Status:  "fail",
		Message: "Unable to login: username or password invalid",
	})
	w.Write(res)
}

func handleLogOutCommand(w http.ResponseWriter, r *http.Request, parameters []string) {
	w.Header().Set("Content-Type", "application/json")
	http.SetCookie(w, &http.Cookie{
		Name:     "SVC_username",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
	})
	w.WriteHeader(200)
	res, _ := json.Marshal(commandResults{
		Status:  "success",
		Message: "Logged out successfully",
	})
	w.Write(res)
}

func handleChangePassword(w http.ResponseWriter, r *http.Request, parameters []string) {
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
	if len(parameters) != 1 {
		w.WriteHeader(400) // Bad Request
		res, _ := json.Marshal(commandResults{
			Status:  "fail",
			Message: "Invalid number of parameters",
		})
		w.Write(res)
		return
	}
	if User_Handler.User_exists(cookie.Value) {
		User_Handler.Change_password(cookie.Value, parameters[0])
		w.WriteHeader(200)
		res, _ := json.Marshal(commandResults{
			Status:  "success",
			Message: "Password changed successfully",
		})
		w.Write(res)
	} else {
		w.WriteHeader(401) // Unauthorized
		res, _ := json.Marshal(commandResults{
			Status:     "fail",
			Message:    "You are not loged in",
			UpdateUser: true,
		})
		w.Write(res)
		return
	}

}

func handleWhoAmICommand(w http.ResponseWriter, r *http.Request, parameters []string) {
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
	w.WriteHeader(200)
	res, _ := json.Marshal(commandResults{
		Status:   "success",
		Message:  "Loged in as " + cookie.Value,
		Username: cookie.Value,
	})
	w.Write(res)
}

func handleAddUserCommand(w http.ResponseWriter, r *http.Request, parameters []string) {
	_, err := r.Cookie("SVC_username")
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

	if len(parameters) < 2 {
		w.WriteHeader(400) // Unauthorized
		res, _ := json.Marshal(commandResults{
			Status:  "fail",
			Message: "Invalid number of parameters : add_user [username] [password] [is_admin](optional, default false) [admin_grade](optional, default 1)",
		})
		w.Write(res)
		return
	}

	var is_admin = false
	var admin_grade uint8 = 1
	if len(parameters) > 2 {
		if parameters[2] == "true" {
			is_admin = true
		}
		if len(parameters) > 3 {
			grade, err := strconv.Atoi(parameters[3])
			if err != nil {
				w.WriteHeader(400) // Unauthorized
				res, _ := json.Marshal(commandResults{
					Status:  "fail",
					Message: "Invalid invalid grade parameter",
				})
				w.Write(res)
				return
			}
			admin_grade = uint8(grade)
		}
	}
	w.WriteHeader(200)
	var results commandResults
	if User_Handler.Add_user(parameters[0], parameters[1], is_admin, admin_grade) {
		results.Status = "success"
		results.Message = "User added successfully"
	} else {
		results.Status = "fail"
		results.Message = "User already exists"
	}
	res, _ := json.Marshal(results)
	w.Write(res)
}

func handleUnknownCommand(w http.ResponseWriter, command string) {
	w.WriteHeader(400)
	println("Unknown command: " + command)
	res, _ := json.Marshal(commandResults{
		Status:  "fail",
		Message: "Unkown command",
	})
	w.Write(res)
}

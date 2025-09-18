package User_Handler

import (
	"encoding/json"
	"os"
	"time"
)

type Register_Request struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Salt       string `json:"salt"`
	Request_At string `json:"request_at"`
}

var Loaded_Requests map[string]Register_Request

func Load_requests() {
	file, err := os.ReadFile("res/config_files/register_requests.json")
	if err != nil {
		err = os.MkdirAll("res/config_files", os.ModePerm)
		if err != nil {
			println("Could not create config_files directory: " + err.Error())
		}
		Loaded_Requests = map[string]Register_Request{}
	} else {
		// Parse JSON and load users
		_ = json.Unmarshal(file, &Loaded_Requests)
	}

}

func save_requests() {
	data, err := json.MarshalIndent(Loaded_Requests, "", "  ")
	if err != nil {
		println("Could not marshal users data: " + err.Error())
		return
	}
	err = os.WriteFile("res/config_files/register_requests.json", data, 0644)
	if err != nil {
		println("Could not write users data to file: " + err.Error())
		return
	}
}

func Request_exists(username string) bool {
	_, exists := Loaded_Requests[username]
	return exists
}

func Insert_account_request(username string, password string) (bool, string) {
	if User_exists(username) {
		return false, "Username already exists"
	}
	if Request_exists(username) {
		return false, "Request already exists with this username"
	}
	salt := generateSalt()
	Loaded_Requests[username] = Register_Request{
		Username:   username,
		Password:   hashPassword(password, salt),
		Salt:       salt,
		Request_At: time.Now().Format(time.RFC3339),
	}
	save_requests()
	return true, "Request placed successfully"
}

func Accept_account_request(username string, admin bool, grade uint8) {
	request := Loaded_Requests[username]
	LoadedUsers[username] = User{
		Username:    username,
		Password:    request.Password,
		Salt:        request.Salt,
		CreatedAt:   time.Now().Format(time.RFC3339),
		LastLogin:   time.Time{}.Format(time.RFC3339),
		Admin:       admin,
		Admin_Grade: grade,
	}
	delete(Loaded_Requests, username)
	Save_users()
	save_requests()
}

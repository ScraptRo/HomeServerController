package API_Handler

import (
	"ServerController/src/Internal_Process_Handler"
	"ServerController/src/User_Handler"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"strconv"
)

const (
	Success      = "success"
	Fail         = "fail"
	Unauthorized = "Unauthorized"
)

type response struct {
	Status       string `json:"status"`
	Process_Type string `json:"process_type"`
	Message      string `json:"message"`
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func list_private_scripts() []string {
	var result []string
	publicEntries, err := os.ReadDir("scripts/private/")
	if err != nil {
		return result
	}
	for i := 0; i < len(publicEntries); i++ {
		result = append(result, publicEntries[i].Name())
	}
	return result
}

func list_public_scripts() []string {
	var result []string
	publicEntries, err := os.ReadDir("scripts/public/")
	if err != nil {
		return result
	}
	for i := 0; i < len(publicEntries); i++ {
		result = append(result, publicEntries[i].Name())
	}
	return result
}

func close_user_connection(request *request_format, info *user_info) []byte {
	info.close_connection = true
	var res response
	res.Status = Success
	res.Process_Type = "exit"
	r, _ := json.Marshal(res)
	return r
}

func runCommandInConsole(request *request_format, info *user_info) []byte {
	var res response
	res.Process_Type = "console_cmd"

	if !info.is_admin {
		res.Status = Unauthorized
		res.Message = "You need to be logged in to be able to run Shell commands"
		b, _ := json.Marshal(res)
		return b
	}

	if len(request.Args) < 1 {
		res.Status = Unauthorized
		res.Message = "You need at least one command: command, args..."
		b, _ := json.Marshal(res)
		return b
	}

	msg := Internal_Process_Handler.RunCommand(request.Args)

	if msg == "" {
		res.Status = Fail
		res.Message = "Empty result"
	} else {
		res.Status = Success
		res.Message = msg
	}

	r, _ := json.Marshal(res)
	return r
}

func login_attempt(request *request_format, info *user_info) []byte {
	var res response
	res.Process_Type = "login_attempt"
	if len(request.Args) != 2 {
		res.Status = Unauthorized
		res.Message = "You need 2 arguments: username, password"
		output, _ := json.Marshal(res)
		return output
	}
	if User_Handler.Authenticate_user(request.Args[0], request.Args[1]) {
		user := User_Handler.LoadedUsers[request.Args[0]]
		info.is_admin = user.Admin
		info.username = user.Username

		res.Status = "Success"
		res.Message = "Logged in successfully"

		output, _ := json.Marshal(res)
		return output
	} else {
		res.Status = Fail
		res.Message = "Invalid username or password"
		output, _ := json.Marshal(res)
		return output
	}
}

func request_account(request *request_format, info *user_info) []byte {
	var res response
	res.Process_Type = "request_account"
	if len(request.Args) != 2 {
		res.Status = Unauthorized
		res.Message = "You need 2 arguments: username, password"
		output, _ := json.Marshal(res)
		return output
	}
	result, message := User_Handler.Insert_account_request(request.Args[0], request.Args[1])

	if result {
		res.Status = Success
	} else {
		res.Status = Fail
	}

	res.Message = message
	output, _ := json.Marshal(res)

	return output
}

func list_account_requests(request *request_format, info *user_info) []byte {
	var res response
	res.Process_Type = "list_account_requests"
	if !info.is_admin {
		res.Status = Unauthorized
		res.Message = "You need to be logged in to have access to this functionality"
		out, _ := json.Marshal(res)
		return out
	}

	res.Status = Success
	list, _ := json.Marshal(User_Handler.Loaded_Requests)
	res.Message = string(list)
	out, _ := json.Marshal(res)
	return out
}

func accept_account_request(request *request_format, info *user_info) []byte {
	var res response
	res.Process_Type = "accept_account_request"
	if !info.is_admin {
		res.Status = Unauthorized
		res.Message = "You need to be logged in to have access to this functionality"
		out, _ := json.Marshal(res)
		return out
	}
	if len(request.Args) != 3 {
		res.Status = Fail
		res.Message = "You need min 3 arguments: username, is_admin, admin_level"
		out, _ := json.Marshal(res)
		return out
	}
	is_admin := false
	if request.Args[1] == "true" {
		is_admin = true
	}
	admin_level, err := strconv.Atoi(request.Args[2])
	if err != nil {
		admin_level = 5
	}
	User_Handler.Accept_account_request(request.Args[0], is_admin, uint8(admin_level))
	res.Status = Success
	res.Message = "User request has been accepted"
	out, _ := json.Marshal(res)
	return out
}

func run_script(request *request_format, info *user_info) []byte {
	var res response
	res.Process_Type = "run_script"
	path := "scripts/public/"
	if info.is_admin && len(request.Args) == 2 && request.Args[1] != "public" {
		path = "scripts/private/"
	} else if len(request.Args) < 1 {
		res.Status = Unauthorized
		res.Message = "You need 1 argument: script_path"
		out, _ := json.Marshal(res)
		return out
	}
	type script_result struct {
		Results string `json:"results"`
		Errors  error  `json:"errors"`
	}
	var scr script_result
	scr.Results, scr.Errors = Internal_Process_Handler.RunScript([]string{path + request.Args[0]})
	script_out_marsh, _ := json.Marshal(scr)
	res.Message = string(script_out_marsh)
	out, _ := json.Marshal(res)
	return out
}

func list_scripts(request *request_format, info *user_info) []byte {
	var res response
	res.Process_Type = "list_scripts"
	if info.username == "" {
		res.Status = Fail
		res.Message = "You need to be logged in"
		out, _ := json.Marshal(res)
		return out
	}
	type total_scripts struct {
		Public_scripts  []string `json:"scripts"`
		Private_scripts []string `json:"private,omitempty"`
	}
	var all_scripts total_scripts
	all_scripts.Public_scripts = list_public_scripts()
	if info.is_admin {
		all_scripts.Private_scripts = list_private_scripts()
	}
	encoded, _ := json.Marshal(all_scripts)
	res.Message = string(encoded)
	res.Status = Success
	out, _ := json.Marshal(res)
	return out
}

func upload_script(request *request_format, info *user_info) []byte {
	var res response
	res.Process_Type = "upload_script"
	if info.username == "" {
		res.Status = Fail
		res.Message = "You need to be logged in"
		out, _ := json.Marshal(res)
		return out
	}
	if len(request.Args) != 3 {
		res.Status = Fail
		res.Message = "You need 3 arguments: is_public(default false, in case you misspell), name_of_the_script, script_content"
		out, _ := json.Marshal(res)
		return out
	}
	is_public := "private"
	if request.Args[0] == "true" {
		is_public = "public"
	}
	result_path := "scripts/" + is_public + "/" + request.Args[1]
	exist, err := exists(result_path)
	if exist || err != nil {
		res.Status = Fail
		res.Message = "Script already exists under this name"
		out, _ := json.Marshal(res)
		return out
	}
	err = os.WriteFile(result_path, []byte(request.Args[2]), 0700)
	if err != nil {
		res.Status = Fail
		res.Message = "Unable to upload script"
		out, _ := json.Marshal(res)
		return out
	}

	res.Status = Success
	res.Message = "Script uploaded successfully"
	out, _ := json.Marshal(res)
	return out
}

func upload_user_file(request *request_format, info *user_info) []byte {
	var res response
	res.Process_Type = "upload_user_file"
	if info.username == "" {
		res.Status = Fail
		res.Message = "You need to be logged in"
		out, _ := json.Marshal(res)
		return out
	}
	if len(request.Args) != 2 {
		res.Status = Fail
		res.Message = "You need 2 arguments for this: path, file_content"
		out, _ := json.Marshal(res)
		return out
	}
	err := os.WriteFile("users_data/"+info.username+"/"+request.Args[0], []byte(request.Args[1]), 0700)
	if err != nil {
		res.Status = Fail
		res.Message = "Unable to upload file"
		out, _ := json.Marshal(res)
		return out
	}

	res.Status = Success
	res.Message = "File uploaded successfully"
	out, _ := json.Marshal(res)
	return out
}

func create_user_folder(request *request_format, info *user_info) []byte {
	var res response
	res.Process_Type = "create_user_folder"
	if info.username == "" {
		res.Status = Fail
		res.Message = "You need to be logged in"
		out, _ := json.Marshal(res)
		return out
	}
	err := os.MkdirAll("users_data/"+info.username+"/"+request.Args[0], 0700)
	if err != nil {
		res.Status = Fail
		res.Message = "Unable to create folder"
		out, _ := json.Marshal(res)
		return out
	}

	res.Status = Success
	res.Message = "Folder created successfully"
	out, _ := json.Marshal(res)
	return out
}

func list_user_folder(request *request_format, info *user_info) []byte {
	var res response
	res.Process_Type = "list_user_folder"
	if info.username == "" {
		res.Status = Fail
		res.Message = "You need to be logged in"
		out, _ := json.Marshal(res)
		return out
	}
	path := "users_data/" + info.username
	exist, err := exists(path)
	if !exist || err != nil {
		os.MkdirAll(path, 0700)
	}
	if len(request.Args) != 1 {
		res.Status = Fail
		res.Message = "You need 1 argument: path"
		out, _ := json.Marshal(res)
		return out
	}
	path += request.Args[0]
	exist, err = exists(path)
	if !exist || err != nil {
		res.Status = Fail
		res.Message = "Unkown path"
		out, _ := json.Marshal(res)
		return out
	}
	entries, err := os.ReadDir("users_data/" + info.username + "/")
	if err != nil {
		res.Status = Fail
		res.Message = "An error ocluded while trying to read folder content"
		out, _ := json.Marshal(res)
		return out
	}
	type item struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	var results []item

	for _, e := range entries {
		results = append(results, item{e.Name(), e.Type().String()})
	}

	out, err := json.Marshal(results)
	if err != nil {
		println("First error:", err.Error())
	}
	res.Message = string(out)
	out, err = json.Marshal(res)
	if err != nil {
		println("Second error:", err.Error())
	}
	return out
}

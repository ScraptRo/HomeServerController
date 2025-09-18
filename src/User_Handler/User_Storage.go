package User_Handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"time"
)

type User struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Salt        string `json:"salt"`
	CreatedAt   string `json:"created_at"`
	LastLogin   string `json:"last_login"`
	Admin       bool   `json:"admin"`
	Admin_Grade uint8  `json:"admin_grade"`
}

var LoadedUsers map[string]User

func generateRandomPassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, 12)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}
	return string(bytes)
}

// Generate salt for password hashing
func generateSalt() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Hash password with salt
func hashPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

func Load_users() {
	file, err := os.ReadFile("res/config_files/users.json")
	if err != nil {
		err = os.MkdirAll("res/config_files", os.ModePerm)
		if err != nil {
			println("Could not create config_files directory: " + err.Error())
		}
		LoadedUsers = nil
	} else {
		// Parse JSON and load users
		_ = json.Unmarshal(file, &LoadedUsers)
	}

	if LoadedUsers == nil {
		// Inform user about the creation of a new user map with default admin
		println("========================================")
		println("=          IMPORTANT NOTICE           =")
		println("========================================")
		println("=\t* No users found, new user map along side with the default admin")
		println("=\t* User credentials will be saved as admin_credentials.txt")
		println("=\t* If the users.json file exists, make sure it's not corrupted!")
		println("=\t* For security reasons, please change the default admin password after your first login!")
		println("=\t* And make sure that the users.json file doesn't get deleted. For now I didn't find a way to protect the credentials in a better way!")
		println("========================================\n")
		LoadedUsers = make(map[string]User)
		password := generateRandomPassword()
		salt := generateSalt()
		LoadedUsers["admin"] = User{
			Username:    "admin",
			Password:    hashPassword(password, salt),
			Salt:        salt,
			CreatedAt:   time.Now().Format(time.RFC3339),
			LastLogin:   time.Time{}.Format(time.RFC3339),
			Admin:       true,
			Admin_Grade: 0,
		}
		println("========================================")
		println("=          ADMIN CREDENTIALS          =")
		println("= Username: admin                     =")
		println("= Password:", password)
		println("========================================")
		os.WriteFile("res/config_files/admin_credentials.txt", []byte("Username: admin\nTemporary Password: "+password), 0644)
		Save_users()
	}
}
func Save_users() {
	data, err := json.MarshalIndent(LoadedUsers, "", "  ")
	if err != nil {
		println("Could not marshal users data: " + err.Error())
		return
	}
	err = os.WriteFile("res/config_files/users.json", data, 0644)
	if err != nil {
		println("Could not write users data to file: " + err.Error())
		return
	}
}
func Add_user(username, password string, admin bool, admin_grade uint8) bool {
	if User_exists(username) {
		return false
	}
	salt := generateSalt()
	LoadedUsers[username] = User{
		Username:    username,
		Password:    hashPassword(password, salt),
		Salt:        salt,
		CreatedAt:   time.Now().Format(time.RFC3339),
		LastLogin:   time.Time{}.Format(time.RFC3339),
		Admin:       admin,
		Admin_Grade: admin_grade,
	}
	Save_users()
	return true
}
func Remove_user(username string) {
	delete(LoadedUsers, username)
	Save_users()
}
func Authenticate_user(username, password string) bool {
	if !User_exists(username) {
		return false
	}
	hashedPassword := hashPassword(password, LoadedUsers[username].Salt)
	return hashedPassword == LoadedUsers[username].Password
}
func List_users() []string {
	var usernames []string
	for username := range LoadedUsers {
		usernames = append(usernames, username)
	}
	return usernames
}
func Change_password(username, newPassword string) {
	user, exists := LoadedUsers[username]
	if !exists {
		return
	}
	user.Password = hashPassword(newPassword, user.Salt)
	LoadedUsers[username] = user
	Save_users()
}
func User_exists(username string) bool {
	_, exists := LoadedUsers[username]
	return exists
}

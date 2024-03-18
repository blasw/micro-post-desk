package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type tokens struct {
	User_Id       uint   `json:"user_id"`
	Username      string `json:"username"`
	Access_Token  string `json:"access_token"`
	Refresh_Token string `json:"refresh_token"`
}

type UserInfo struct {
	User_Id  uint
	Username string
}

// TODO Needs refactoring
func ValidateUser(c *gin.Context) (*UserInfo, bool) {

	access_token, err := c.Cookie("access_token")
	if err != nil {
		return nil, false
	}
	refresh_token, err := c.Cookie("refresh_token")
	if err != nil {
		return nil, false
	}

	targetUrl := fmt.Sprintf("http://%v/users/auth", os.Getenv("USERS_LOADBALANCER"))

	payload := &tokens{
		Access_Token:  access_token,
		Refresh_Token: refresh_token,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		//Debug only
		fmt.Println("Unable to marshal json")
		return nil, false
	}

	resp, err := http.Post(targetUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		//Debug only
		fmt.Println("Unable to retrieve response from Users loadbalancer")
		return nil, false
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false
	}

	// if user is authorized we will get http.StatusOK either without body (if tokens are valid) or with new tokens (if access token is outdated). If we have a body we should set new cookies

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// Debug only
		fmt.Println("Unable to read response's body")
		return nil, true
	}
	var tokens *tokens
	err = json.Unmarshal(body, &tokens)
	if err != nil {
		// Debug only
		fmt.Println("Unable to unmarshal json")
		return nil, true
	}
	if tokens.Access_Token != "" {
		c.SetCookie("access_token", tokens.Access_Token, 3600*24, "/", "localhost", false, true)
	}
	if tokens.Refresh_Token != "" {
		c.SetCookie("refresh_token", tokens.Refresh_Token, 3600*24*7, "/", "localhost", false, true)
	}

	fmt.Println("-------------------------" + strconv.FormatUint(uint64(tokens.User_Id), 10) + " " + tokens.Username)

	return &UserInfo{User_Id: tokens.User_Id, Username: tokens.Username}, true
}

package main

import (
	"fmt"
	"github.com/xwi/go_uuwiseApi"
)

func main() {
	u := UuApiClient.New("软件ID", "软件KEY", "userName", "userPassword")
	u.AfterPropertiesSet()
	err := u.UserLogin()
	if err == nil {
		fmt.Println("登录成功")
		result := u.AutoRecognition("E:\\genimage.jpg", "3004")
		fmt.Println(result)
	}
}

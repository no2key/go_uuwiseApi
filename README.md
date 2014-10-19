GO版本优优云调用Api
========
## 安装

	go get github.com/Xwi/go_uuwiseApi

或

	gopm get github.com/Xwi/go_uuwiseApi


## API 文档

[Go Walker](http://gowalker.org/github.com/Xwi/go_uuwiseApi).

## 示例

请查看 [main.go](examples/main.go) 文件作为使用示例。

### 用例
```go
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
```
## 授权许可

本项目采用 Apache v2 开源授权许可证，完整的授权说明已放置在 [LICENSE](LICENSE) 文件中。

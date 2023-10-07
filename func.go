package main

//此文件用于预定义辅助功能函数
import (
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"time"
)

// 函数功能：从文件中读取数据
func loadTodosFromFile() ([]TODO, error) {
	data, err := ioutil.ReadFile(todosFile)
	if err != nil {
		return nil, err
	}

	var todos []TODO
	err = json.Unmarshal(data, &todos)
	if err != nil {
		return nil, err
	}

	return todos, nil
}

// 函数功能：将数据保存至文件中
func saveTodosToFile(todos []TODO) error {
	data, err := json.Marshal(todos)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(todosFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func loadUsersFromFile() ([]USER, error) {
	data, err := ioutil.ReadFile(usersFile)
	if err != nil {
		return nil, err
	}

	var users []USER
	err = json.Unmarshal(data, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func saveUsersToFile(users []USER) error {
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(usersFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// 辅助函数：检查索引是否在切片中
func containsIndex(index int, indexList []int) bool {
	for _, i := range indexList {
		if i == index {
			return true
		}
	}
	return false
}

func generateJWTToken(username string) (*jwt.Token, error) {
	// 设置令牌的有效期为一小时
	expirationTime := time.Now().Add(1 * time.Hour)

	// 创建JWT声明
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Subject:   username,
	}

	// 使用正确的密钥
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey) // 使用正确的密钥来签名令牌
	if err != nil {
		return nil, err
	}

	// 将签名后的令牌字符串设置到令牌中
	token.Raw = tokenString

	return token, nil
}

func authenticate(c *gin.Context) { //鉴权中间件
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(401, gin.H{"status": "未提供令牌"})
		c.Abort()
		return
	}

	// 解析JWT令牌
	claims, err := parseJWTToken(tokenString)
	if err != nil {
		c.JSON(401, gin.H{"status": "令牌无效"})
		c.Abort()
		return
	}

	// 从令牌中获取用户名
	username := claims.Subject
	if username == "" {
		c.JSON(401, gin.H{"status": "令牌中缺少用户名"})
		c.Abort()
		return
	}

	// 将用户名添加到请求上下文中，以便后续处理函数可以访问它
	c.Set("user", username)

	// 继续处理请求
	c.Next()
}

func parseJWTToken(tokenString string) (*jwt.StandardClaims, error) { //解析jwt令牌
	claims := &jwt.StandardClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 使用secretKey 作为密钥来解析令牌
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌是否有效
	if !token.Valid {
		return nil, errors.New("无效的令牌")
	}

	return claims, nil
}

func addTokenToHeader(c *gin.Context, token *jwt.Token) {
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(500, ErrGenerateToken)
		c.Abort()
		return
	}

	c.Header("Authorization", "Bearer "+tokenString)
}

func tokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息，这里可以根据你的需求获取
		user, _ := c.Get("user")
		username, ok := user.(string)
		if !ok || username == "" {
			c.JSON(401, gin.H{"status": "用户未登录或无效的用户"})
			c.Abort()
			return
		}

		// 在这里生成令牌并将其放入请求头
		token, err := generateJWTToken(username)
		if err != nil {
			c.JSON(500, ErrGenerateToken)
			c.Abort()
			return
		}

		addTokenToHeader(c, token)

		// 继续处理请求
		c.Next()
	}
}

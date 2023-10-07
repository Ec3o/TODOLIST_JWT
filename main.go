package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"sort"
	"strconv"
	"time"
)

type TODO struct {
	Username string    `json:"username"`
	Index    int       `json:"index"`
	Content  string    `json:"content"`
	Done     bool      `json:"done"`
	Deadline time.Time `json:"deadline"`
}

type USER struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var todosFile = "todos.json"
var usersFile = "users.json"
var deletedTodoIndexes []int
var jwtKey = []byte("Ec3o")

func main() {
	r := gin.Default()

	// 使用中间件来对需要身份验证的路由进行保护
	authGroup := r.Group("/")
	authGroup.Use(tokenMiddleware())
	authGroup.Use(authenticate)
	authGroup.POST("/todo", TodoCreation)          //增
	authGroup.DELETE("/todo/:index", TodoDeletion) //删(不改动序号)
	authGroup.PUT("/todo/:index", TodoUpdate)      //改
	authGroup.GET("/todo", ListTodos)              //查(使用条件筛选)
	authGroup.GET("/todo/:index", GetTodo)         //获取单个todo信息

	r.POST("/register", useregister)
	r.POST("/login", userlogin)
	r.Run(":8100")
}

func TodoCreation(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	fmt.Println("Authorization Header:", tokenString)
	if tokenString == "" {
		c.JSON(401, ErrNoToken)
		return
	}

	// 解析 JWT 令牌
	claims, _ := parseJWTToken(tokenString)
	if claims == nil {
		c.JSON(401, ErrInvalidToken)
		return
	}

	var todo TODO
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(400, ErrInvalidTODOFormat)
		return
	}

	if todo.Deadline.IsZero() {
		defaultDeadline := time.Now().Add(time.Hour * 24 * 7)
		todo.Deadline = defaultDeadline
	} else {
		parsedDeadline, err := time.Parse(time.RFC3339, todo.Deadline.String())
		if err != nil || parsedDeadline.Before(time.Now()) {
			c.JSON(400, ErrInvalidDeadline)
			return
		}
	}

	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrReadTODOData)
		return
	}

	// 计算用户的索引
	userIndex := 1
	for _, t := range existingTodos {
		if t.Username == claims.Subject {
			userIndex++
		}
	}

	// 为新的 Todo 分配 index 序号
	todo.Index = userIndex
	todo.Username = claims.Subject // 添加用户名字段

	existingTodos = append(existingTodos, todo)

	err = saveTodosToFile(existingTodos)
	if err != nil {
		c.JSON(500, ErrSaveTODOData)
		return
	}

	c.JSON(200, gin.H{"status": "数据提交成功"})
}

func TodoDeletion(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(401, ErrNoToken)
		return
	}

	// 解析 JWT 令牌以获取当前用户
	claims, _ := parseJWTToken(tokenString)
	if claims == nil {
		c.JSON(401, ErrInvalidToken)
		return
	}

	// 获取当前用户
	currentUser := claims.Subject

	indexToDelete, err := strconv.Atoi(c.Param("index"))
	if err != nil || indexToDelete < 0 {
		c.JSON(404, ErrTODOIndexNotExist)
		return
	}

	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrTODONotFound)
		return
	}

	// 遍历待办事项列表，找到与当前用户匹配的待办事项并匹配索引
	for index, todo := range existingTodos {
		if todo.Username == currentUser && index == indexToDelete {
			// 标记待办事项为已删除
			todo.Content = "此Todo已被删除"
			todo.Done = true

			// 更新待办事项回到列表
			existingTodos[index] = todo

			err = saveTodosToFile(existingTodos)
			if err != nil {
				c.JSON(500, ErrSaveTODOData)
				return
			}

			c.JSON(200, gin.H{"status": "删除成功", "被删除的数据是": todo})
			return
		}
	}

	// 如果没有匹配的待办事项，返回错误
	c.JSON(404, ErrTODOIndexNotExist)
}

func TodoUpdate(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(401, ErrNoToken)
		return
	}

	// 解析 JWT 令牌以获取当前用户
	claims, _ := parseJWTToken(tokenString)
	if claims == nil {
		c.JSON(401, ErrInvalidToken)
		return
	}

	// 获取当前用户
	currentUser := claims.Subject

	indexToUpdate, err := strconv.Atoi(c.Param("index"))
	if err != nil || indexToUpdate < 0 {
		c.JSON(404, ErrTODOIndexNotExist)
		return
	}

	var todo TODO
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(400, ErrInvalidTODOFormat)
		return
	}

	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrReadTODOData)
		return
	}

	// 遍历待办事项列表，找到与当前用户匹配的待办事项并匹配索引
	for index, existingTodo := range existingTodos {
		if existingTodo.Username == currentUser && index == indexToUpdate {
			// 更新待办事项内容
			existingTodo.Content = todo.Content
			existingTodo.Done = todo.Done
			existingTodo.Deadline = todo.Deadline

			// 更新待办事项回到列表
			existingTodos[index] = existingTodo

			err = saveTodosToFile(existingTodos)
			if err != nil {
				c.JSON(500, ErrSaveTODOData)
				return
			}

			c.JSON(200, gin.H{"status": "修改成功"})
			return
		}
	}

	// 如果没有匹配的待办事项，返回错误
	c.JSON(404, ErrTODOIndexNotExist)
}

func ListTodos(c *gin.Context) {
	// 从请求上下文中获取当前用户的用户名
	user, _ := c.Get("user")
	username, ok := user.(string)
	if !ok || username == "" {
		c.JSON(401, gin.H{"status": "用户未登录或无效的用户"})
		return
	}

	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrReadTODOData)
		return
	}

	// 获取查询参数
	deadline := c.DefaultQuery("deadline", "")
	reverse := c.DefaultQuery("reverse", "false")
	finished := c.DefaultQuery("finished", "")

	// 转换 reverse 字符串为布尔值
	reverseSort := (reverse == "true")

	// 根据 finished 参数过滤待办事项
	filteredTodos := []TODO{}
	for index, todo := range existingTodos {
		// 检查索引是否在 deletedTodoIndexes 中，如果在就跳过
		if containsIndex(index, deletedTodoIndexes) {
			continue
		}

		// 只返回属于当前用户的待办事项
		if todo.Username != username {
			continue
		}

		if (finished == "true" && todo.Done) || (finished == "false" && !todo.Done) || finished == "" {
			if deadline == "" || (deadline != "" && todo.Deadline.String() <= deadline) {
				filteredTodos = append(filteredTodos, todo)
			}
		}
	}

	// 根据 reverseSort 参数排序
	if reverseSort {
		sort.Slice(filteredTodos, func(i, j int) bool {
			return filteredTodos[i].Deadline.After(filteredTodos[j].Deadline)
		})
	} else {
		sort.Slice(filteredTodos, func(i, j int) bool {
			return filteredTodos[i].Deadline.Before(filteredTodos[j].Deadline)
		})
	}

	// 返回结果
	todosWithIndex := []map[string]interface{}{}

	for index, todo := range filteredTodos {
		todoWithIndex := map[string]interface{}{
			"index":    index,
			"content":  todo.Content,
			"done":     todo.Done,
			"deadline": todo.Deadline.String(), // 将时间转换为字符串
		}
		todosWithIndex = append(todosWithIndex, todoWithIndex)
	}

	c.JSON(200, todosWithIndex)
}

func GetTodo(c *gin.Context) {
	// 从请求上下文中获取当前用户的用户名
	user, _ := c.Get("user")
	username, ok := user.(string)
	if !ok || username == "" {
		c.JSON(401, gin.H{"status": "用户未登录或无效的用户"})
		return
	}

	index, err := strconv.Atoi(c.Param("index"))
	if err != nil || index < 0 {
		c.JSON(404, ErrTODOIndexNotExist)
		return
	}

	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrReadTODOData)
		return
	}

	if index >= len(existingTodos) {
		c.JSON(404, ErrTODOIndexNotExist)
		return
	}

	// 检查当前用户是否有权限获取该待办事项
	if existingTodos[index].Username != username {
		c.JSON(403, gin.H{"status": "无权限获取该待办事项"})
		return
	}

	todoWithIndex := map[string]interface{}{
		"index":    index,
		"content":  existingTodos[index].Content,
		"done":     existingTodos[index].Done,
		"deadline": existingTodos[index].Deadline.String(), // 将时间转换为字符串
	}

	c.JSON(200, todoWithIndex)
}

func useregister(c *gin.Context) {
	var user USER
	if err := c.BindJSON(&user); err != nil {
		c.JSON(400, ErrInvalidUSERFormat)
		return
	}

	if len(user.Password) <= 6 { //密码长度过短提示重新设置
		c.JSON(400, ErrInvalidPassword)
		return
	}

	existingUsers, err := loadUsersFromFile()
	if err != nil {
		c.JSON(500, ErrReadUserData)
		return
	}

	// 检查是否已经存在相同的用户名
	for _, existingUser := range existingUsers {
		if existingUser.Username == user.Username {
			c.JSON(400, gin.H{"status": "用户名已经被注册"})
			return
		}
	}

	existingUsers = append(existingUsers, user)
	err = saveUsersToFile(existingUsers)
	if err != nil {
		c.JSON(500, ErrSaveUserData)
		return
	}

	c.JSON(200, gin.H{"status": "用户注册成功"})
}

func userlogin(c *gin.Context) {
	var user USER
	if err := c.BindJSON(&user); err != nil {
		c.JSON(400, ErrInvalidUSERFormat)
		return
	}

	existingUsers, err := loadUsersFromFile()
	if err != nil {
		c.JSON(500, ErrReadUserData)
		return
	}

	var foundUser USER
	for _, existingUser := range existingUsers {
		if existingUser.Username == user.Username && existingUser.Password == user.Password {
			foundUser = existingUser
			break
		}
	}

	if foundUser.Username != "" {
		// 将用户名信息添加到上下文中
		c.Set("user", foundUser.Username)

		// 生成 JWT 令牌
		token, err := generateJWTToken(foundUser.Username)
		if err != nil {
			c.JSON(500, ErrGenerateToken)
			return
		}

		// 将令牌放入响应的 Header 中
		addTokenToHeader(c, token)

		c.JSON(200, gin.H{"status": "用户登录成功", "token": token})
	} else {
		c.JSON(404, ErrUserlogin)
	}
}

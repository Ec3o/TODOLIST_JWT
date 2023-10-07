## 后端任务说明文档的第二版.

## Created by Ec3o.

我对学长指出的一些项目问题作出了一些更新,问题如下:

1. 时间应该可以定义为 `time.Time` 或者 `int64` 类型，一般前端传的都是 Unix 时间戳
2. 可以考虑将匿名函数抽离出来
3. 可以在一个文件或者包中将所有错误提前定义好
4. 高并发时存在问题，如果很多请求需要同时读取/写入怎么办？（这是一个比较进阶的问题，没法解决也没问题）
   可以看一看 `sync.Mutex` 的用法
5. 可以考虑一下完成我的进阶需求
   我说的是删除前面的 TODO 的时候，后面的 TODO 的编号不会改变
   为什么会有这种需求，你想想看，如果以后能给 TODO 关联一些东西（比如说要实现 A 是 B 的子 TODO），这时就要在 B 中存下 A 的编号，并且编号不能改变）

### 抽离匿名函数

我抽离了部分函数内容,以让程序更具有可读性.

主函数被简化如下:

```go
func main() {
	r := gin.Default()

	r.POST("/todo", TodoCreation)
	r.DELETE("/todo/:index", TodoDeletion)
	r.PUT("/todo/:index", TodoUpdate)
	r.GET("/todo", ListTodos)
	r.GET("/todo/:index", GetTodo)
	r.GET("/unfinished_todo", ListUnfinishedTodos)
	r.Run(":8100")
}
```

对应函数功能定义如下:

```go
func TodoCreation(c *gin.Context) {
    var todo TODO
    if err := c.BindJSON(&todo); err != nil {
       c.JSON(400, ErrInvalidTODOFormat)
       return
    }

    if todo.Deadline == "" {
       defaultDeadline := time.Now().Add(time.Hour * 24 * 7)
       todo.Deadline = defaultDeadline.Format("2006-01-02")
    } else {
       parsedDeadline, err := time.Parse("2006-01-02", todo.Deadline)
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

    existingTodos = append(existingTodos, todo)

    err = saveTodosToFile(existingTodos)
    if err != nil {
       c.JSON(500, ErrSaveTODOData)
       return
    }

    c.JSON(200, gin.H{"status": "数据提交成功"})
}
```

与之类似的多种匿名函数已被提取出来并设置成为新函数,便于维护与修改.

### 定义错误文件

错误文件已被定义,存入文件`error.go`文件中,并存储于main.go同一目录下.

```go
package main
// TODOError 此文件用于预定义错误类型
type TODOError struct {
	Message string `json:"error"`
}

var (
	ErrInvalidTODOFormat = TODOError{"抱歉，您提供的TODO数据格式不正确"}
	ErrInvalidDeadline   = TODOError{"无效的截止时间，格式应为 '2006-01-02' 并且不能是过去的时间"}
	ErrReadTODOData      = TODOError{"无法读取TODO数据"}
	ErrSaveTODOData      = TODOError{"无法保存TODO数据"}
	ErrTODOIndexNotExist = TODOError{"抱歉，您访问的ToDo目前不存在，请先创建"}
	ErrTODONotFound      = TODOError{"抱歉，您要删除的ToDo目前不存在，请先创建"}
	ErrNoTodayTODO       = TODOError{"今天没有截止的TODO"}
	ErrNoWeekTODO        = TODOError{"本周没有截止的TODO"}
	ErrNoUnfinishedTODO  = TODOError{"没有未完成的TODO"}
)

```

### Todo编号不变问题

目前没想到什么好的解决方式,我选择修改deletetodo函数解决问题.基本思路是把被删除的todo置换成空todo并加入索引,使其在筛选中不会被显示.

```go
var deletedTodoIndexes []int
```

定义一个被删除的todo的索引切片.

```go
func TodoDeletion(c *gin.Context) {
    index, err := strconv.Atoi(c.Param("index"))
    if err != nil || index < 0 {
       c.JSON(404, ErrTODOIndexNotExist)
       return
    }

    existingTodos, err := loadTodosFromFile()
    if err != nil {
       c.JSON(500, ErrTODONotFound)
       return
    }

    if index >= len(existingTodos) {
       c.JSON(404, ErrTODOIndexNotExist)
       return
    }

    // 将索引添加到已删除的待办事项列表中
    deletedTodoIndexes = append(deletedTodoIndexes, index)

    // 更新已删除的待办事项回到列表
    existingTodos[index].Content = "此Todo已被删除"
    existingTodos[index].Done = true

    err = saveTodosToFile(existingTodos)
    if err != nil {
       c.JSON(500, ErrSaveTODOData)
       return
    }

    c.JSON(200, gin.H{"status": "删除成功", "被删除的数据是": existingTodos[index]})
}
```

修改大概就是这样.

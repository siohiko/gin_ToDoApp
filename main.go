package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/gin-gonic/gin"
	"net/http"
	"gopkg.in/go-playground/validator.v8"
	"golang.org/x/crypto/bcrypt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"unicode/utf8"
)

const (
	MinCost     int = 4
	MaxCost     int = 31
	DefaultCost int = 10
)


type User struct {
	gorm.Model
	UserId string `gorm:"unique" validate:"required"`
	Name string `validate:"required"`
	Password []byte `validate:"required"`
}


func main() {

	db := gormConnect()
	defer db.Close()
	db.AutoMigrate(&User{})

	router := gin.Default()
	router.Static("styles", "./styles")
	router.LoadHTMLGlob("templates/*")
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("todoapp", store))

	//	v1 route
	v1 := router.Group("/v1")
	{
		v1.GET("/top", topPageEndPoint)
		v1.GET("/create_account_page", createAccountPageEndPoint)
		v1.POST("/register", registerEndPoint)
		v1.POST("/login", loginEndPoint)
		v1.GET("/mypage", sessionCheck(), mypageEndPoint)
		v1.POST("/logout", logoutEndPoint)
	}
	router.Run(":8080")
}


//*****************************//
//**********EndPoints**********//
//*****************************//

func topPageEndPoint(c *gin.Context) {
	c.HTML(http.StatusOK, "top.tmpl", gin.H{
		"title": "Top Page",
	})
}



func createAccountPageEndPoint(c *gin.Context) {
	c.HTML(http.StatusOK, "createAccount.tmpl", gin.H{})
}



func registerEndPoint(c *gin.Context) {
	//エラーメッセージ格納用スライス
	var errorMessages []string 

	config := &validator.Config{TagName: "validate"}
	validate := validator.New(config)

	user_id := c.PostForm("user_id")
	name := c.PostForm("name")
	password := c.PostForm("password")
	lengthPass := utf8.RuneCountInString(password)

	//送信されたパスワードの文字数バリデーション
	if lengthPass < 8 || 16 < lengthPass {
		var errorMessage string
		errorMessage = "Password must be at least 8 and no more than 16 characters long"
		errorMessages = append(errorMessages, errorMessage)
	}

	//パスワードハッシュ時にエラーが起きた場合の処理
	bs, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		var errorMessage string
		errorMessage = "Password must be at least 8 and no more than 16 characters long"
		errorMessages = append(errorMessages, errorMessage)
	}

	user := &User{
		UserId:     user_id,
		Name:       name,
		Password:   bs,
	}

	//ここからモデルレベルでのバリデーション
	errs := validate.Struct(user)

	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {

			fieldName := err.Field
			var errorMessage string
			var errorCause string
			var typ = err.Tag
			
			switch typ {
				case "required":
					errorCause = "required"
			}

			errorMessage = "error message for" + fieldName + "." + " because " + errorCause
			errorMessages = append(errorMessages, errorMessage)
		}
	}

	//エラーメッセージが一つもなければユーザー登録成功
	if len(errorMessages) > 0 { 
		c.HTML(http.StatusFound, "createAccount.tmpl", gin.H{
			"errorMessages": errorMessages,
		})
	} else {
		dbInsert(user)
		c.Redirect(http.StatusFound, "/v1/top")
	}
}



func loginEndPoint(c *gin.Context) {
	user_id := c.PostForm("user_id")
	password := c.PostForm("password")

	var user User

	findUserById(&user, string(user_id))

	err := bcrypt.CompareHashAndPassword(user.Password, []byte(password))

	if err != nil {
		c.HTML(http.StatusUnauthorized, "top.tmpl", gin.H{})
	} else {
		session := sessions.Default(c)
		session.Set("user_id", user.UserId)
    session.Save()
		c.Redirect(http.StatusFound, "/v1/mypage")
	}
}


func mypageEndPoint(c *gin.Context) {
	c.HTML(http.StatusOK, "mypage.tmpl", gin.H{
	})
}



func logoutEndPoint(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/v1/top")
}
//*****************************//
//*****************************//
//*****************************//




func gormConnect() *gorm.DB {
  DBMS     := "mysql"
  USER     := "todoapp"
  PASS     := "12345678"
  PROTOCOL := "tcp(localhost:3306)"
  DBNAME   := "todoapp"

  CONNECT := USER+":"+PASS+"@"+PROTOCOL+"/"+DBNAME+"?parseTime=true"
  db,err := gorm.Open(DBMS, CONNECT)

  if err != nil {
    panic(err.Error())
  }
  return db
}



func dbInsert(user *User ) {
	db := gormConnect()
	defer db.Close()
	result := db.Create(user)

	if result.Error != nil {
		panic(result.Error)
	}
}


func findUserById(user *User, user_id string ) {
	db := gormConnect()
	defer db.Close()
	db.Where("user_id = ?", user_id).First(&user)
}


func sessionCheck() gin.HandlerFunc {
	return func(c *gin.Context) {

			session := sessions.Default(c)
			userId := session.Get("user_id")

			if userId == nil {
					c.Redirect(http.StatusMovedPermanently, "/v1/top")
					c.Abort()
			} else {
					c.Next()
			}
	}
}
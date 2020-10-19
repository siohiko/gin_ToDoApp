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

	"errors"
	"github.com/go-sql-driver/mysql"
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

	router := setupRouter()
	router.Run(":8080")
}



func setupRouter() *gin.Engine {
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
	return router
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

	postedUser := postedUser(c)
	lengthPass := utf8.RuneCountInString(string(postedUser.Password))

	//送信されたパスワードの文字数バリデーション
	if lengthPass < 8 || 16 < lengthPass {
		errorMessages = append(errorMessages, "Password must be at least 8 and no more than 16 characters long")
	}

	//パスワードハッシュ時にエラーが起きた場合の処理
	bs, err := bcrypt.GenerateFromPassword(postedUser.Password, bcrypt.DefaultCost)
	if err != nil {
		errorMessages = append(errorMessages, "Password must be at least 8 and no more than 16 characters long")
	}

	postedUser.Password = bs

	//ここからモデルレベルでのバリデーション
	errs := validate.Struct(postedUser)

	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {

			fieldName := err.Field
			var errorCause string
			var typ = err.Tag
			
			switch typ {
				case "required":
					errorCause = "required"
			}

			errorMessages = append(errorMessages, "error message for" + fieldName + "." + " because " + errorCause)
		}
	}

	//エラーメッセージが一つもなければユーザー登録成功
	if len(errorMessages) > 0 { 
		c.HTML(http.StatusFound, "createAccount.tmpl", gin.H{
			"errorMessages": errorMessages,
		})
	} else {
		errMsg := dbInsert(postedUser)
		if errMsg != "" {
			errorMessages = append(errorMessages, errMsg)
			c.HTML(http.StatusFound, "createAccount.tmpl", gin.H{
				"errorMessages": errorMessages,
			})
		} else {
			c.Redirect(http.StatusSeeOther, "/v1/top")
		}
	}
}



func loginEndPoint(c *gin.Context) {
	postedUser := postedUser(c)
	
	var user User

	if err := findUserBy(&user, "user_id", string(postedUser.UserId)); err != nil {
		c.HTML(http.StatusUnauthorized, "top.tmpl", gin.H{})
		return
	}


	err := bcrypt.CompareHashAndPassword(user.Password, postedUser.Password)
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



func dbInsert(user *User ) (errMsg string){
	db := gormConnect()
	defer db.Close()
	result := db.Create(user)
	errMsg = ""

	if result.Error != nil {
		if driverErr, ok := result.Error.(*mysql.MySQLError); ok {
			errMsg = mySQLErrorMsgHandling(driverErr)
		}
	}

	return errMsg
}

func mySQLErrorMsgHandling(driverErr *mysql.MySQLError) string {
	var errMsg string 
	switch driverErr.Number {
	case 1062:
			errMsg = "That user ID has already been used"
	default:
		errMsg = ""
	}
	return errMsg
}

func findUserBy(user *User, columnName string, value string) error {
	db := gormConnect()
	defer db.Close()
	if err := db.Where(columnName + " = ?", value).First(&user).Error; err != nil {
		return errors.New("error")
	}
	return nil
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


func postedUser(c *gin.Context) *User {
	//ginのバインド関数は、送られてくるパスワードの型とUser構造体のパスワードの型が違うため、400を吐くから使わない方針。
	var postedUser User
	postedUser.UserId = c.PostForm("user_id")
	postedUser.Name = c.PostForm("name")
	//送られてきたパスワードはstring型なので、byte型に変換して入れ直す
	postedUser.Password = []byte(c.PostForm("password"))

	return &postedUser
}
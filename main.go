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
)

const (
	MinCost     int = 4
	MaxCost     int = 31
	DefaultCost int = 10
)

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
	}
	router.Run(":8080")
}




func topPageEndPoint(c *gin.Context) {
	session := sessions.Default(c)

	c.HTML(http.StatusOK, "top.tmpl", gin.H{
		"title": "Top Page",
	})
}



func createAccountPageEndPoint(c *gin.Context) {
	c.HTML(http.StatusOK, "createAccount.tmpl", gin.H{})
}



func registerEndPoint(c *gin.Context) {

	config := &validator.Config{TagName: "validate"}
	validate = validator.New(config)

	user_id := c.PostForm("user_id")
	name := c.PostForm("name")
	password := c.PostForm("password")

	bs, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		c.Redirect(http.StatusFound, "/v1/top")
		return
	}

	user := &User{
		UserId:     user_id,
		Name:       name,
		Password:   bs,
	}

	var errorMessages []string 
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

	db := gormConnect()
	defer db.Close()
	db.Where("user_id = ?", string(user_id)).First(&user)

	err := bcrypt.CompareHashAndPassword(user.Password, []byte(password))

	if err != nil {
		c.HTML(http.StatusUnauthorized, "top.tmpl", gin.H{})
	} else {
		session := sessions.Default(c)
		session.Set("user_id", user.UserId)
    session.Save()
		c.Redirect(http.StatusFound, "/v1/top")
	}
}


type User struct {
	gorm.Model
	UserId string `gorm:"unique" validate:"required"`
	Name string `validate:"required"`
	Password []byte `validate:"required"`
}

var validate *validator.Validate

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
package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/gin-gonic/gin"
	"net/http"
	"gopkg.in/go-playground/validator.v8"
)

func main() {

	db := gormConnect()
	defer db.Close()
	db.AutoMigrate(&User{})

	router := gin.Default()
	router.Static("styles", "./styles")
	router.LoadHTMLGlob("templates/*")

	//	v1 route
	v1 := router.Group("/v1")
	{
		v1.GET("/top", topPageEndPoint)
		v1.GET("/create_account_page", createAccountPageEndPoint)
		v1.POST("/register", registerEndPoint)
	}
	router.Run(":8080")
}




func topPageEndPoint(c *gin.Context) {
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

	user := &User{
		UserId:     user_id,
		Name:       name,
		Password:   password,
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



type User struct {
	gorm.Model
	UserId string `gorm:"unique" validate:"required"`
	Name string `validate:"required"`
	Password string `validate:"required"`
}

var validate *validator.Validate

func gormConnect() *gorm.DB {
  DBMS     := "mysql"
  USER     := "todoapp"
  PASS     := "12345678"
  PROTOCOL := "tcp(localhost:3306)"
  DBNAME   := "todoapp"

  CONNECT := USER+":"+PASS+"@"+PROTOCOL+"/"+DBNAME
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
package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"github.com/stretchr/testify/assert"
	"fmt"
)
var testRouter = setupRouter()


func TestMain(m *testing.M) {
	func() {
		fmt.Println("Prepare test")
	}()
	m.Run()
	func() {
		fmt.Println("Teardown test")
	}()
}

//登録成功するユーザーデータ
var validUser = &User{
	UserId: "valid_user",
	Name: "valid_user_name",
	Password: []byte("12345678"),
}

//パスワードの文字数制限に引っかかるユーザー
var incorrectUserForPass = &User{
	UserId: "incorrect_user_for_pass",
	Name: "incorrect_user_for_pass_name",
	Password: []byte("1234"),
}

//ユーザーIDが未入力の時に引っかかるユーザー
var incorrectUserForNill = &User{
	UserId: "",
	Name: "incorrect_user_for_nill",
	Password: []byte("12345678"),
}


func TestTopPageEndPoint(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/top", nil)
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCreateAccountPageEndPoint(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/create_account_page", nil)
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}


func TestRegisterValidUser(t *testing.T) {

	values := setMockUserToPostData(validUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST", 
		"/v1/register",
		strings.NewReader(values.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testRouter.ServeHTTP(w, req)

	//登録成功後、リダイレクトされているかテスト
	assert.Equal(t, 303, w.Code)

	//実際にDBに登録されているかテスト
	var user User
	err := findUserBy(&user, "user_id", validUser.UserId)
	assert.Empty(t, err)

	dbDeleteForTest(&user)
}


func TestRegisterInvalidPassword(t *testing.T) {

	values := setMockUserToPostData(incorrectUserForPass)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST", 
		"/v1/register",
		strings.NewReader(values.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testRouter.ServeHTTP(w, req)

	//登録失敗後、リダイレクトされているかテスト
	assert.Equal(t, 302, w.Code)

	//実際にDBに登録されているかテスト
	var user User
	err := findUserBy(&user, "user_id", incorrectUserForPass.UserId)
	assert.Error(t, err)
}


func TestRegisterInvalidUserId(t *testing.T) {

	values := setMockUserToPostData(incorrectUserForPass)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST", 
		"/v1/register",
		strings.NewReader(values.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testRouter.ServeHTTP(w, req)

	//登録失敗後、リダイレクトされているかテスト
	assert.Equal(t, 302, w.Code)

	//実際にDBに登録されていないかのテスト
	var user User
	err := findUserBy(&user, "user_id", incorrectUserForPass.UserId)
	assert.Error(t, err)
}


func TestRegisterDuplicateData(t *testing.T) {

	//先に正しいユーザーデータを登録しておく
	registerUser := *validUser
	bs, _ := bcrypt.GenerateFromPassword(validUser.Password, bcrypt.DefaultCost)
	registerUser.Password = bs
	if err := dbInsert(&registerUser); err != "" {
		panic("An error occurred when registering test data")
	}

	values := setMockUserToPostData(validUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST", 
		"/v1/register",
		strings.NewReader(values.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testRouter.ServeHTTP(w, req)

	//登録失敗後、リダイレクトされているかテスト
	assert.Equal(t, 302, w.Code)

	var user User
	findUserBy(&user, "user_id", validUser.UserId)
	dbDeleteForTest(&user)
}



func TestLoginWidthValidUser(t *testing.T) {

	//先に正しいユーザーデータを登録しておく
	registerUser := *validUser
	bs, _ := bcrypt.GenerateFromPassword(validUser.Password, bcrypt.DefaultCost)
	registerUser.Password = bs
	if err := dbInsert(&registerUser); err != "" {
		panic("An error occurred when registering test data")
	}

	values := setMockUserToPostData(validUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST", 
		"/v1/login",
		strings.NewReader(values.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testRouter.ServeHTTP(w, req)

	//セッションがちゃんとセットされているかどうか
	session := w.Header()["Set-Cookie"][0]
	assert.Contains(t, session, "todoapp")

	//ログイン成功後、リダイレクトされているかテスト
	assert.Equal(t, 302, w.Code)

	var user User
	findUserBy(&user, "user_id", validUser.UserId)
	dbDeleteForTest(&user)
}

func TestLoginWidthInValidUser(t *testing.T) {

	//先に正しいユーザーデータを登録しておく
	registerUser := *validUser
	bs, _ := bcrypt.GenerateFromPassword(validUser.Password, bcrypt.DefaultCost)
	registerUser.Password = bs
	if err := dbInsert(&registerUser); err != "" {
		panic("An error occurred when registering test data")
	}

	//登録成功するユーザーデータ(パスワードが違う)
	var invalidUser = &User{
		UserId: "valid_user",
		Name: "valid_user_name",
		Password: []byte("123456789"),
	}

	values := setMockUserToPostData(invalidUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST", 
		"/v1/login",
		strings.NewReader(values.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testRouter.ServeHTTP(w, req)

	//ログイン成功後、リダイレクトされているかテスト
	assert.Equal(t, 401, w.Code)

	var user User
	findUserBy(&user, "user_id", validUser.UserId)
	dbDeleteForTest(&user)
}

func TestLogout(t *testing.T) {

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST", 
		"/v1/logout",
		nil,
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testRouter.ServeHTTP(w, req)

	//セッションがちゃんとけされているかどうか
	session := w.Header()["Set-Cookie"]
	assert.Empty(t, session)

	assert.Equal(t, 302, w.Code)
}



func setMockUserToPostData(user *User) *url.Values {
	values := &url.Values{}
	values.Set("user_id", user.UserId)
	values.Add("name", user.Name)
	values.Add("password", string(user.Password))

	return values
}

func dbDeleteForTest(user *User) {
	db := gormConnect()
	db.Unscoped().Delete(&user)
}
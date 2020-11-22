package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

//DB global DB variable .
var DB *sql.DB

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "54321"
	dbname   = "postgres"
)

func main() {
	e := echo.New()
	e.POST("/login", LoginUser)
	e.POST("/signup", SignUpUser)
	e.PUT("/updatepost", UpdatePostCount)
	initDB()
	e.Logger.Fatal(e.Start(":8000"))
	defer DB.Close()
}
func initDB() {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error in connecting db")
		panic(err)
	}
}

//UserCredential for user login
type UserCredential struct {
	Email    string `jsonn:"email", db:"email"`
	Username string `json:"username", db:"username"`
	Password string `json:"password", db:"password"`
}

//LoginUser to allow login
func LoginUser(c echo.Context) error {
	creds := &UserCredential{}
	err := json.NewDecoder(c.Request().Body).Decode(creds)
	defer c.Request().Body.Close()
	if err != nil {
		log.Fatalf("Failed reading request body %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error)
	}
	result := DB.QueryRow("select password from user where name=$1 or email = $2", creds.Username, creds.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error)
	}
	storedCreds := &UserCredential{}
	err = result.Scan(&storedCreds.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error)
	}
	if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {

		return echo.NewHTTPError(http.StatusUnauthorized, err.Error)
	}
	return c.JSON(http.StatusOK, "Successfully Logged In ...")
}

//SignUpUser to allow signing in
func SignUpUser(c echo.Context) error {
	creds := &UserCredential{}
	err := json.NewDecoder(c.Request().Body).Decode(creds)
	defer c.Request().Body.Close()
	if err != nil {
		log.Fatalf("Failed reading request body %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	DB.Exec("insert into user (name,email,password) values ($1, $2, $3)", creds.Username, creds.Email, string(hashedPassword))
	return c.JSON(http.StatusOK, "Successfully Signed In ...")
}

//UpdatePostCount to update user posts
func UpdatePostCount(c echo.Context) error {
	incrementflag, _ := strconv.ParseBool(c.FormValue("IncPostCount"))
	userID := c.FormValue("userID")
	result := DB.QueryRow("select post_count from user where id=$1", userID)
	var postcount int
	err := result.Scan(&postcount)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error)
	}
	if incrementflag {
		postcount++
	} else {
		postcount--
	}
	DB.Exec("update user set post_count = $1 where id=$2", postcount, userID)
	return c.JSON(http.StatusOK, "PostCount Updated Successfully ...")
}

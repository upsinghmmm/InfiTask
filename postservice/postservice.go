package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "54321"
	dbname   = "postgres"
	authkey  = "authkey"
)

//DB global DB variable .
var DB *sql.DB

//User reference type for UserService
type User struct {
	UserID       int
	IncPostCount bool
}

//PostModel to update postTable
type PostModel struct {
	ID          int    `json:"id"`
	UserID      int    `json:"UserId"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

func main() {
	e := echo.New()
	//Use Bearer Token for API call and pass value of constant authkey as token for Authentication
	e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == authkey, nil
	}))
	e.DELETE("/deletePost", PostDelete)
	e.POST("/createPost", Create)
	initDB()
	e.Logger.Fatal(e.Start(":8002"))
	defer DB.Close()
}
func initDB() {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error in connecting Db %v", err)
		panic(err)
	}
}

//Create to handle post creation
func Create(c echo.Context) error {
	post := PostModel{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&post)
	if err != nil {
		log.Fatalf("Failed reading request body %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error)
	}
	log.Printf("Post Creation Object %#v", post)
	sqlStatement := `INSERT INTO public.post(id, "userId", title, description, created) sVALUES ($1, $2, $3, $4,$5)`
	_, err = DB.Exec(sqlStatement, post.ID, post.UserID, post.Title, post.Description, time.Now())
	if err != nil {
		panic(err)
	}

	user := User{
		UserID:       post.UserID,
		IncPostCount: true,
	}
	// Initialize http client
	client := &http.Client{}
	// marshal User to json
	json, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}
	//set HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, "http://localhost:8000/updatepost", bytes.NewBuffer(json))
	if err != nil {
		panic(err)
	}
	// set request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	return c.JSON(http.StatusOK, "Post Created ...")
}

//PostDelete to handle post deletion
func PostDelete(c echo.Context) error {
	post := PostModel{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&post)
	if err != nil {
		log.Fatalf("Failed reading the request body %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error)
	}
	log.Printf("Post Deletion Object %#v", post)
	sqlStatement := `DELETE FROM post WHERE id = $1;`
	_, err = DB.Exec(sqlStatement, post.ID)
	if err != nil {
		panic(err)
	}

	user := User{
		UserID:       post.UserID,
		IncPostCount: false,
	}

	// initialize http client
	client := &http.Client{}

	// marshal User to json
	json, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}

	// set HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, "http://localhost:8000/updatepost", bytes.NewBuffer(json))
	if err != nil {
		panic(err)
	}

	// set request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	return c.JSON(http.StatusOK, "Post Deleted ...")
}

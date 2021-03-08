package main

import (
	"fmt"
	"net/http"
	"log"
	"github.com/jackc/pgx/v4/pgxpool"
	"os"
	"context"
	"auth/pageLoader"
	"auth/pageStruct"
)

var db *pgxpool.Pool


func main() {
	dbpool, err := pgxpool.Connect(context.Background(),os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	db = dbpool
	defer db.Close()
	http.HandleFunc("/",handleSessionState)
	http.HandleFunc("/login",handleSessionState)
	log.Fatal(http.ListenAndServe(":8080",nil))
}


func handleSessionState (w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	_, err := r.Cookie("session")
	paths := map[string]func(w2 http.ResponseWriter,r2 *http.Request) {
		"/":homePage,
		"/login":login,
	}
	if err != nil {
		if r.URL.Path == "/login" {
			login(w,r)
			return
		}
		fmt.Println("redirecting",err)
		http.Redirect(w,r,"/login",302)
		return
	}
	if r.URL.Path == "/login" {
		http.Redirect(w,r,"/",302)
		return
	}

	if paths[r.URL.Path] != nil {
		paths[r.URL.Path](w,r)
		return
	}
	errorPage :=  pageStruct.PageData{Title: "404"}
	pageLoader.LoadPage(w,"404",errorPage)
}


func homePage(w http.ResponseWriter, r *http.Request) {
	username,cooErr := r.Cookie("session")
	if cooErr != nil {
		fmt.Println("failed to read username",cooErr)
		return
	}
	_,updateErr := db.Exec(context.Background(),"insert into agents values($1,$2)",username.Value, r.UserAgent())

	if updateErr != nil {fmt.Println(updateErr)}

	var refs []string
	useragents ,agentsErr := db.Query(context.Background(),"SELECT * from agents where username=$1",username.Value)

	defer useragents.Close()
	type Agent struct {
		agent string
		username string
	}
	for useragents.Next() {
		agent := Agent{}
		agentsErr = useragents.Scan(&agent.username,&agent.agent)
		if agentsErr != nil {
			fmt.Println("agentErr",agentsErr)

		}
		refs = append(refs,agent.agent)
	}

	if useragents.Err() != nil {

	}

	homePageData := pageStruct.PageData{Title: "Home",Activities: refs}
	pageLoader.LoadPage(w,"index",homePageData)
}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Login page")

	username := r.FormValue("username")
	psswd := r.FormValue("password")
	if username == "" && psswd == "" {
		loginData := pageStruct.PageData{Title: "Login",Login: true}
		pageLoader.LoadPage(w,"login",loginData)

		return
	}
	if userExists(username,psswd) {
		fmt.Println("user exist")
		http.Redirect(w,r,"/",302)
		return
	}
	_,err := createUser(username,psswd)
	if err != nil {
		fmt.Fprintf(w,"Failed to create account")
	}
	loginData := pageStruct.PageData{Title: "Login",Login: true}
	pageLoader.LoadPage(w,"login",loginData)

	fmt.Print("Failed to create user")
	cookie := http.Cookie{Name:"session",Value:username}
	http.SetCookie(w,&cookie)
	http.Redirect(w,r,"/",302)
}

func createUser (username string,psswd string) (bool,error) {
	_,err := db.Exec(context.Background(), "insert into auth values($1,$2)",username,psswd)
	if err != nil {
		fmt.Println(err)
		return false ,err
	}
	return true,nil


}

func userExists (username string, password string) bool {
	var savedUsername string;
	err := db.QueryRow(context.Background(),"select username from auth where username=$1",username).Scan(&savedUsername)
	fmt.Println("query Err",err)
	fmt.Println("found",savedUsername)
	if err != nil {
		return false
	}
	if savedUsername != "" {
		return true
	}
	return false
}


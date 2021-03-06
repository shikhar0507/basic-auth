package main

import (
       "fmt"
       "net/http"
       "log"
       "html/template"
       "github.com/jackc/pgx/v4/pgxpool"
       "os"
       	"context"
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
     
     log.Fatal(http.ListenAndServe(":8080",nil))
}


func handleSessionState (w http.ResponseWriter, r *http.Request) {
     fmt.Println(r.URL.Path)
     sessionCookie, err := r.Cookie("session")
     
     if err != nil {
     	fmt.Println("redirecting",err)
	login(w,r)
	return
     }
     fmt.Println(sessionCookie)
     switch path := r.URL.Path; path {
     	    case "/home", "/login":
	    	  homePage(w,r)
            default:
		fmt.Println("route not found")
	   	
     }
 
}


func homePage(w http.ResponseWriter, r *http.Request) {
     username,cooErr := r.Cookie("session")
     if cooErr != nil {
     	fmt.Println("failed to read username")
     }
     fmt.Println("home page cookie",username.Value)
     _,updateErr := db.Exec(context.Background(),"insert into agents values($1,$2)",username.Value, r.UserAgent())
     if updateErr != nil {fmt.Println(updateErr)}
     t,parseErr := template.ParseFiles("./public/index.html")
     if parseErr != nil {
     	fmt.Println(parseErr)
	return
     }
     type Home struct {
     	  Title string
	  Headline string
	  Refs []string
     }
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
	 fmt.Println("agent",agent.agent)
	 refs = append(refs,agent.agent)
     }

     if useragents.Err() != nil {
     	
     }
     

     page := Home {
     	  Title: "Login",
	  Headline : "This is Login Page",
	  Refs :  refs,
     }
     err := t.Execute(w,page)
     if err != nil {
     	fmt.Println(err)
     }
}

func login(w http.ResponseWriter, r *http.Request) {
    
     
     username := r.FormValue("username")
     psswd := r.FormValue("password")
     
     if username != "" && psswd != ""  {
     	
     	if userExists(username,psswd) {
	   fmt.Println("user exist")
	   homePage(w,r)
	   return
	}
	if createUser(username,psswd) {
	   cookie := http.Cookie{Name:"session",Value:username}
	   http.SetCookie(w,&cookie)
	   homePage(w,r)
	}
     	return
     }
     type Login struct {
     	  Title string
	  Headline string
     }
     
     t,parseErr := template.ParseFiles("./public/login.html")
     if parseErr != nil {
     	fmt.Println(parseErr)
	return
     }
     page := Login {
     	  Title: "Login",
	  Headline : "This is Login Page",
     }
     err := t.Execute(w,page)
     if err != nil {
     	fmt.Println(err)
     }
}

func createUser (username string,psswd string) bool {
    _,err := db.Exec(context.Background(), "insert into auth values($1,$2)",username,psswd)
    if err != nil {
       fmt.Println(err)
       return false
    }
    
    fmt.Println("user created successfully")
    return true
  
    
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


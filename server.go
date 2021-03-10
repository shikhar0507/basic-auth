package main

import (
	"auth/pageLoader"
	"auth/pageStruct"
	"auth/requestDecoder"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"time"
)

var db *pgxpool.Pool

type AuthBody struct {
	Username string
	Psswd string
}
func main() {



	dbpool, err := pgxpool.Connect(context.Background(),os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	db = dbpool
	defer db.Close()

	http.HandleFunc("/",handleSessionState)
	http.HandleFunc("/login",login)
	http.HandleFunc("/signup",signup)

	//api
	http.HandleFunc("/signup-user",handleSignup)
	http.HandleFunc("/login-user",handleSignin)
	http.HandleFunc("/logout",handleLogout)
	http.HandleFunc("/history",handleHistory)
	http.HandleFunc("/favicon.ico", faviconHandler)
	log.Fatal(http.ListenAndServe("127.0.0.1:8080",nil))
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/favicon.ico")
}

func handleSessionState (w http.ResponseWriter, r *http.Request) {

	_, err := r.Cookie("session")

	if err != nil {
		if r.URL.Path == "/signup" {
			fmt.Println("load signup")
			signup(w,r)
			return
		}

		if r.URL.Path == "/login" {
			fmt.Println("load login")
			login(w,r)
			return
		}
		fmt.Println("redirecting",err)
		http.Redirect(w,r,"/signup",302)
		return
	}

	sessionUsername,_,err := getSession(r)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w,r,"/login",http.StatusNotFound)
	}

	paths := map[string]func(w2 http.ResponseWriter,r2 *http.Request,username string) {
		"/":homePage,
		"/login":homePage,
		"/signup":homePage,

	}
	if paths[r.URL.Path] != nil {
		paths[r.URL.Path](w,r,sessionUsername)
		return
	}
	errorPage :=  pageStruct.PageData{Title: "404"}
	pageLoader.LoadPage(w,"404",errorPage)
}


func getSession(r *http.Request) (string,string,error) {
	sessionId := getSessionCookie(r)
	var sessionResult string
	var sessionUsername string
	err := db.QueryRow(context.Background(),"select * from sessions where sessionid=$1",sessionId).Scan(&sessionUsername,&sessionResult)
	if err != nil {
		return "", "", err
	}
	return sessionUsername,sessionResult,nil
}

func homePage(w http.ResponseWriter, r *http.Request,username string) {

	fmt.Println(username)
	if username == "" {
		panic("empty username")
		return
	}
	_,updateErr := db.Exec(context.Background(),"insert into agents values($1,$2)",username, r.UserAgent())
	//
	if updateErr != nil {fmt.Println(updateErr)}
	//
	var refs []string
	useragents ,agentsErr := db.Query(context.Background(),"SELECT * from agents where username=$1",username)
	//
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
		fmt.Println("agentErr",useragents.Err().Error())

	}
	homePageData := pageStruct.PageData{Title: "Home",Activities: refs}
	pageLoader.LoadPage(w,"/",homePageData)
}


func handleSignin(w http.ResponseWriter,r *http.Request) {
	var signinBody AuthBody
	result := requestDecoder.Decode(w,r,&signinBody)
	if signinBody.Username == "" {
		http.Error(w,"Username cannot be empty",http.StatusBadRequest)
		return
	}
	if signinBody.Psswd == "" {
		http.Error(w,"Password cannot be empty",http.StatusBadRequest)
		return
	}
	if result.Status != 200 {
		http.Error(w,result.Message,result.Status)
		return
	}
	if userExists(signinBody.Username,signinBody.Psswd) == false {
		fmt.Println("User does not exist")
		result.Message = "Account does not exist"
		result.Status = http.StatusNotFound
		r,err := json.Marshal(result)
		w.WriteHeader(http.StatusNotFound)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprint(w,string(r))
		return
	}
	uuid,err := generateUUID()
	if err != nil {
		http.Error(w,"Internal Server error",http.StatusInternalServerError)
		return
	}
	_,err = db.Exec(context.Background(),"insert into sessions values($1,$2)",signinBody.Username,uuid)
	if err != nil {
		fmt.Println(err)
		http.Error(w,"Internal Server error",http.StatusInternalServerError)
		return
	}
	//fmt.Print("Failed to create user")
	cookie := http.Cookie{Name:"session",Value:uuid}
	http.SetCookie(w,&cookie)
	result.Message = "Login successfull"
	success, er := json.Marshal(result)
	if er != nil {
		log.Fatal(er)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w,string(success))


}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	type History struct {
		Host string
		Url string
		Time time.Time
		Title string
	}

	var history History
	history.Time = time.Now()

	requestResult := requestDecoder.Decode(w,r,&history)
	if requestResult.Status != 200 {
		fmt.Fprintf(w,requestResult.Message,requestResult.Status)
		return
	}
	if history.Host == "" || history.Url == "" {
		fmt.Fprintf(w,"Both host and url are required",http.StatusBadRequest)
		return
	}


}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	uname,uid,err := getSession(r)
	fmt.Println("logout")
	if err != nil {
		log.Fatal(err)
		http.Error(w,"Error logging out",http.StatusInternalServerError)
		return
	}
	_,err = db.Exec(context.Background(),"delete from sessions where username=$1 AND sessionid=$2",uname,uid)
	if err != nil {
		log.Fatal(err)

		http.Error(w,"Error logging out",http.StatusInternalServerError)
		return
	}
	coo,err := r.Cookie("session")
	coo.MaxAge = -1
	coo.Value = ""
	http.SetCookie(w,coo)
	type Logout struct {
		 Message string
		 Status int
	}
	var logout Logout
	success ,err := json.Marshal(&logout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(w,string(success))
}

// 1) parse the request and check if format matches
// 2) check if user exists
// 3) if exists then reject else create new user

func handleSignup(w http.ResponseWriter,r *http.Request) {

	var signupBody AuthBody
	//var result requestDecoder.Result

	result := requestDecoder.Decode(w,r,&signupBody)
	if signupBody.Username == "" {
		http.Error(w,"Username cannot be empty",http.StatusBadRequest)
		return
	}

	if signupBody.Psswd == "" {
		http.Error(w,"Password cannot be empty",http.StatusBadRequest)
		return
	}
	fmt.Println("check for  user")
	if userExists(signupBody.Username,signupBody.Psswd) {

		result.Message = "Account already  exist"
		result.Status = http.StatusConflict
		r,err := json.Marshal(result)
		w.WriteHeader(http.StatusConflict)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprint(w,string(r))
		return
	}
	if result.Status == 200 {
		// if request is successfully parsed
		err := createUser(signupBody.Username,signupBody.Psswd)
		if err != nil {
			http.Error(w,"Error creating success response",http.StatusInternalServerError)
			return
		}
		successResponse, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			http.Error(w,"Error creating success response",http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w,string(successResponse))

		return
	}
	fmt.Println("error",result)
	http.Error(w,result.Message,result.Status)
}


func getSessionCookie (r *http.Request) string {
	coo,err := r.Cookie("session")
	if err != nil {
		return ""
	}
	return coo.Value
}

func login(w http.ResponseWriter, r *http.Request) {
	loginData := pageStruct.PageData{Title: "Login",Login: true}
	pageLoader.LoadPage(w,"/login",loginData)
}

func signup(w http.ResponseWriter, r *http.Request) {
	loginData := pageStruct.PageData{Title: "Signup",Login: true}
	err := pageLoader.LoadPage(w,"/signup",loginData)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Fprint(w,"Hello world")
}


func generatePsswdHash (psswd string) ([]byte,error) {

	bytePsswd := []byte(psswd)
	hash,err := bcrypt.GenerateFromPassword(bytePsswd,bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
		return nil,err
	}
	return hash,nil

}

func createUser (username string,psswd string) error {
	hash, err := generatePsswdHash(psswd)
	if err != nil {
		return err
	}
	fmt.Println("hash",string(hash))
	_,err = db.Exec(context.Background(), "insert into auth values($1,$2)",username,hash)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil


}

func generateUUID () (string,error) {
	b := make([]byte,16)
	_,err:= rand.Read(b)
	if err != nil {
		return "",err
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x",b[0:4],b[4:6],b[6:8],b[8:10],b[10:]),nil
}

func userExists (username string,psswd string) bool {
	var savedUsername string;
	var savedHash string;
	err := db.QueryRow(context.Background(),"select * from auth where username=$1",username).Scan(&savedUsername,&savedHash)
	if err != nil {
		return false
	}
	fmt.Println(savedUsername,savedHash)

	er := bcrypt.CompareHashAndPassword([]byte(savedHash),[]byte(psswd))
	if er != nil {
		return false
	}
	fmt.Println("isSame")
	return true

}



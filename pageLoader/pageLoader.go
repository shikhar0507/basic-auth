package pageLoader
import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"auth/pageStruct"

)


var errorPage,loginPage,signupPage,homePage *template.Template
//var loginPage *template.Template
//var signupPage *template.Template
//var homePage *template.Template

var parseErr error
func init() {
	// pre parse the templates
	fmt.Println("Init loader")
	homePage ,parseErr = template.ParseFiles("./public/index.html")
	errorPage, parseErr = template.ParseFiles("./public/404.html")
	loginPage ,parseErr = template.ParseFiles("./public/login.html")
	signupPage ,parseErr = template.ParseFiles("./public/signup.html")

	if parseErr != nil {
		log.Fatal(parseErr)
	}
}

func LoadPage(writer http.ResponseWriter, filename string, data pageStruct.PageData) error  {
	var page *template.Template
	switch filename {
	case "/":
		page = homePage
		break
	case "/login":
		page = loginPage
		break
	case "/signup":
		page = signupPage
		break
	case "404":
		page = errorPage
		break
	}
	if page == nil {
		return errors.New("template-not-found")
	}
	err := page.Execute(writer,data)

	if err != nil {
		fmt.Println("template execute",err)
		return err
	}
	return nil
}

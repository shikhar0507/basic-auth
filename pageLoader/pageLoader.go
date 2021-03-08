package pageLoader
import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"auth/pageStruct"

)


var errorPage *template.Template
var requestedPage *template.Template
var parseErr error
func init() {
	// pre parse the templates
	fmt.Println("Init loader")
	requestedPage ,parseErr = template.ParseFiles("./public/index.html")
	errorPage, parseErr = template.ParseFiles("./public/404.html")
	if parseErr != nil {
		log.Fatal(parseErr)
	}
}

func LoadPage(writer http.ResponseWriter, filename string, data pageStruct.PageData) error  {
	page := requestedPage
	if filename == "404" {
		page = errorPage
	}
	fmt.Println(data)
	err := page.Execute(writer,data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return errors.New("template-not-found")
}

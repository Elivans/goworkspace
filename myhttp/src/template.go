  import (
  	"http"
  	"io/ioutil"
  	"os"
  	"template"
  )
  
  func editHandler(w http.ResponseWriter, r *http.Request) {
  	title := r.URL.Path[lenPath:]
  	p, err := loadPage(title)
  	if err != nil {
  		p = &page{title: title}
  	}
  	t, _ := template.ParseFile("edit.html", nil)
  	t.Execute(p, w)
  }
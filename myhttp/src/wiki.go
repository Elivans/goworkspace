// wiki.go
package main

import (
	"fmt"
	"io/ioutil"
)

type page struct {
	title string
	body  []byte
}

func (p *page) save() error {
	filename := p.title + ".txt"
	return ioutil.WriteFile(filename, p.body, 0600)
}

func loadPage(title string) (*page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &page{title: title, body: body}, nil
}
func main() {
	p1 := &page{title: "TestPage", body: []byte("This is a sample page.")}
	p1.save()
	p2, _ := loadPage("TestPage")
	fmt.Println(string(p2.body))
}

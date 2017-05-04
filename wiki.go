// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func loginHandler(w http.ResponseWriter, r *http.Request, title string) {
	if r.Method == "POST" {
		r.ParseForm()
        // logic part of log in
        username := strings.Join(r.Form["username"]," ")
        password := strings.Join(r.Form["password"]," ")

        //Open Database
        db, err := sql.Open("mysql","root:1234@tcp(127.0.0.1:3306)/golang")
		if err != nil {
			fmt.Println(err)
		}

		//Find user
		var id int
		rows, err := db.Query("SELECT * FROM USER WHERE username=? AND password=?", username, password)
		i := 0
		for rows.Next(){
		   i++
		   err = rows.Scan(&id)
		   //handle error and process user
		}
		if i == 0 {
		  fmt.Println("meong")
		} else {
			fmt.Printf("Username is %s\n", id)
		}

		defer rows.Close()
		defer db.Close()


	} else {
		p, err := loadPage(title)
		if err != nil {
			http.Redirect(w, r, "/regiister/"+title, http.StatusFound)
			return
		}
		renderTemplate(w, "login", p)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request, title string) {
	if r.Method == "GET" {
		p, err := loadPage(title)
		if err != nil {
			p = &Page{Title: title}
		}
		renderTemplate(w, "register", p)
	} else {
		r.ParseForm()
        // logic part of register
        name := strings.Join(r.Form["name"]," ")
        username := strings.Join(r.Form["username"]," ")
        password := strings.Join(r.Form["password"]," ")
        email := strings.Join(r.Form["email"]," ")

        //Open Database
        db, err := sql.Open("mysql","root:1234@tcp(127.0.0.1:3306)/golang")
		if err != nil {
			fmt.Println(err)
		}

		//Insert User Information
	    stmt, es := db.Prepare("INSERT INTO USER(name,username, password, email) VALUES (?,?,?,?)")
	    if es != nil {
	        panic(es.Error())
	    }
	    _, er := stmt.Exec(name, username, password, email)
	    if er != nil {
	        panic(er.Error())
	    }   
		defer db.Close()
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseFiles("register.html", "login.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(register|save|login)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func validate() {
	
}

func main() {
	http.HandleFunc("/login/", makeHandler(loginHandler))
	http.HandleFunc("/register/", makeHandler(registerHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	http.ListenAndServe(":8080", nil)
}

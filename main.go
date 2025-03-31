package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"net/http/cgi"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

type Application struct {
	fio, phone, email, birthdate, gender, bio string
	langs                                     []string
}

func validate(appl Application) []string {
	var re *regexp.Regexp

	var valid []string

	pattern := `^([А-ЯA-Z][а-яa-z]+ ){2}[А-ЯA-Z][а-яa-z]+$`
	re = regexp.MustCompile(pattern)

	if !re.MatchString(appl.fio) {
		valid = append(valid, "Поле ФИО некорректно заполнено")
	}

	pattern = `^(\+7|8)9\d{9}$`
	re = regexp.MustCompile(pattern)

	if !re.MatchString(appl.phone) {
		valid = append(valid, "Поле Телефон некорректно заполнено")
	}

	pattern = `^[A-Za-z][\w\.-_]+@\w+(\.[a-z]{2,})+$`
	re = regexp.MustCompile(pattern)

	if !re.MatchString(appl.email) {
		valid = append(valid, "Поле E-mail некорректно заполнено")
	}

	return valid
}

func insertData(appl Application) {
	db, _ := sql.Open("mysql", "u68870:4189913@/u68870")
	defer db.Close()

	insert, _ := db.Query(fmt.Sprintf("INSERT INTO APPLICATION(NAME, PHONE, EMAIL, BIRTHDATE, GENDER, BIO) VALUES ('%s', '%s', '%s', '%s', '%s', '%s')", appl.fio, appl.phone, appl.email, appl.birthdate, appl.gender, appl.bio))
	defer insert.Close()

	sel, _ := db.Query("SELECT ID FROM APPLICATION ORDER BY ID DESC LIMIT 1")
	defer sel.Close()

	var id int
	for sel.Next() {
		sel.Scan(&id)
	}

	for _, name := range appl.langs {
		sel, _ := db.Query(fmt.Sprintf("SELECT ID FROM ProgrammingLanguages WHERE NAME='%s'", name))
		defer sel.Close()

		var plId int
		for sel.Next() {
			sel.Scan(&plId)
		}

		insert, _ := db.Query(fmt.Sprintf("INSERT INTO APPLICATION_PL (APPLICATION_ID, PL_ID) VALUES ('%d', '%d')", id, plId))
		defer insert.Close()
	}
}

func applicationHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("index.html")

	var valid []string

	if r.Method == http.MethodPost {
		r.ParseForm()

		appl := Application{
			fio:       r.FormValue("fio"),
			phone:     r.FormValue("phone"),
			email:     r.FormValue("email"),
			birthdate: r.FormValue("birthdate"),
			gender:    r.FormValue("gender"),
			langs:     r.PostForm["langs[]"],
			bio:       r.FormValue("bio")}

		valid = validate(appl)

		if len(valid) == 0 {
			valid = append(valid, "Данные успешно сохранены")
			insertData(appl)
		}
	}

	tmpl.Execute(w, valid)
}

func main() {
	cgi.Serve(http.HandlerFunc(applicationHandler))
}

package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"net/http/cgi"
	"net/url"
	"regexp"
	"time"

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

	if appl.fio == "" {
		valid = append(valid, "Заполните ФИО")
	} else if !re.MatchString(appl.fio) {
		valid = append(valid, "ФИО должно состоять из 3 слов (кириллица/латиница), каждое с заглавной буквы")
	}

	pattern = `^(\+7|8)9\d{9}$`
	re = regexp.MustCompile(pattern)

	if appl.phone == "" {
		valid = append(valid, "Заполните Телефон")
	} else if !re.MatchString(appl.phone) {
		valid = append(valid, "Телефон должен начинаться с +7 или 8 и содержать 11 цифр")
	}

	pattern = `^[A-Za-z][\w\.-_]+@\w+(\.[a-z]{2,})+$`
	re = regexp.MustCompile(pattern)

	if appl.email == "" {
		valid = append(valid, "Заполните E-mail")
	} else if !re.MatchString(appl.email) {
		valid = append(valid, "Email должен быть в формате example@domain.com и начинаться с буквы")
	}

	if len(appl.langs) == 0 {
		valid = append(valid, "Выберите хотя бы 1 язык программирования")
	}

	pattern = `^\d{4}(-\d{2}){2}$`
	re = regexp.MustCompile(pattern)
	if !re.MatchString(appl.birthdate) {
		valid = append(valid, "Заполните Дату рождения")
	}

	if appl.bio == "" {
		valid = append(valid, "Заполните Биографию")
	}

	return valid
}

func insertData(appl Application, w http.ResponseWriter) {
	db, err := sql.Open("mysql", "u68870:4189913@/u68870")

	if err != nil {
		fmt.Fprintf(w, "Ошибка подключения к БД")
		return
	}

	defer db.Close()

	insert, err := db.Query(fmt.Sprintf("INSERT INTO APPLICATION(NAME, PHONE, EMAIL, BIRTHDATE, GENDER, BIO) VALUES ('%s', '%s', '%s', '%s', '%s', '%s')", appl.fio, appl.phone, appl.email, appl.birthdate, appl.gender, appl.bio))
	defer insert.Close()

	if err != nil {
		fmt.Fprintf(w, "Ошибка вставки")
		return
	}

	sel, err := db.Query("SELECT ID FROM APPLICATION ORDER BY ID DESC LIMIT 1")
	defer sel.Close()

	if err != nil {
		fmt.Fprintf(w, "Ошибка выборки")
		return
	}

	var id int
	for sel.Next() {
		sel.Scan(&id)
	}

	for _, name := range appl.langs {
		sel, err := db.Query(fmt.Sprintf("SELECT ID FROM ProgrammingLanguages WHERE NAME='%s'", name))
		defer sel.Close()

		if err != nil {
			fmt.Fprintf(w, "Ошибка выборки SELECT ID FROM ProgrammingLanguages WHERE NAME")
			return
		}

		var plId int
		for sel.Next() {
			sel.Scan(&plId)
		}

		insert, err := db.Query(fmt.Sprintf("INSERT INTO APPLICATION_PL (APPLICATION_ID, PL_ID) VALUES ('%d', '%d')", id, plId))

		if err != nil {
			fmt.Fprintf(w, "Ошибка вставки INSERT INTO APPLICATION_PL (APPLICATION_ID, PL_ID) VALUES")
			return
		}

		defer insert.Close()
	}
}

func getCookieValue(r *http.Request, name string) string {
	if cookie, err := r.Cookie(name); err == nil {
		val, _ := url.QueryUnescape(cookie.Value)
		return val
	}
	return ""
}

func setCookie(w http.ResponseWriter, name, value string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		Expires:  expires,
		Path:     "/",
		HttpOnly: true,
	})
}

func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	})
}

func applicationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("index.html")

	if err != nil {
		fmt.Fprintf(w, "Ошибка template parse files")
		return
	}

	var valid []string
	var formData = make(map[string]string)
	var savedLangs = make(map[string]bool)

	if r.Method == http.MethodGet {
		formData["fio"] = getCookieValue(r, "saved_fio")
		formData["phone"] = getCookieValue(r, "saved_phone")
		formData["email"] = getCookieValue(r, "saved_email")
		formData["birthdate"] = getCookieValue(r, "saved_birthdate")
		formData["gender"] = getCookieValue(r, "saved_gender")
		formData["bio"] = getCookieValue(r, "saved_bio")

		// Загрузка ошибок из cookies
		if errMsg := getCookieValue(r, "error_msg"); errMsg != "" {
			valid = append(valid, errMsg)
			clearCookie(w, "error_msg")
		}
	}

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
			expires := time.Now().Add(365 * 24 * time.Hour)
			setCookie(w, "saved_fio", appl.fio, expires)
			setCookie(w, "saved_phone", appl.phone, expires)
			setCookie(w, "saved_email", appl.email, expires)
			setCookie(w, "saved_birthdate", appl.birthdate, expires)
			setCookie(w, "saved_gender", appl.gender, expires)
			setCookie(w, "saved_bio", appl.bio, expires)

			valid = append(valid, "Данные успешно сохранены")
			insertData(appl, w)
		} else {
			// Сохраняем ошибку в cookie до следующего запроса
			setCookie(w, "error_msg", valid[0], time.Now().Add(5*time.Minute))
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}
	}

	// Помечаем выбранные языки
	for _, lang := range r.PostForm["langs[]"] {
		savedLangs[lang] = true
	}

	data := struct {
		Valid      []string
		FormData   map[string]string
		SavedLangs map[string]bool
	}{
		Valid:      valid,
		FormData:   formData,
		SavedLangs: savedLangs,
	}

	tmpl.Execute(w, data)
}

func main() {
	cgi.Serve(http.HandlerFunc(applicationHandler))
}

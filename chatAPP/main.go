package main

import (
	"flag"
	"github.com/kondo1018008/chat/trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type templateHandler struct {
	once sync.Once
	filename string
	temp1 *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	t.once.Do(func() { //一度だけ関数を呼び出す
		t.temp1 = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
		/*
		template.Must : テンプレのよみこみ
		template.ParseFiles :　テンプレのファイルを指定する？
		filepath.Join : ファイルのパスを結合する
		*/
	})
	print(t.temp1.Execute(w, r)) //wにテンプレをデータとして書き出す
}

func main(){
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	flag.Parse() // parse the flags

	gomniauth.SetSecurityKey("ChatApp19990907")
	gomniauth.WithProviders(
		facebook.New("","","http://localhost:8080/auth/callback/facebook"),
		github.New("","","http://localhost:8080/auth/callback/github"),
		google.New("263741945932-6t9sf5as84afhtdo893s4o30jb31bak7.apps.googleusercontent.com", "AHmWO1hdlKD3S4U8WfUJKB9s", "http://localhost:8080/auth/callback/google"),
		)

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	go r.run()

	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
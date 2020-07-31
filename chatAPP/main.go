package main

import (
	"flag"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"html/template"
	"log"
	"net/http"
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
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil{
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	print(t.temp1.Execute(w, data)) //wにテンプレをデータとして書き出す
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

	r := newRoom()//roomインスタンスの生成。r.tracer以外が初期化される。
	//r.tracer = trace.New(os.Stdout)　//本ではここで初期化されていたが、room.goで初期化することとする。

	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"})) //認証済でないユーザは"/login"にリダイレクトされる。
	http.Handle("/login", &templateHandler{filename: "login.html"})//OAuth認証のプロバイダ選択画面
	http.HandleFunc("/auth/", loginHandler)//プロバイダのページに振り分け
	http.Handle("/room", r) //クライアントがwebsocketにアップグレードされていないので、リクエストを送るとエラーを吐いてサーバが停止する。

	go r.run()

	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
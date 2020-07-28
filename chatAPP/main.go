package main

import (
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
	print(t.temp1.Execute(w, nil)) //wにテンプレをデータとして書き出す
}

func main(){
	r := newRoom()

	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	go r.run()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
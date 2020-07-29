package main

import (
	"flag"
	"github.com/kondo1018008/chat/trace"
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

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	go r.run()

	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
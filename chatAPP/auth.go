package main

import (
	"fmt"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
	"log"
	"net/http"
	"strings"
)

type authHandler struct{
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	if cookie, err := r.Cookie("auth"); err == http.ErrNoCookie || cookie.Value == ""{
		//未認証
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}else if err != nil {
		//何らかのエラーが発生
		panic(err.Error())
	}else{
		//成功，ラップされたハンドラを呼び出す．
		h.next.ServeHTTP(w,r)
	}
}

func MustAuth(handler http.Handler) http.Handler{
	return &authHandler{next: handler}
}
//サードパーティーへのログインの処理を受け持つ
//パスの形式: /auth/{action}/{provider}
func loginHandler(w http.ResponseWriter, r *http.Request){
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		// log.Println("TODO:ログイン処理", provider)
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln("認証プロバイダーの取得に失敗しました：", provider, "-", err)
		}
		loginUrl, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			log.Fatalln("GetBeginAuthURLの呼出中にエラーが発生しました：", provider, "-", err)
		}
		w.Header().Set("Location", loginUrl) //プロバイダのログインURLをリダイレクト先に設定する。
		w.WriteHeader(http.StatusTemporaryRedirect) //ステータスコード設定
	case "callback":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln("認証プロバイダーの取得に失敗しました：", provider, "-", err)
		}

		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			log.Fatalln("認証を完了できませんでした：", provider, "-", err)
		}
		user, err := provider.GetUser(creds)
		if err != nil {
			log.Fatalln("ユーザーの取得に失敗しました：", provider, "-", err)
		}
		authCookieValue := objx.New(map[string]interface{}{
			"name": user.Name(),
			"avatar_url": user.AvatarURL(),
		}).MustBase64()
		http.SetCookie(w, &http.Cookie{
			Name: "auth",
			Value: authCookieValue,
			Path: "/",
		})
		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "アクション%sには非対応です", action)
	}
}
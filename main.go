package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"go_test/session"
	_ "go_test/session/memory"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"
)

var globalSessions *session.Manager

func init() {
	globalSessions, _ = session.NewManager("memory", "gosessionid", 60)
	go globalSessions.GC()
}

func main() {
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/login", login)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/count", count)

	time.AfterFunc(time.Second*10, func() {
		targetUrl := "http://localhost:9090/upload"
		filename := "./upload.gtpl"
		postFile(filename, targetUrl)
	})

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

func count(w http.ResponseWriter, r *http.Request) {
	sess, err := globalSessions.Session(w, r)
	if err != nil || sess == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	ct := sess.Get("countnum")
	if ct == nil {
		sess.Set("countnum", 1)
	} else {
		sess.Set("countnum", (ct.(int) + 1))
	}

	t, _ := template.ParseFiles("count.gtpl")
	w.Header().Set("Conet-Type", "text/html")
	t.Execute(w, sess.Get("countnum"))
}

func login(w http.ResponseWriter, r *http.Request) {
	sess := globalSessions.SessionStart(w, r)
	r.ParseForm()
	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.gtpl")
		err := t.Execute(w, sess.Get("username"))
		if err != nil {
			log.Println()
		}
	} else {
		sess.Set("username", r.Form["username"])
		http.Redirect(w, r, "/count", http.StatusFound)
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println(r.Form)
	fmt.Fprint(w, "pong!")
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		curtime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(curtime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer file.Close()

		fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile("./tmp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer f.Close()
		io.Copy(f, file)
	}
}

func postFile(filename, targetUrl string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Println("error writting to buffer.")
		return err
	}

	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("error openning file.")
		return err
	}
	defer f.Close()

	_, err = io.Copy(fileWriter, f)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(resp.Status)
	fmt.Println(string(respBody))

	return nil
}

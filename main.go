package main

import (
	"context"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/thrgamon/learning_rank/authentication"
	"github.com/thrgamon/learning_rank/env"
	"github.com/thrgamon/learning_rank/repo"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v4/pgxpool"
)

//go:embed views/*
var views embed.FS

var Db *pgxpool.Pool
var Store *sessions.CookieStore
var Templates map[string]*template.Template
var Log *log.Logger

func main() {
	Db = initDB()
	defer Db.Close()

	cacheTemplates()

	Log = log.New(os.Stdout, "logger: ", log.Lshortfile)
	Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

	authentication.Store = Store
	authentication.Db = Db
	authentication.Log = Log

	r := mux.NewRouter()
	r.HandleFunc("/login", authentication.LoginHandler)
	r.HandleFunc("/logout", authentication.Logout)
	r.HandleFunc("/callback", authentication.CallbackHandler)
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/resource", AddResourceHandler)
	r.HandleFunc("/up/{resourceId:[0-9]+}", UpvoteHandler)
	r.HandleFunc("/down/{resourceId:[0-9]+}", DownvoteHandler)
	r.PathPrefix("/public/").HandlerFunc(serveResource)

	srv := &http.Server{
		Handler:      handlers.CombinedLoggingHandler(os.Stdout, r),
		Addr:         "0.0.0.0:" + env.GetEnvWithFallback("PORT", "8080"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	println("Server listening")
	log.Fatal(srv.ListenAndServe())
}

type PageData struct {
	User      repo.User
	Resources []repo.Resource
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getUserFromSession(r)
	resourceRepo := repo.NewResourceRepo(Db)
	resources, err := resourceRepo.GetAll(r.Context(), user.ID)

	if err != nil {
    println(err.Error())
		http.Error(w, "There was an unexpected error", http.StatusInternalServerError)
		Log.Println(err.Error())
		return
	}

	var pageData PageData
	pageData = PageData{Resources: resources, User: user}

	RenderTemplate(w, "home", pageData)
}

func UpvoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceId, _ := strconv.ParseUint(vars["resourceId"], 10, 64)
	user, _ := getUserFromSession(r)
	resourceRepo := repo.NewResourceRepo(Db)

	err := resourceRepo.Upvote(r.Context(), user.ID, uint(resourceId))

	if err != nil {
		http.Error(w, "There was an unexpected error", http.StatusInternalServerError)
		Log.Println(err.Error())
		return
	}

	http.Redirect(w, r, "/", 303)
}

func AddResourceHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	name := r.FormValue("name")
	link := r.FormValue("link")
	tags := r.FormValue("tags")

	resourceRepo := repo.NewResourceRepo(Db)
	user, _ := getUserFromSession(r)
	err := resourceRepo.Add(r.Context(), user.ID, link, name, tags)

	if err != nil {
		http.Error(w, "There was an unexpected error", http.StatusInternalServerError)
		Log.Println(err.Error())
		return
	}

	http.Redirect(w, r, "/", 303)
}

func DownvoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceId, _ := strconv.ParseUint(vars["resourceId"], 10, 64)
	user, _ := getUserFromSession(r)
	resourceRepo := repo.NewResourceRepo(Db)

	err := resourceRepo.Downvote(r.Context(), user.ID, uint(resourceId))

	if err != nil {
		http.Error(w, "There was an unexpected error", http.StatusInternalServerError)
		Log.Println(err.Error())
		return
	}

	http.Redirect(w, r, "/", 303)
}

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := Templates[tmpl].Execute(w, data)

	if err != nil {
		http.Error(w, "There was an unexpected error", http.StatusInternalServerError)
		Log.Println(err.Error())
		return
	}
}

func cacheTemplates() {
	re := regexp.MustCompile(`^[a-zA-Z\/]*\.html`)
	templates := make(map[string]*template.Template)
	// Walk the template directory and parse all templates that aren't fragments
	err := fs.WalkDir(views, ".",
		func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if re.MatchString(path) {
				normalisedPath := strings.TrimSuffix(strings.TrimPrefix(path, "views/"), ".html")
				templates[normalisedPath] = template.Must(
					template.ParseFS(
						views,
						path,
					),
				)
			}

			return nil
		})

	if err != nil {
		log.Fatal(err.Error())
	}

	// Assign to global variable so we can access it when rendering templates
	Templates = templates

}

func initDB() *pgxpool.Pool {
	conn, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}

	return conn
}

// Handler for serving static assets with modified time to help
// caching
func serveResource(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(filepath.Join(".", r.URL.Path))
	if err != nil {
		http.Error(w, r.RequestURI, http.StatusNotFound)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		http.Error(w, r.RequestURI, http.StatusNotFound)
		return
	}
	modTime := fi.ModTime()

	http.ServeContent(w, r, r.URL.Path, modTime, f)
}

func getUserFromSession(r *http.Request) (repo.User, bool) {
	sessionState, _ := Store.Get(r, "auth")
	userRepo := repo.NewUserRepo(Db)
	userId, ok := sessionState.Values["user_id"].(string)

	if ok {
		_, user := userRepo.Get(r.Context(), userId)
		return user, true
	} else {
		return repo.User{}, false
	}
}

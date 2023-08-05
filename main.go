package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

// initialise to load environment variable from .env file
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	r.Get("/", index)
	r.Post("/run", run)
	log.Println("\033[93mBreeze started. Press CTRL+C to quit.\033[0m")
	http.ListenAndServe(":"+os.Getenv("PORT"), r)
}

// index
func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/index.html")
	t.Execute(w, nil)
}

// call the LLM and return the response
func run(w http.ResponseWriter, r *http.Request) {
	prompt := struct {
		Input string `json:"input"`
	}{}
	// decode JSON from client
	err := json.NewDecoder(r.Body).Decode(&prompt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// create the LLM
	llm, err := openai.NewChat(openai.WithModel(os.Getenv("OPENAI_MODEL")))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chatmsg := []schema.ChatMessage{
		schema.SystemChatMessage{Content: "Hello, I am a friendly AI assistant."},
		schema.HumanChatMessage{Content: prompt.Input},
	}

	completion, err := llm.Call(context.Background(), chatmsg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	response := struct {
		Input    string `json:"input"`
		Response string `json:"response"`
	}{
		Input:    prompt.Input,
		Response: completion.GetContent(),
	}
	json.NewEncoder(w).Encode(response)
}

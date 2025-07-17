package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"os"

	"aidityasadhakim/shinkanzen-jp-quiz/internal"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/genai"
)

//go:embed views/*.html llm-knowledgebase/*
var templateFS embed.FS // Declare an embed.FS variable

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}
	return &Templates{
		templates: template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS, "views/*.html", "llm-knowledgebase/*")),
	}
}

type QuizQuestion struct {
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectAnswer int      `json:"correctAnswer"`
	Explanation   string   `json:"explanation"`
}

type PageData struct {
	Title string
	QuizQuestion
}

type AnswerResponse struct {
	Question       string   `json:"question"`
	Options        []string `json:"options"`
	CorrectAnswer  int      `json:"correctAnswer"`
	SelectedAnswer int      `json:"selectedAnswer"`
	Explanation    string   `json:"explanation"`
	IsCorrect      bool     `json:"isCorrect"`
}

func main() {
	godotenv.Load(".env")
	geminiApiKey := os.Getenv("GEMINI_API_KEY")
	e := echo.New()
	e.Use(middleware.Logger())
	e.Renderer = NewTemplates()
	e.Static("/static", "css")
	e.Static("/images", "images")

	if geminiApiKey == "" {
		e.Logger.Fatal("GEMINI_API_KEY is not set in .env file")
	}

	geminiClient, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: geminiApiKey,
	})
	if err != nil {
		e.Logger.Fatal("Failed to create Gemini client:", err)
	}

	grammarData := &internal.GrammarData{}
	grammarData, err = internal.ReadGrammarData("llm-knowledgebase/n3.json", templateFS)
	if err != nil {
		e.Logger.Fatal("Failed to read grammar data:", err)
	}

	quizSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"question": map[string]interface{}{
				"type":        "string",
				"description": "The quiz question in Japanese",
			},
			"options": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"minItems":    4,
				"maxItems":    4,
				"description": "Four answer options in Japanese",
			},
			"correctAnswer": map[string]interface{}{
				"type":        "integer",
				"minimum":     0,
				"maximum":     3,
				"description": "Index of the correct answer (0-3)",
			},
			"explanation": map[string]interface{}{
				"type":        "string",
				"description": "Brief explanation of the correct answer in Japanese",
			},
		},
		"required": []string{"question", "options", "correctAnswer", "explanation"},
	}

	e.GET("/", func(c echo.Context) error {
		pageData := PageData{
			Title: "Shinkanzen JP Quiz",
		}

		return c.Render(200, "index.html", pageData)
	})

	e.GET("/new-question", func(c echo.Context) error {
		randomIndex := rand.Intn(len(grammarData.GrammarPoints))
		randomGrammarPoint := grammarData.GrammarPoints[randomIndex]

		// Call the Gemini API using the client
		ctx := c.Request().Context()
		response, err := geminiClient.Models.GenerateContent(ctx,
			"gemini-2.5-flash",
			genai.Text(fmt.Sprintf("Generate a grammar quiz question based with this grammar point details %+v the question must be in fill in the blank question. Generate a multiple answer choice question with 4 options, one of which is correct. The question should be in Japanese and the options should be in Japanese as well.", randomGrammarPoint)),
			&genai.GenerateContentConfig{
				ResponseMIMEType:   "application/json",
				ResponseJsonSchema: quizSchema,
			},
		)

		if err != nil {
			return c.String(500, "Failed to generate text")
		}

		quiz := QuizQuestion{}
		if err := json.Unmarshal([]byte(response.Text()), &quiz); err != nil {
			return c.String(500, "Failed to parse response")
		}
		return c.Render(200, "question.html", quiz)
	})

	e.Logger.Fatal(e.Start(":8080"))
}

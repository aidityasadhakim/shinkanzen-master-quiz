package main

import (
	"context"
	"encoding/json"
	"html/template"
	"io"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/genai"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type QuizQuestion struct {
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectAnswer int      `json:"correctAnswer"`
	Explanation   string   `json:"explanation"`
}

func main() {
	godotenv.Load(".env")
	geminiApiKey := os.Getenv("GEMINI_API_KEY")
	e := echo.New()
	e.Use(middleware.Logger())
	e.Renderer = NewTemplates()

	if geminiApiKey == "" {
		e.Logger.Fatal("GEMINI_API_KEY is not set in .env file")
	}

	geminiClient, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: geminiApiKey,
	})
	if err != nil {
		e.Logger.Fatal("Failed to create Gemini client:", err)
	}

	// Store the client for later use in handlers
	_ = geminiClient

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", struct{}{})
	})

	e.POST("/clickme", func(c echo.Context) error {
		// Handle the click event here
		return c.Render(200, "clicked", struct{}{})
	})

	e.GET("/ai-test", func(c echo.Context) error {
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

		// Call the Gemini API using the client
		ctx := c.Request().Context()
		response, err := geminiClient.Models.GenerateContent(ctx,
			"gemini-2.5-flash",
			genai.Text("Generate a grammar quiz question based on the Shinkanzen Japanese book series. Generate a multiple answer choice question with 4 options, one of which is correct. The question should be in Japanese and the options should be in Japanese as well."),
			&genai.GenerateContentConfig{
				ResponseMIMEType:   "application/json",
				ResponseJsonSchema: quizSchema,
			},
		)

		quiz := QuizQuestion{}
		if err := json.Unmarshal([]byte(response.Text()), &quiz); err != nil {
			return c.String(500, "Failed to parse response")
		}

		if err != nil {
			return c.String(500, "Failed to generate text")
		}
		return c.Render(200, "quiz", quiz)
	})
	e.Logger.Fatal(e.Start(":8080"))
}

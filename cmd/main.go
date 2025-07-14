package main

import (
	"context"
	// "encoding/json"
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
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}
	return &Templates{
		templates: template.Must(template.New("").Funcs(funcMap).ParseGlob("views/*.html")),
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

	e.GET("/", func(c echo.Context) error {
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
		_ = quizSchema
		_ = geminiClient

		// Call the Gemini API using the client
		// ctx := c.Request().Context()
		// response, err := geminiClient.Models.GenerateContent(ctx,
		// 	"gemini-2.5-flash",
		// 	genai.Text("Generate a grammar quiz question based on the Shinkanzen Japanese book series. Generate a multiple answer choice question with 4 options, one of which is correct. The question should be in Japanese and the options should be in Japanese as well."),
		// 	&genai.GenerateContentConfig{
		// 		ResponseMIMEType:   "application/json",
		// 		ResponseJsonSchema: quizSchema,
		// 	},
		// )

		quiz := QuizQuestion{
			Question:      "Sample Question?",
			Options:       []string{"Option 1", "Option 2", "Option 3", "Option 4"},
			CorrectAnswer: 0,
			Explanation:   "This is a sample explanation.",
		}
		// if err := json.Unmarshal([]byte(response.Text()), &quiz); err != nil {
		// 	return c.String(500, "Failed to parse response")
		// }

		if err != nil {
			return c.String(500, "Failed to generate text")
		}

		pageData := PageData{
			Title:        "Shinkanzen JP Quiz",
			QuizQuestion: quiz,
		}

		return c.Render(200, "index.html", pageData)
	})

	e.GET("/new-question", func(c echo.Context) error {
		// ctx := c.Request().Context()
		quiz := QuizQuestion{
			Question:      "Sample Question2?",
			Options:       []string{"Option 1", "Option 2", "Option 3", "Option 4"},
			CorrectAnswer: 0,
			Explanation:   "This is a sample explanation.",
		}
		return c.Render(200, "question.html", quiz)
	})

	// Handle answer submission
	// e.POST("/answer", func(c echo.Context) error {
	// 	selectedAnswer := c.FormValue("selectedAnswer")
	// 	correctAnswer := c.FormValue("correctAnswer")
	// 	explanation := c.FormValue("explanation")

	// 	// Convert string values to integers
	// 	selected := 0
	// 	correct := 0
	// 	if selectedAnswer != "" {
	// 		if s, err := strconv.Atoi(selectedAnswer); err == nil {
	// 			selected = s
	// 		}
	// 	}
	// 	if correctAnswer != "" {
	// 		if s, err := strconv.Atoi(correctAnswer); err == nil {
	// 			correct = s
	// 		}
	// 	}

	// 	// Get the current question data (you might want to store this in session)
	// 	// For now, we'll generate a new question to get the structure
	// 	ctx := c.Request().Context()
	// 	quizSchema := map[string]interface{}{
	// 		"type": "object",
	// 		"properties": map[string]interface{}{
	// 			"question": map[string]interface{}{
	// 				"type":        "string",
	// 				"description": "The quiz question in Japanese",
	// 			},
	// 			"options": map[string]interface{}{
	// 				"type": "array",
	// 				"items": map[string]interface{}{
	// 					"type": "string",
	// 				},
	// 				"minItems":    4,
	// 				"maxItems":    4,
	// 				"description": "Four answer options in Japanese",
	// 			},
	// 			"correctAnswer": map[string]interface{}{
	// 				"type":        "integer",
	// 				"minimum":     0,
	// 				"maximum":     3,
	// 				"description": "Index of the correct answer (0-3)",
	// 			},
	// 			"explanation": map[string]interface{}{
	// 				"type":        "string",
	// 				"description": "Brief explanation of the correct answer in Japanese",
	// 			},
	// 		},
	// 		"required": []string{"question", "options", "correctAnswer", "explanation"},
	// 	}

	// 	response, err := geminiClient.Models.GenerateContent(ctx,
	// 		"gemini-2.5-flash",
	// 		genai.Text("Generate a grammar quiz question based on the Shinkanzen Japanese book series. Generate a multiple answer choice question with 4 options, one of which is correct. The question should be in Japanese and the options should be in Japanese as well."),
	// 		&genai.GenerateContentConfig{
	// 			ResponseMIMEType:   "application/json",
	// 			ResponseJsonSchema: quizSchema,
	// 		},
	// 	)

	// 	if err != nil {
	// 		return c.String(500, "Failed to generate question for answer page")
	// 	}

	// 	quiz := QuizQuestion{}
	// 	if err := json.Unmarshal([]byte(response.Text()), &quiz); err != nil {
	// 		return c.String(500, "Failed to parse response")
	// 	}

	// 	// Create answer response
	// 	answerResponse := AnswerResponse{
	// 		Question:       quiz.Question,
	// 		Options:        quiz.Options,
	// 		CorrectAnswer:  correct,
	// 		SelectedAnswer: selected,
	// 		Explanation:    explanation,
	// 		IsCorrect:      selected == correct,
	// 	}

	// 	return c.Render(200, "answer", answerResponse)
	// })

	e.Logger.Fatal(e.Start(":8080"))
}

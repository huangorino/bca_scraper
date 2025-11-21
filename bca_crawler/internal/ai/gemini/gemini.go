package gemini

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/genai"
)

func InitGemini(ctx context.Context) *genai.Client {
	// Provide your API key here
	apiKey := ""

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func GenerateText(ctx context.Context, client *genai.Client, prompt string) (string, error) {
	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("[Error] generate text: %w", err)
	}

	return result.Text(), nil
}

func ParseBoardroomChanges(ctx context.Context, client *genai.Client, html string) (string, error) {
	// First, prompt Gemini to parse the HTML
	prompt := "Please parse the following HTML to extract boardroom changes information. " +
		"Focus on identifying the company, the person, their position, and any qualifications mentioned. " +
		"Format the output as a JSON object with 'company' and 'person' keys and 'qualifications' array. " +
		"The 'company' object should include 'name', 'stock_code'. " +
		"The 'person' object should include 'name', 'age', 'gender', 'nationality'. " +
		"Each qualification in the array should have 'level', 'field_of_study', 'institute', and 'additional_info'.\n\n" + html

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("[Error] parsing boardroom changes: %w", err)
	}

	return result.Text(), nil
}

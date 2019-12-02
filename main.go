package rfa

import (
	"context"
	"fmt"
	"log"
	"os"

	//"golang.org/x/oauth2"
	//"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	//"cloud.google.com/go/storage"
	"cloud.google.com/go/vision/apiv1"
)

func main() {
	rfa()
}

func rfa() {
	ocr("./sample.JPG")

}

func ocr(filename string) {
	ctx := context.Background()

	apiKey := "<API_KEY>"
	apiKeyOption := option.WithAPIKey(apiKey)

	client, err := vision.NewImageAnnotatorClient(ctx, apiKeyOption)
	//client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	defer file.Close()
	image, err := vision.NewImageFromReader(file)
	if err != nil {
		log.Fatalf("Failed to create image: %v", err)
	}

	texts, err := client.DetectTexts(ctx, image, nil, 10)
	if err != nil {
		log.Fatalf("Failed to detect labels: %v", err)
	}

	for _, text := range texts {
		fmt.Println(text.Description)
	}
}

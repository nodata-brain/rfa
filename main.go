package rfa

import (
	"context"
	"log"
	"net/url"
	"os"

	//"golang.org/x/oauth2"
	//"golang.org/x/oauth2/google"
	//"google.golang.org/api/option"
	"github.com/ChimeraCoder/anaconda"

	//"cloud.google.com/go/storage"
	"google.golang.org/api/vision/v1"
)

func main() {
	rfa()
}

func rfa() {
	getTweet()
	ocr("./sample.JPG")
}

func getTweet() {
	anaconda.SetConsumerKey(os.Getenv("Key"))
	anaconda.SetConsumerSecret(os.Getenv("Sec"))
	api := anaconda.NewTwitterApi(os.Getenv("Token"), os.Getenv("TokenSec"))

	v := url.Values{}
	v.Add("user_id", "SugitaniDev")

	timeline, err := api.GetUserTimeline(v)
	if err != nil {
		return
	}

	log.Println(timeline)
}

func ocr(filename string) {
	ctx := context.Background()

	visionService, err := vision.NewService(ctx)
	if err != nil {
		return
	}
	log.Println(visionService)
}

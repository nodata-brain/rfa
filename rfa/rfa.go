package rfa

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/ChimeraCoder/anaconda"

	"cloud.google.com/go/vision/apiv1"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

func Rfa(w http.ResponseWriter, r *http.Request) {
	file := getTweet("")
	texts := ocr(file.Name())
	for _, text := range texts {
		fmt.Fprintf(w, text.Description)
	}
	defer file.Close()
}

func getTweet(usr string) *os.File {
	anaconda.SetConsumerKey(os.Getenv("Key"))
	anaconda.SetConsumerSecret(os.Getenv("Sec"))
	api := anaconda.NewTwitterApi(os.Getenv("Token"), os.Getenv("TokenSec"))

	v := url.Values{}
	v.Add("user_id", usr)
	v.Add("count", "1")

	timeline, _ := api.GetUserTimeline(v)
	url := getImg(timeline)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	file := createTemp("save")

	io.Copy(file, resp.Body)
	return file

}
func createTemp(file string) *os.File {
	tmpfile, err := ioutil.TempFile("", file)
	if err != nil {
		panic(err)
	}
	return tmpfile
}

func getImg(timeline []anaconda.Tweet) string {
	return timeline[0].Entities.Media[0].Media_url_https
}

func getTime(timeline []anaconda.Tweet) time.Time {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	t, _ := time.Parse("Mon Jan 02 15:04:05 -0700 2006", timeline[0].CreatedAt)
	return t.In(loc)
}

func ocr(filename string) []*pb.EntityAnnotation {
	ctx := context.Background()

	client, err := vision.NewImageAnnotatorClient(ctx)
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

	return texts
}

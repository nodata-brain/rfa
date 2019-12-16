package rfa

import (
	"context"
	//"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/ChimeraCoder/anaconda"

	"cloud.google.com/go/vision/apiv1"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

type Rf struct {
	Ot  string
	Cal string
	Run string
}

func Rfa(w http.ResponseWriter, r *http.Request) {
	file := getTweet("")
	rf := Rf{}
	texts := ocr(file.Name())
	rf.getRfaData(w, texts)
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

func (rf *Rf) getRfaData(w http.ResponseWriter, texts []*pb.EntityAnnotation) {
	for i, text := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(texts[0].Description, -1) {
		if i == 3 {
			rf.Ot = text
		}
		if i == 5 {
			rf.Cal = text
		}
		if i == 7 {
			rf.Run = text
		}
	}
}

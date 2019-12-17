package rfa

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/ChimeraCoder/anaconda"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/vision/apiv1"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

type Rf struct {
	Ot  string
	Cal string
	Run string
}

func Rfa(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	file, err := getTweet("")
	if err != nil {
		log.Println(err)
		return
	}
	rf := Rf{}
	texts, err := ocr(ctx, file.Name())
	if err != nil {
		log.Println(err)
		return
	}
	rf.getRfaData(w, texts)
	err = rf.insertData(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	defer file.Close()
}

func getTweet(usr string) (*os.File, error) {
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
		return nil, err
	}
	defer resp.Body.Close()

	file, TempErr := createTemp("save")
	if TempErr != nil {
		return nil, TempErr
	}

	io.Copy(file, resp.Body)
	return file, nil

}
func createTemp(file string) (*os.File, error) {
	tmpfile, err := ioutil.TempFile("", file)
	if err != nil {
		return nil, err
	}
	return tmpfile, nil
}

func getImg(timeline []anaconda.Tweet) string {
	return timeline[0].Entities.Media[0].Media_url_https
}

func getTime() string {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	n := time.Now().In(loc)
	f := n.Format("200601021404")
	return f
}

func getPostTime(timeline []anaconda.Tweet) time.Time {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	t, _ := time.Parse("Mon Jan 02 15:04:05 -0700 2006", timeline[0].CreatedAt)
	return t.In(loc)
}

func ocr(ctx context.Context, filename string) ([]*pb.EntityAnnotation, error) {

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return nil, err
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
		return nil, err
	}
	defer file.Close()
	image, err := vision.NewImageFromReader(file)
	if err != nil {
		log.Fatalf("Failed to create image: %v", err)
		return nil, err
	}

	texts, err := client.DetectTexts(ctx, image, nil, 10)
	if err != nil {
		log.Fatalf("Failed to detect labels: %v", err)
		return nil, err
	}

	return texts, nil
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

type Task struct {
	Cal float64
}

func (rf *Rf) insertData(ctx context.Context) error {

	re := regexp.MustCompile(`\d+\.?\d*|\.\d+`).FindString(rf.Cal)
	data, err := strconv.ParseFloat(re, 64)
	if err != nil {
		return err
	}

	client, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	if err != nil {
		return err
	}
	defer client.Close()
	task := Task{Cal: data}

	states := client.Collection("Rfa")
	t := getTime()
	ny := states.Doc(t)
	_, nyerr := ny.Create(ctx, task)
	if nyerr != nil {
		return nyerr
	}
	return nil
}

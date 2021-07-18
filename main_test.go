package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
)

type request struct {
	fileName       string
	template       string
	topic          string
	name           string
	job            string
	eventTime      string
	outputImage    string
	putTwitterInfo bool
}

func initTestData() []request {
	requests := make([]request, 0)
	requests = append(requests, request{
		fileName:       "avatar4.png",
		template:       "1",
		topic:          "Go ile Http Client Paketi Geliştirme ve Test Süreçleri\n",
		name:           "Erhan Yakut",
		job:            "Senior Software Architect at Binalyze",
		eventTime:      "30 Temmuz Cuma 21:00",
		outputImage:    "erkan-template1.png",
		putTwitterInfo: false,
	})
	requests = append(requests, request{
		fileName:       "avatar4.png",
		template:       "2",
		topic:          "Go ile Http Client Paketi Geliştirme ve Test Süreçleri\n",
		name:           "Erhan Yakut",
		job:            "Senior Software Architect at Binalyze",
		eventTime:      "30 Temmuz Cuma 21:00",
		outputImage:    "erkan-template2.png",
		putTwitterInfo: true,
	})
	requests = append(requests, request{
		fileName:       "avatar3.png",
		template:       "1",
		topic:          "Go Compilerına Tilde (~) Operatorü Eklemek",
		name:           "Furkan Türkal",
		job:            "Backend Developer at Trendyol",
		eventTime:      "4 Haziran Cuma 21:00",
		outputImage:    "furkan-template1.png",
		putTwitterInfo: false,
	})
	requests = append(requests, request{
		fileName:       "avatar3.png",
		template:       "2",
		topic:          "Go Compilerına Tilde (~) Operatorü Eklemek",
		name:           "Furkan Türkal",
		job:            "Backend Developer at Trendyol",
		eventTime:      "4 Haziran Cuma 21:00",
		outputImage:    "furkan-template2.png",
		putTwitterInfo: false,
	})
	requests = append(requests, request{
		fileName:       "avatar.jpeg",
		template:       "1",
		topic:          "Go+Vue.js ile Resim Hatırlatma Uygulaması Yapmak",
		name:           "Abdulsamet İleri",
		job:            "Full Stack Developer at Modanisa",
		eventTime:      "17 Haziran Perşembe 21:00",
		outputImage:    "abdulsamet-template1.png",
		putTwitterInfo: false,
	})
	requests = append(requests, request{
		fileName:       "avatar.jpeg",
		template:       "2",
		topic:          "Go+Vue.js ile Resim Hatırlatma Uygulaması Yapmak",
		name:           "Abdulsamet İleri",
		job:            "Full Stack Developer at Modanisa",
		eventTime:      "17 Haziran Perşembe 21:00",
		outputImage:    "abdulsamet-template2.png",
		putTwitterInfo: false,
	})
	requests = append(requests, request{
		fileName:       "avatar2.jpeg",
		template:       "1",
		topic:          "Go İle Network Programlama",
		name:           "Oğuzhan Yılmaz",
		job:            "CTO at Masomo Games",
		eventTime:      "26 Haziran Cumartesi 21:00",
		outputImage:    "oguzhan-template1.png",
		putTwitterInfo: false,
	})
	requests = append(requests, request{
		fileName:       "avatar2.jpeg",
		template:       "2",
		topic:          "Go İle Network Programlama",
		name:           "Oğuzhan Yılmaz",
		job:            "CTO at Masomo Games",
		eventTime:      "26 Haziran Cumartesi 21:00",
		outputImage:    "oguzhan-template2.png",
		putTwitterInfo: false,
	})

	return requests
}

func Test_CreateCoverImage(t *testing.T) {
	t.Run("When avatar is not specified", func(t *testing.T) {
		res, req := createHttpReq(http.MethodGet, "/api/v1/categories", nil)
		e := echo.New()
		echoTextContext := e.NewContext(req, res)
		err := createCoverImage(echoTextContext)
		require.NoError(t, err)

		var mfc map[string]interface{}
		errJson := json.Unmarshal(res.Body.Bytes(), &mfc)
		require.NoError(t, errJson)

		require.Equal(t, "request Content-Type isn't multipart/form-data", mfc["ErrorString"])
	})
	t.Run("All event template samples are generated successfully?", func(t *testing.T) {
		for _, r := range initTestData() {
			body, contentType := fileUploadRequest(r.fileName)
			res, req := createHttpReq(http.MethodPost, "/create", body)
			req.Header.Add("Content-Type", contentType)

			req.Form = url.Values{}
			req.Form.Set("template", r.template)
			req.Form.Set("topic", r.topic)
			req.Form.Set("name", r.name)
			req.Form.Set("job", r.job)
			req.Form.Set("eventTime", r.eventTime)
			req.Form.Set("putTwitterInfo", strconv.FormatBool(r.putTwitterInfo))
			e := echo.New()
			echoTextContext := e.NewContext(req, res)

			templateMap = initTemplateMap()
			templateToProps = initializeTemplateToProps()

			err := createCoverImage(echoTextContext)
			require.NoError(t, err)

			err = ioutil.WriteFile(
				fmt.Sprintf("output/%s", r.outputImage),
				res.Body.Bytes(),
				0777,
			)
			require.NoError(t, err)
		}

	})
}

func createHttpReq(method string, endpoint string, body *bytes.Buffer) (w *httptest.ResponseRecorder, req *http.Request) {
	if body == nil {
		body = bytes.NewBuffer(make([]byte, 512))
	}
	req = httptest.NewRequest(method, endpoint, body)
	rec := httptest.NewRecorder()
	return rec, req
}

func fileUploadRequest(fileName string) (body *bytes.Buffer, contentType string) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	body = new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("avatar", "avatar")
	if err != nil {
		fmt.Println(err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Println(err)
	}
	err = writer.Close()
	if err != nil {
		fmt.Println(err)
	}
	return body, writer.FormDataContentType()
}

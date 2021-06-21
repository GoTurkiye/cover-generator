package main

import (
	"bytes"
	_ "embed"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/labstack/echo"
	"github.com/nfnt/resize"
	"golang.org/x/image/font"
	"image"
	"image/png"
	"net/http"
)

//go:embed static/logo.png
var logo []byte

//go:embed static/index.html
var index []byte

//go:embed fonts/CuprumFFU.ttf
var cuprum []byte

//go:embed templates/1.png
var template1 []byte

//go:embed templates/2.png
var template2 []byte

var templateMap map[string][]byte

type Props struct {
	TopicHexColor           string
	AvatarContainerCircleBg string
	CircleY                 float64
	CircleR                 float64
	NameTextY               float64
	JobTextY                float64
	JobTextColor            string
	EventDateBg             string
	EventRectangleY         float64
}

var templateToProps map[string]Props

func initializeTemplateToProps() map[string]Props {
	initMap := make(map[string]Props, 0)

	initMap["1"] = Props{
		TopicHexColor:           "#2E2D29",
		AvatarContainerCircleBg: "2D414A",
		CircleY:                 590,
		CircleR:                 100,
		NameTextY:               570,
		JobTextY:                590,
		JobTextColor:            "#7F9EA3",
		EventDateBg:             "#476C7C",
		EventRectangleY:         740,
	}
	initMap["2"] = Props{
		TopicHexColor:           "#291D07",
		AvatarContainerCircleBg: "#8F6863",
		CircleY:                 540,
		CircleR:                 100,
		NameTextY:               520,
		JobTextY:                540,
		JobTextColor:            "#47350F",
		EventDateBg:             "#4F3A0B",
		EventRectangleY:         720,
	}
	return initMap
}

func initTemplateMap() map[string][]byte {
	initMap := make(map[string][]byte, 0)
	initMap["1"] = template1
	initMap["2"] = template2

	return initMap
}

func main() {
	templateMap = initTemplateMap()
	templateToProps = initializeTemplateToProps()

	e := echo.New()
	e.HideBanner = true

	e.GET("/", func(c echo.Context) error {
		_, err := c.Response().Write(index)
		return err
	})
	e.GET("/logo", func(c echo.Context) error {
		_, err := c.Response().Write(logo)
		return err
	})
	e.POST("/create", CreateCoverImage)

	e.Logger.Fatal(e.Start(":1323"))
}

func loadFontWithSpecificSize(size float64) (font.Face, error) {
	f, err := truetype.Parse(cuprum)
	if err != nil {
		return nil, err
	}
	face := truetype.NewFace(f, &truetype.Options{
		Size: size,
	})

	return face, nil
}

func CreateCoverImage(c echo.Context) error {
	templateId := c.FormValue("template")

	topic := c.FormValue("topic")
	avatar, errAvatar := c.FormFile("avatar")
	if errAvatar != nil {
		return c.JSON(http.StatusBadRequest, errAvatar)
	}
	name := c.FormValue("name")
	job := c.FormValue("job")
	eventTime := c.FormValue("eventTime")

	img, err := png.Decode(bytes.NewReader(templateMap[templateId]))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	imgWidth := img.Bounds().Size().X
	imgHeight := img.Bounds().Size().Y

	dc := gg.NewContext(imgWidth, imgHeight)
	dc.DrawImage(img, 0, 0)

	face, err := loadFontWithSpecificSize(65)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	dc.SetFontFace(face)

	dc.SetHexColor(templateToProps[templateId].TopicHexColor)
	topicW, _ := dc.MeasureString(topic)
	dc.DrawString(topic,
		float64(imgWidth/2)-(topicW/2),
		390)

	file, err := avatar.Open()
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	defer file.Close()

	avatarImg, _, err := image.Decode(file)

	// Konuşmacı Avatarını Koy
	avatarResized := resize.Resize(200, 200, avatarImg, resize.Lanczos3)
	dc.SetHexColor(templateToProps[templateId].AvatarContainerCircleBg)
	//
	dc.DrawCircle(float64(imgWidth/2), templateToProps[templateId].CircleY, templateToProps[templateId].CircleR)
	dc.Clip()
	dc.DrawImageAnchored(avatarResized, imgWidth/2, int(templateToProps[templateId].CircleY), 0.5, 0.5)
	dc.ResetClip()

	// İsmi ve nerede çalıştığı
	face, _ = loadFontWithSpecificSize(50)
	dc.SetFontFace(face)
	dc.SetHexColor("#00040B")
	dc.DrawString(name, float64(imgWidth/2)+120, templateToProps[templateId].NameTextY)
	face, _ = loadFontWithSpecificSize(30)
	dc.SetFontFace(face)
	dc.SetHexColor(templateToProps[templateId].JobTextColor)
	dc.DrawStringWrapped(job,
		float64(imgWidth/2)+120, templateToProps[templateId].JobTextY, 0, 0, 250, 1, gg.AlignLeft)
	//

	// Event date ve time
	face, _ = loadFontWithSpecificSize(40)
	dc.SetFontFace(face)
	dc.SetHexColor(templateToProps[templateId].EventDateBg)
	wDate, hDate := dc.MeasureString(eventTime)
	dateRectX := float64(imgWidth/2) - (wDate / 2)
	dc.DrawRectangle(dateRectX, templateToProps[templateId].EventRectangleY, wDate+30, hDate+10)
	dc.Fill()

	dc.SetHexColor("#FFFFFF")
	dc.DrawString(eventTime, dateRectX+(wDate+30-wDate)/2, templateToProps[templateId].EventRectangleY+hDate)

	return png.Encode(c.Response().Writer, dc.Image())
}

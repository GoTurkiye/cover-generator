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
	"strconv"
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

//go:embed images/twitter.png
var twitterIcon []byte

//go:embed images/github.png
var githubIcon []byte

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

var templateMap map[string][]byte
var templateToProps map[string]Props

func main() {
	templateMap = initTemplateMap()
	templateToProps = initializeTemplateToProps()

	e := echo.New()
	e.HideBanner = true

	e.GET("/", renderHomePage)
	e.GET("/logo", renderLogo)
	e.POST("/create", createCoverImage)

	e.Logger.Fatal(e.Start(":1323"))
}

func renderHomePage(c echo.Context) error {
	_, err := c.Response().Write(index)
	return err
}

func renderLogo(c echo.Context) error {
	_, err := c.Response().Write(logo)
	return err
}

func createCoverImage(c echo.Context) error {
	templateId := c.FormValue("template")

	topic := c.FormValue("topic")
	avatar, errAvatar := c.FormFile("avatar")
	if errAvatar != nil {
		return c.JSON(http.StatusBadRequest, errAvatar)
	}
	name := c.FormValue("name")
	job := c.FormValue("job")
	eventTime := c.FormValue("eventTime")

	putTwitterInfo, _ := strconv.ParseBool(c.FormValue("putTwitterInfo"))
	twitterName := c.FormValue("twitterName")

	putGithubInfo, _ := strconv.ParseBool(c.FormValue("putGithubInfo"))
	githubName := c.FormValue("githubName")

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

	// Put Speaker Avatar
	avatarResized := resize.Resize(200, 200, avatarImg, resize.Lanczos3)
	dc.SetHexColor(templateToProps[templateId].AvatarContainerCircleBg)
	dc.DrawCircle(float64(imgWidth/2), templateToProps[templateId].CircleY, templateToProps[templateId].CircleR)
	dc.Clip()
	dc.DrawImageAnchored(avatarResized, imgWidth/2, int(templateToProps[templateId].CircleY), 0.5, 0.5)
	dc.ResetClip()

	// Put Speaker's Name and Company
	face, _ = loadFontWithSpecificSize(50)
	dc.SetFontFace(face)
	dc.SetHexColor("#00040B")
	dc.DrawString(name, float64(imgWidth/2)+120, templateToProps[templateId].NameTextY)

	face, _ = loadFontWithSpecificSize(30)
	dc.SetFontFace(face)
	dc.SetHexColor(templateToProps[templateId].JobTextColor)

	const rectangleWidth = 340
	dc.DrawStringWrapped(job,
		float64(imgWidth/2)+120, templateToProps[templateId].JobTextY, 0, 0, rectangleWidth, 1, gg.AlignLeft)

	twitterIconY := templateToProps[templateId].JobTextY
	jobTextWidth, jobTextHeight := dc.MeasureString(job)
	if jobTextWidth > rectangleWidth {
		twitterIconY += jobTextWidth - rectangleWidth
	} else {
		twitterIconY += jobTextHeight + 20
	}

	// Twitter and Github should be together because we set github icon position
	// relative to twitter icon position!

	if putTwitterInfo {
		twitterImage, err := png.Decode(bytes.NewReader(twitterIcon))
		if err == nil {
			dc.DrawImage(twitterImage, (imgWidth/2)+120, int(twitterIconY))
			face, _ = loadFontWithSpecificSize(20)
			dc.SetFontFace(face)
			dc.DrawString("@"+twitterName, float64(imgWidth/2)+150, twitterIconY+20)
		}
	}
	if putGithubInfo {
		githubImage, err := png.Decode(bytes.NewReader(githubIcon))
		if err == nil {
			dc.DrawImage(githubImage, (imgWidth/2)+120, int(twitterIconY+30))
			face, _ = loadFontWithSpecificSize(20)
			dc.SetFontFace(face)
			dc.DrawString("/"+githubName, float64(imgWidth/2)+150, twitterIconY+50)
		}
	}

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

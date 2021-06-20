package main

import (
	_ "embed"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/labstack/echo"
	"github.com/nfnt/resize"
	"golang.org/x/image/font"
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

func main() {
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
	e.GET("/test", CreateCoverImage)
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
	x, _ := strconv.ParseFloat(c.QueryParam("x"), 64)
	y, _ := strconv.ParseFloat(c.QueryParam("y"), 64)
	rw, _ := strconv.ParseFloat(c.QueryParam("rw"), 64)
	rh, _ := strconv.ParseFloat(c.QueryParam("rh"), 64)
	ah, _ := strconv.ParseInt(c.QueryParam("ah"), 10, 32)
	aw, _ := strconv.ParseInt(c.QueryParam("aw"), 10, 32)
	ax, _ := strconv.ParseFloat(c.QueryParam("ax"), 64)
	ay, _ := strconv.ParseFloat(c.QueryParam("ay"), 64)
	ar, _ := strconv.ParseFloat(c.QueryParam("ar"), 64)

	fontSize, _ := strconv.ParseFloat(c.QueryParam("size"), 64)

	_, _, _, _, _, _, _, _, _, _ = x, y, rw, rh, fontSize, ax, ay, ar, ah, aw
	name := c.FormValue("tel")
	_ = name

	img, err := gg.LoadImage("templates/1.png")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	imgWidth := img.Bounds().Size().X
	imgHeight := img.Bounds().Size().Y

	dc := gg.NewContext(imgWidth, imgHeight)
	dc.DrawImage(img, 0, 0)
	face, err := loadFontWithSpecificSize(70)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	dc.SetFontFace(face)

	dc.SetHexColor("#2E2D29")
	dc.DrawStringAnchored("Go+Vue.js ile Resim Hatırlatma Uygulaması Yapmak",
		float64(imgWidth/2), 370, 0.5, 0.5)

	avatar, err := gg.LoadImage("avatar.jpeg")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// Konuşmacı Avatarını Koy
	avatarResized := resize.Resize(200, 200, avatar, resize.Lanczos3)
	dc.SetHexColor("#2D414A")
	dc.DrawCircle(float64(imgWidth/2), 590, 100)
	dc.Clip()
	dc.DrawImageAnchored(avatarResized, imgWidth/2, 590, 0.5, 0.5)
	dc.ResetClip()
	//

	// İsmi ve nerede çalıştığı
	face, _ = loadFontWithSpecificSize(50)
	dc.SetFontFace(face)
	dc.SetHexColor("#00040B")
	dc.DrawString("Abdulsamet İleri", float64(imgWidth/2)+120, 570)
	face, _ = loadFontWithSpecificSize(30)
	dc.SetFontFace(face)
	dc.SetHexColor("#7F9EA3")
	dc.DrawStringWrapped("Full Stack Developer at Modanisa",
		float64(imgWidth/2)+120, 590, 0, 0, 250, 1, gg.AlignLeft)
	//

	// Event date ve time
	face, _ = loadFontWithSpecificSize(50)
	dc.SetFontFace(face)
	dc.SetHexColor("#476C7C")
	date := "17 Haziran Perşembe 21:00"
	wDate, hDate := dc.MeasureString(date)
	dateRectX := float64(imgWidth/2) - (wDate / 2)
	dc.DrawRectangle(dateRectX, 740, wDate+100, hDate+20)
	dc.Fill()
	dc.SetHexColor("#FFFFFF")
	dateStrX := dateRectX + (wDate+100-wDate)/2
	dc.DrawStringAnchored(date, dateStrX, 740+hDate, 0, 0)
	//

	return png.Encode(c.Response().Writer, dc.Image())
}

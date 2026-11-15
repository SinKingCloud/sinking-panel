package captcha

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"sync"

	image1 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_1"
	image10 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_10"
	image11 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_11"
	image12 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_12"
	image13 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_13"
	image14 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_14"
	image15 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_15"
	image16 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_16"
	image2 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_2"
	image3 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_3"
	image4 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_4"
	image5 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_5"
	image6 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_6"
	image7 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_7"
	image8 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_8"
	image9 "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_9"
	"github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/base/helper"
	"github.com/wenlng/go-captcha/v2/slide"
)

var (
	slideOnce        sync.Once
	slideErr         error
	slideGraphs      []*slide.GraphImage
	slideBackgrounds = [...]func(string) ([]byte, error){
		image1.Asset,
		image2.Asset,
		image3.Asset,
		image4.Asset,
		image5.Asset,
		image6.Asset,
		image7.Asset,
		image8.Asset,
		image9.Asset,
		image10.Asset,
		image11.Asset,
		image12.Asset,
		image13.Asset,
		image14.Asset,
		image15.Asset,
		image16.Asset,
	}
)

// GenSlide 生成滑动验证码
func GenSlide() (result *slide.Block, masterImage string, tileImage string, height int, width int, err error) {
	// 图块只读共享；每次仅解码随机选中的原始背景，生成结束后释放。
	slideOnce.Do(func() {
		graphs, err := tiles.GetTiles()
		if err != nil {
			slideErr = fmt.Errorf("加载验证码图块失败: %w", err)
			return
		}
		slideGraphs = make([]*slide.GraphImage, 0, len(graphs))
		for _, graph := range graphs {
			slideGraphs = append(slideGraphs, &slide.GraphImage{
				OverlayImage: graph.OverlayImage,
				MaskImage:    graph.MaskImage,
				ShadowImage:  graph.ShadowImage,
			})
		}
	})
	if slideErr != nil {
		return nil, "", "", 0, 0, slideErr
	}
	index := helper.RandIndex(len(slideBackgrounds))
	data, err := slideBackgrounds[index](fmt.Sprintf("sourcedata/images/image-v2-%d/image.jpg", index+1))
	if err != nil {
		return nil, "", "", 0, 0, fmt.Errorf("加载验证码背景失败: %w", err)
	}
	background, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", "", 0, 0, fmt.Errorf("解码验证码背景失败: %w", err)
	}
	builder := slide.NewBuilder(
		slide.WithGenGraphNumber(1),
		slide.WithEnableGraphVerticalRandom(true),
	)
	builder.SetResources(
		slide.WithGraphImages(slideGraphs),
		slide.WithBackgrounds([]image.Image{background}),
	)
	slideCapt := builder.Make()
	captData, err := slideCapt.Generate()
	if err != nil {
		return nil, "", "", 0, 0, errors.New("生成验证码失败")
	}
	result = captData.GetData()
	if result == nil {
		return nil, "", "", 0, 0, errors.New("创建验证码失败")
	}
	masterImage, err = captData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, "", "", 0, 0, errors.New("生成验证码主图失败")
	}
	tileImage, err = captData.GetTileImage().ToBase64()
	if err != nil {
		return nil, "", "", 0, 0, errors.New("生成验证码副图失败")
	}
	img := slideCapt.GetOptions().GetImageSize()
	height = img.Height
	width = img.Width
	return result, masterImage, tileImage, height, width, err
}

// CheckSlide 判断滑动验证码
func CheckSlide(x int, y int, result *slide.Block) bool {
	if result == nil {
		return false
	}
	return slide.CheckPoint(int64(x), int64(y), int64(result.X), int64(result.Y), 4)
}

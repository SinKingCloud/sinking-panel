# captcha：图片验证码和滑动验证码

导入路径：`server/app/util/captcha`。本包负责生成挑战图片和判断滑动坐标，不负责 HTTP 路由、挑战 ID、过期时间、次数限制或一次性消费。完整业务接入可参考 [鉴权服务](../../service/auth/captcha.go)。

## 字符图片验证码

| API | 返回 |
| --- | --- |
| `NewCaptchaCode(code string) *Image` | 使用指定文本生成图片 |
| `NewCaptchaRandCode(num int) (*Image, string)` | 生成指定字符数的图片及正确答案 |
| `NewTtf() []byte` | 内置字体数据，通常不需要调用方使用 |

字符验证码采用内置字体、256×128 画布和预设干扰/颜色。配置由包内部单例决定，当前不提供按请求修改大小或字体的接口。第一次字体初始化失败会 panic。

```go
package example

import (
    "bytes"
    "image/png"
    "server/app/util/captcha"
)

func NewPNG() (imageBytes []byte, answer string, err error) {
    image, answer := captcha.NewCaptchaRandCode(4)
    var buffer bytes.Buffer
    if err := png.Encode(&buffer, image); err != nil {
        return nil, "", err
    }
    // answer 只保存在服务端，响应前端时只返回 imageBytes 和挑战 ID。
    return buffer.Bytes(), answer, nil
}
```

字符答案如何比较（大小写、空格等）由业务定义。不要把生成函数返回的答案放进公开 JSON。

## 滑动验证码

`GenSlide()` 返回 `(result *slide.Block, masterImage string, tileImage string, height int, width int, err error)`：

- `result` 保存正确位置等挑战数据，应绑定挑战 ID 存入服务端缓存。
- `masterImage`、`tileImage` 为底层图片编码器产生的 Base64 字符串，供前端展示主图和滑块。
- 返回尺寸顺序是 **height、width**，不要颠倒。
- `CheckSlide(x, y, result) bool` 比较提交坐标和正确坐标，容差为 4；result 为 nil 时返回 false。

下面示例展示“生成公开数据 + 服务端保存挑战”的分界，缓存键、请求鉴权和验证消费由外层负责：

```go
package example

import (
    "time"
    "server/app/util/cache"
    "server/app/util/captcha"
    "github.com/wenlng/go-captcha/v2/slide"
)

type SlideImages struct {
    Master string `json:"master"`
    Tile string `json:"tile"`
    Height int `json:"height"`
    Width int `json:"width"`
}

func CreateSlide(store cache.Interface, challengeID string) (*SlideImages, error) {
    result, master, tile, height, width, err := captcha.GenSlide()
    if err != nil { return nil, err }
    store.SetWithExpire("captcha:"+challengeID, result, 2*time.Minute)
    return &SlideImages{master, tile, height, width}, nil
}

func CheckStoredSlide(store cache.Interface, challengeID string, x, y int) bool {
    key := "captcha:"+challengeID
    value, ok := store.Get(key).(*slide.Block)
    store.Delete(key)
    return ok && captcha.CheckSlide(x, y, value)
}
```

该示例的 Get/Delete 不是原子操作；若存在并发验证请求，应在业务层保证同一挑战只被消费一次。挑战 ID 应不可猜测，并绑定当前用户/会话。前端缩放图片后需要把坐标还原到原图坐标系。

## 生命周期和限制

包初始化时加载内置主图和滑块资源；加载失败使用 `log.Fatalln` 结束进程。无需网络下载资源，也无需调用方关闭图片对象。本包使用共享生成器，未提供每请求隔离配置；若扩展并发配置或替换资源，应先确认底层库并发约束。

源码：[字符验证码](captcha.go)、[滑动验证码](slider.go)、[返回类型](entities.go)。相关工具：[cache](../cache/README.md)、[context](../context/README.md)。

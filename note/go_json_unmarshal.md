# json.Unmarshal踩坑+部分源码分析

## 背景
业务中有两个接口, 一个鉴权一个不鉴权, 路径如下: `/nc/getAInfo` 和 `/tk/getAInfo`。 `/tk`表示鉴权,如果是鉴权接口, 则需要先获取token, 然后通过token中记录的某个业务字段AId查询, 如果是`/nc`则表示不鉴权, 直接通过请求body中的{"AId":123}来获取

我们使用的是gin框架, 本质上两个接口都使用了同一个参数结构体, 唯一区别是tk的多了一步校验token和将tokenValue手动赋值的操作, 最终也走到了同一个service的处理函数中。代码示例如下：

```go
// GetAInfoParam 请求参数接收结构体
type GetAInfoParam struct {
	AId int64 `form:"aId"`
	*InternalToken
}

// tk路由规则
router.POST("/tk/getAInfo", 
    func(c *gin.Context) {
			var param model.GetAInfoParam
			// 绑定参数
			app.MustBindAndValid(c, &param)
			// /tk的校验并获取token, 如果校验不通过会直接返回错误
			param.InternalToken = app.MustGetInternalToken(c)
			// 调用service
			service.GetAInfo(c, &param)
		})
// nc路由规则
router.POST("/nc/getAInfo",
    func(c *gin.Context) {
			var param model.GetAInfoParam
			// 绑定参数
			app.MustBindAndValid(c, &param)
			// 调用service
			service.GetAInfo(c, &param)
		})

// service处理规则
func (s *service) GetAInfo(
	c *gin.Context,
	param *model.GetAInfoParam,
) {
	var aId int64 = 0
	if param.InternalToken != nil {
		aId = param.InternalToken.AId
	}
	if param.AId != 0 {
		aId = param.AId
	}
    if aId == 0 {
        // 报错返回错误
        return
    }
    // 后续处理逻辑
}
```

可以看到上面的代码的逻辑非常简单, 在service中, 如果token有值, 则认为上层来源是tk，从token中获取, 如果没有则认为是nc，从请求参数中获取。这段代码已经跑了好几年。直到有一天，突然测试提了个bug，在测试环境调用/nc/getAInfo 会返回报错提示aId为0，完整的请求如下：

```shell
curl 'https://xxx/nc/getAInfo' \
  -H 'accept: application/json, text/plain, */*' \
  -H 'accept-language: zh-CN,zh;q=0.9,en;q=0.8' \
  -H 'cloud-env: rongke' \
  -H 'content-type: application/json' \
  -H 'priority: u=1, i' \
  -H 'sec-ch-ua: "Not(A:Brand";v="99", "Google Chrome";v="133", "Chromium";v="133"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "macOS"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-site' \
  -H 'user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36' \
  --data-raw '{"aId":259}'
```

## 问题分析
请求看起来非常正常, 但怎么就会报错aId=0呢，于是我又重新看了眼代码，哦原来是因为`GetAInfoParam`是一个嵌套结构体, 为了方便取值, 所以我们把app.InternalToken的指针也放进去了, `*app.InternalToken`中也包含一个AId字段。而且最上层的AId字段只有一个 `form:"aId"` 的标签, 并没有声明`json:"aId"`。**所以应该是数据反序列化的时候错误的把aId赋值到了`*app.InternalToken`中**, debug验证了下，确实是这个问题，正当我准备提交代码的时候, 突然觉得问题应该没有那么简单, 这个代码都是几年前写的了, 怎么到今天才出问题呢？于是我又重新分析了一遍代码, 发现问题原来就出在这`*app.InternalToken`里面, 下面附简化版源码


```go 
type InternalToken struct {
    // *Info1 // 一个早期的已废弃结构体
    *Info2 // 一个正常的结构体
}

type Info1 struct {
    AId int64 `json:"aId"`
    // 省略其它...
}

type Info2 struct {
    AId int64 `json:"aId"`
    // 省略其它...
}
```

可以看到, 在`InternalToken`中, `Info1`和`Info2`都是指针的结构, `*Info1` 是一个早期的废弃的结构, 大家正常都会使用`Info2`结构, `InternalToken`的赋值也只会给`Info2`赋值, 但是本着兼容历史代码的原则, `InternalToken`中还保留了`Info1`的字段, 但是查看提交记录, 发现`*Info1`这行在前几天被删掉了。
> ps: 为了方便描述, 上面的示例代码中使用 `// *Info1 // 一个早期的已废弃结构体` 的注释的方式表示`*Info1` 这行已经被删除了


现在问题就比较明显了, 先说结论后分析源码：    
**`*Info1`被删除前**, 由于`Info1`和`Info2`的结构中都有`json:"aId"`的标签, 在Unmarshal的时候, 不知道该给哪个赋值, 反而成功地给`GetAInfoParam.AId`赋了值。所以 由于`InternalToken`中的`Info2`

**`*Info1`被删除后**, 由于只有`Info2`中有`json:"aId"`的标签, 而`GetAInfoParam.AId`字段中没有明确json地标签, 在Unmarshal的时候, 只会给`GetAInfoParam.InternalToken.Info2.AId`赋值, 而`GetAInfoParam.AId`没有被赋值, 所以报错。

##  源码分析
未完成待补充。。
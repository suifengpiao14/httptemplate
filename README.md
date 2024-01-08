# http 请求构造器

**使用场景：**
1. curl 请求转换,当第三方curl请求复杂或者不符合内部规范时，可以增加一层配置模版作为转换，从而屏蔽第三方api对内部的入侵
2. 将对api接口请求转换为templateId+data 模式，能方便实现统一代理请求数据
3. 抓包工具获取到http 协议文本后，可以解析出json结构，进一步生成基本的接口文档字段
**核心功能：**
 解析http协议文本生成http.Request 对象
**周边功能：** 
1. 使用 template/text 解析go模板，生成 http协议文本
2. 将http.Reqeust 对象转换为RequestDTO对象,方便常规思维下修改http请求数据，构造http请求
3. 内置基于 github.com/go-resty/resty/v2 的请求函数RestyRequestFn
4. 内置http 请求日志广播
5. 内置支持http代理的http.Transport 生成函数
6. 内置常用模板函数
7. 实现response解析生成

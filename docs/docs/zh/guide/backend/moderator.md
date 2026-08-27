# 评论审核

Artalk 支持多种评论过滤方式以拦截垃圾评论，支持通过 Akismet、腾讯云内容安全、阿里云内容安全 等在线 API 实现垃圾评论识别和拦截，也可通过本地离线关键词词库过滤评论，结合[验证码](./captcha.md)功能进一步提高评论审核强度，让垃圾评论没有容身之地。

你可以在 [控制中心](../frontend/sidebar.md#设置) 的设置界面修改此配置，也可以通过 [配置文件](./config.md#评论审核-moderator) 或 [环境变量](../env.md#评论审核) 进行配置。

## 配置文件

完整的 `moderator` 配置如下：

```yaml
# 评论审核
moderator:
  pending_default: false # 发表新评论默认为 “待审状态”
  api_fail_block: false # 垃圾检测 API 请求错误仍然拦截
  # akismet.com 反垃圾
  akismet_key: ''
  # OpenAI-compatible LLM 审核
  llm:
    enabled: false
    endpoint: '' # 完整的 Chat Completions 接口地址
    api_key: ''
    headers: {} # 附加请求 Header
    model: ''
    timeout: 10 # 秒
    max_input_chars: 8000
    policy: 将垃圾广告、辱骂、仇恨、色情、违法或明显与页面无关的内容设为 review。
  # 腾讯云文本内容安全 (tms)
  tencent: # https://cloud.tencent.com/document/product/1124/64508
    enabled: false
    secret_id: ''
    secret_key: ''
    region: ap-guangzhou
  # 阿里云内容安全
  aliyun: # https://help.aliyun.com/document_detail/28417.html
    enabled: false
    access_key_id: ''
    access_key_secret: ''
    region: cn-shanghai
  # 关键词词库过滤
  keywords:
    enabled: false
    pending: false # 匹配成功设为待审状态
    files: # 支持多个词库文件
      - './data/词库_1.txt'
    file_sep: "\n" # 词库文件内容分割符
    replace_to: 'x' # 替换字符
```

## 默认待审模式

开启发表新评论默认为 “待审状态”：

```yaml
moderator:
  pending_default: true
```

## Akismet

[Akismet](https://akismet.com/) 是 WordPress 提供的面向全球范围的老牌垃圾拦截 API，通常对一些英文的垃圾评论十分凑效。Akismet 提供了 Personal 免费版本，适用于个人博客站点。

![](/images/akismet/1.png)

你能在 [Akismet 官网](https://akismet.com/) 轻松地申请 `akismet_key`，并填入配置文件中，即可启用 Akismet 垃圾拦截。

![](/images/akismet/2.png)

```yaml
moderator:
  akismet_key: your_key
```

## LLM 评论审核

Artalk 可以将评论正文发送到兼容 OpenAI Chat Completions 的接口。请求只包含评论正文和配置的审核规则，不包含邮箱、IP、User-Agent 或登录标识。

配置完整的接口地址和服务商支持的模型：

```yaml
moderator:
  # 建议开启：服务商出错时让评论进入待审。
  api_fail_block: true
  llm:
    enabled: true
    endpoint: https://api.example.com/v1/chat/completions
    api_key: your_key
    headers:
      CF-Access-Client-Id: your_access_id
      CF-Access-Client-Secret: your_access_secret
    model: your_model
    timeout: 10
    max_input_chars: 8000
    policy: 将垃圾广告、辱骂、仇恨、色情、违法或明显与页面无关的内容设为 review。
```

模型必须返回 JSON 对象，`decision` 只能是 `allow` 或 `review`。`review` 会把评论放入现有待审队列，LLM 不会直接删除评论。评论超过 `max_input_chars`、请求超时、响应无效或服务商报错时，由 `api_fail_block` 决定放行还是转待审。

`headers` 可用于 Cloudflare Access 等网关。Artalk 始终发送 `Content-Type: application/json`；如果配置了 `api_key`，Bearer 认证会覆盖自定义的 `Authorization`。Header 值可能包含凭证，请勿将填写后的值提交到公开配置文件。

也可以通过环境变量传入 JSON 对象：

```bash
ATK_MODERATOR_LLM_HEADERS='{"CF-Access-Client-Id":"your_access_id","CF-Access-Client-Secret":"your_access_secret"}'
```

LLM 审核使用 Artalk 的异步审核管线。如果没有开启 `pending_default`，评论可能在 LLM 返回并将其转为待审前短暂可见。开启 `pending_default` 后，`allow` 结果仍会保持待审状态，需要管理员手动通过。

## 腾讯云文本内容安全

可参考：[腾讯云文档](https://cloud.tencent.com/document/product/1124/64508)

开通「文本内容安全」后，在「访问管理」-「API 密钥管理」新增具有权限的 Secret，然后填入配置：

```yaml
moderator:
  tencent:
    enabled: true
    secret_id: ''
    secret_key: ''
    region: ap-guangzhou
```

## 阿里云内容安全

可参考：[阿里云文档](https://help.aliyun.com/document_detail/28417.html)

开通「阿里云内容安全」后，阿里云后台创建 Access Key 并填入配置：

```yaml
moderator:
  aliyun:
    enabled: true
    access_key_id: ''
    access_key_secret: ''
    region: cn-shanghai
```

## 关键词词库过滤

如果你不想依赖于远程 API，可以在本地配置导入词库文件，让 Artalk 根据词语来检测垃圾评论：

```yaml
moderator:
  keywords:
    enabled: true
    pending: false # 匹配成功设为待审状态
    files: # 支持多个词库文件
      - ./data/词库_1.txt
    file_sep: "\n" # 词库文件内容分割符
    replace_to: 'x' # 替换字符
```

- **pending**：当成功匹配时，是否将评论设为待审核状态。
- **files**：词库文件。允许多个文件，Artalk 启动时会合并词库。
- **file_sep**：词库文件内容分割符。例如：文件中每行一个词语，该项配置 `\n`。
  该配置不能为空。控制中心会将换行、回车和制表符分别显示为 `\n`、`\r`、`\t`；若分隔符需要是字面量 `\n`，请输入 `\\n`。
- **replace_to**：替换字符。例如：该项设置为 `x`，你可以将 `pending` 设置为 `false`，评论自动过审，但匹配到的词语会被替换为 `x`，例如 `fxxk`、`xxxx`。

注：`replace_to` 不建议使用 `*` 星号，应为它和 Markdown 的加粗语法冲突。

## 使用验证码

你可以开启 Artalk 的验证码功能，支持图片和滑动验证码，[参考此处](./captcha.md)。

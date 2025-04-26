
真正的下载代码在`downloader/downloader.go`文件中，核心下载逻辑分为以下几个部分：

1. 核心下载函数是`writeFile`方法（第67-83行），它负责实际的HTTP请求和数据写入：
```go
func (downloader *Downloader) writeFile(url string, file *os.File, headers map[string]string) (int64, error) {
    res, err := request.Request(http.MethodGet, url, nil, headers)
    if err != nil {
        return 0, err
    }
    defer res.Body.Close()

    barWriter := downloader.bar.NewProxyWriter(file)
    // 使用io.Copy读取响应体并写入文件
    written, copyErr := io.Copy(barWriter, res.Body)
    if copyErr != nil && copyErr != io.EOF {
        return written, errors.Errorf("file copy error: %s", copyErr)
    }
    return written, nil
}
```

2. 下载调用流程如下：
   - 入口是`Download`方法（第627行）
   - 对于单个部分的视频，调用`save`或`multiThreadSave`方法（第681-684行）
   - 对于多个部分的视频，使用协程并行下载每个部分（第703-723行）

3. 具体调用逻辑：
   - 视频下载请求由`main.go`启动
   - 经过解析器（parser）解析URL
   - 使用对应网站的提取器（如bilibili提取器）提取视频信息
   - 将提取的信息传递给`downloader.Download`方法进行下载
   - 下载完成后，如果是多部分视频会进行合并（第733-738行）

核心代码中使用了分片下载、断点续传、多线程并行下载等技术，以提高下载效率。对于大文件，还支持按块（chunk）下载，通过设置HTTP Range头实现。

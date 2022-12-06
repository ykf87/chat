package funcs

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	// "github.com/tidwall/gjson"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//删除slice重复值
func RemoveRepByMap(slc []string) []string {
	result := []string{}
	tempMap := map[string]byte{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

//获取文件的 Content-Type
func GetFileContentType(out *os.File) (string, error) {
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}
	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

//随机数
func Random(min, max int64) int64 {
	return rand.Int63n(max-min-1) + min + 1
}

//数据库ip转ipv4
func InetNtoA(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

//IPV4转数据库ip
func InetAtoN(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}

//网络请求
//发起网络请求
//如果发起的是get请求,uri请自行拼接
//uri为完整的http连接地址
func Request(method, uri string, data []byte, header map[string]string, proxy string) ([]byte, error) {
	var body io.Reader
	if method == "POST" && data != nil {
		// cont, err := json.Marshal(data)
		// if err == nil {
		// 	body = bytes.NewBuffer(cont)
		// }
		body = bytes.NewBuffer(data)
	}

	tr := &http.Transport{TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	}}

	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err == nil { //使用传入代理
			tr.Proxy = http.ProxyURL(proxyUrl)
		}
	}

	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest(method, uri, body)
	if header == nil {
		header = make(map[string]string)
		header["Content-Type"] = "application/json"
		header["accept-language"] = "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7"
		header["pragma"] = "no-cache"
		header["cache-control"] = "no-cache"
		header["upgrade-insecure-requests"] = "1"
		header["user-agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36"
	} else {
		// header["Content-Type"] = "application/json"
		// header["accept-language"] = "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7"
		// header["pragma"] = "no-cache"
		// header["cache-control"] = "no-cache"
		// header["upgrade-insecure-requests"] = "1"
		// header["user-agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36"
	}

	for k, v := range header {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return respbody, errors.New(fmt.Sprintf("请求发生错误:\r\n\turi: %s.\r\n\thttpcode: %d.\r\n\tmessage: %s", uri, resp.StatusCode, string(respbody)))
	}
	return respbody, nil
}

//文件base64
//返回 base64格式的正文内容 和 文件content-type
func ParseBase64(content string) (string, string) {
	b, err := regexp.MatchString(`^data:(.+?);base64,`, content)
	if !b {
		fmt.Println(err)
		return "", ""
	}

	re, _ := regexp.Compile(`^data:(.+?);base64,(.+)`)
	allData := re.FindStringSubmatch(content)
	if len(allData) == 3 {
		return allData[2], allData[1]
	}

	return "", ""
}

func GetBase64Type(tp string) string {
	tp = strings.ToLower(tp)
	switch tp {
	case "image/jpeg":
		return ".jpg"
	case "image/pjpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/x-png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/bmp":
		return ".bmp"
	case "image/x-icon":
		return ".ico"
	case "application/zip":
		return ".zip"
	case "video/avi":
		return ".avi"
	case "application/vnd.rn-realmedia-vbr":
		return ".rmvb"
	case "audio/mpeg":
		return ".mp3"
	case "audio/wav":
		return ".wav"
	case "text/plain":
		return ".txt"
	case "application/msword":
		return ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx"
	case "application/vnd.ms-excel":
		return ".xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return ".xlsx"
	case "application/vnd.ms-powerpoint":
		return ".ppt"
	case "application/pdf":
		return ".pdf"
	case "video/mp4":
		return ".mp4"
	case "font/ttf":
		return ".ttf"
	case "text/csv":
		return ".csv"
	}
	return tp
}

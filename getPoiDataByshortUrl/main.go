package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type PoiDetailResponse struct {
	StatusCode int `json:"status_code"`
	PoiInfo    struct {
		PoiID        string  `json:"poi_id"`
		PoiName      string  `json:"poi_name"`
		PoiLongitude float64 `json:"poi_longitude"`
		PoiLatitude  float64 `json:"poi_latitude"`
		AddressInfo  struct {
			Province   string `json:"province"`
			City       string `json:"city"`
			District   string `json:"district"`
			Address    string `json:"address"`
			SimpleAddr string `json:"simple_addr"`
		} `json:"address_info"`
	} `json:"poi_info"`
}

func main() {
	shortUrl := "https://v.douyin.com/jSthVy0Jn-g/ 1@0.com"

	poiId, err := extractPoiIDFromShortUrl(shortUrl)
	if err != nil {
		fmt.Printf("获取POI ID失败: %v\n", err)
		return
	}
	fmt.Printf("成功获取POI ID: %s\n", poiId)

	poiDetail, err := getPoiDetail(poiId)
	if err != nil {
		fmt.Printf("获取POI详情失败: %v\n", err)
		return
	}

	fmt.Println("\n=== POI 详细信息 ===")
	fmt.Printf("名称: %s\n", poiDetail.PoiInfo.PoiName)
	fmt.Printf("地址: %s\n", poiDetail.PoiInfo.AddressInfo.Address)
	fmt.Printf("省份: %s\n", poiDetail.PoiInfo.AddressInfo.Province)
	fmt.Printf("城市: %s\n", poiDetail.PoiInfo.AddressInfo.City)
	fmt.Printf("经纬度: 经度 %.6f, 纬度 %.6f\n",
		poiDetail.PoiInfo.PoiLongitude,
		poiDetail.PoiInfo.PoiLatitude)
}

func extractPoiIDFromShortUrl(rawUrl string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	cleanUrl := strings.Split(rawUrl, " ")[0]

	resp, err := client.Head(cleanUrl)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location, err := resp.Location()
		if err != nil {
			return "", fmt.Errorf("获取重定向url失败: %v", err)
		}
		return extractPoiID(location.String())
	}

	return "", fmt.Errorf("未发生重定向，状态码: %d", resp.StatusCode)
}

func extractPoiID(finalUrl string) (string, error) {
	u, err := url.Parse(finalUrl)
	if err != nil {
		return "", fmt.Errorf("url解析失败: %v", err)
	}

	query := u.Query()
	if poiId := query.Get("poi_id"); poiId != "" {
		return poiId, nil
	}

	re := regexp.MustCompile(`poi_id=([^&]+)`)
	matches := re.FindStringSubmatch(finalUrl)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("未找到poi_id参数")
}

func getPoiDetail(poiId string) (*PoiDetailResponse, error) {
	apiUrl := fmt.Sprintf("https://www.iesdouyin.com/web/api/v2/poi/detail/?poi_id=%s", poiId)

	resp, err := http.Get(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("返回错误码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	var result PoiDetailResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("json解析失败: %v", err)
	}

	return &result, nil
}

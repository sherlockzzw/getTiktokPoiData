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
	shortURL := "https://v.douyin.com/jSthVy0Jn-g/ 1@0.com"

	poiID, err := extractPoiIDFromShortURL(shortURL)
	if err != nil {
		fmt.Printf("获取POI ID失败: %v\n", err)
		return
	}
	fmt.Printf("成功获取POI ID: %s\n", poiID)

	poiDetail, err := getPoiDetail(poiID)
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

func extractPoiIDFromShortURL(rawURL string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	cleanURL := strings.Split(rawURL, " ")[0]

	resp, err := client.Head(cleanURL)
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

func extractPoiID(finalURL string) (string, error) {
	u, err := url.Parse(finalURL)
	if err != nil {
		return "", fmt.Errorf("url解析失败: %v", err)
	}

	query := u.Query()
	if poiID := query.Get("poi_id"); poiID != "" {
		return poiID, nil
	}

	re := regexp.MustCompile(`poi_id=([^&]+)`)
	matches := re.FindStringSubmatch(finalURL)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("未找到poi_id参数")
}

func getPoiDetail(poiID string) (*PoiDetailResponse, error) {
	apiURL := fmt.Sprintf("https://www.iesdouyin.com/web/api/v2/poi/detail/?poi_id=%s", poiID)

	resp, err := http.Get(apiURL)
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

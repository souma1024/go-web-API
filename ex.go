package main

import (
	"io"
	"net/http"
	"fmt"
	"encoding/json"
	"strconv"
	"strings"
)

type ResponseParams struct {
	ObjectID int `json:"objectID"`
	ArtistDisplayName string `json:"artistDisplayName"`
	Title string `json:"title"`
	ArtistNationality string `json:"artistNationality"`
	PrimaryImage string `json:"primaryImage"`
}

var response []ResponseParams

func getQuery(r *http.Request) (int, string, string) {
	query := r.URL.Query()
	country := query.Get("c")
	artist := query.Get("n")
	page, err := strconv.Atoi(query.Get("p")) //strconv.Atoiで文字列数字を数字（int）に変換 数字にできなかったらエラー
	if err != nil {
		fmt.Println("Error:", err)
		page = 0
	}
	return page, country, artist
} 

func sendRequest() {
	for i := 1; i < 4000; i++ {
		url := fmt.Sprintf("https://collectionapi.metmuseum.org/public/collection/v1/objects/%d", i)
		res, err := http.Get(url)
		if err != nil {
			fmt.Println("Error", err)
			continue
		}

		if res.StatusCode != 200 {
			fmt.Println("Error: status code", res.StatusCode)
			continue
		}

		body, _ := io.ReadAll(res.Body) //ボディ部を読み込んでバイト列として取得
		res.Body.Close()                //httpレスポンスのボディ部を閉じている、読み取り終わったら閉じないといけないらしい

		var single ResponseParams
		if err := json.Unmarshal(body, &single); err != nil {  //ResponseParams構造体に変換
			fmt.Println("Error", err)						   //Unmarshalはフィールド名と一致したものを自動的に格納
			continue
		}
		fmt.Println(i)
		response = append(response, single) //スライス（動的配列）に要素を追加
	}
}

func doFilter(page int, country string, artist string) []ResponseParams {
	var filterResponse []ResponseParams
	var isCountry bool = false
	var isArtist bool = false

	if country != "" {
		isCountry = true
	}
	if artist != "" {
		isArtist = true
	}

	if isCountry && isArtist {
		for i := 0; i < len(response); i++ {
			if strings.Contains(response[i].ArtistDisplayName, artist) && strings.Contains(response[i].ArtistNationality, country){
				filterResponse = append(filterResponse, response[i])
			}
		}
	}

	if isCountry && !isArtist {
		for i := 0; i < len(response); i++ {
			if strings.Contains(response[i].ArtistNationality, country){
				filterResponse = append(filterResponse, response[i])
			}
		}
	}

	if !isCountry && isArtist {
		for i := 0; i < len(response); i++ {
				if strings.Contains(response[i].ArtistDisplayName, artist){
					filterResponse = append(filterResponse, response[i])
				}
			}
	}

	if !isCountry && !isArtist {
		for i := 0; i < len(response); i++ {
			filterResponse = append(filterResponse, response[i])
		}
	}
	
	start := page * 20
	end := start + 20
	if start >= len(filterResponse) {
		return []ResponseParams{}
	}
	if end > len(filterResponse) {
		end = len(filterResponse)
	}
	return filterResponse[start:end]
}

func imagesHandler(w http.ResponseWriter, r *http.Request) {
	page, country, artist := getQuery(r)
	var imagesResponse []ResponseParams

	imagesResponse = doFilter(page, country, artist)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(imagesResponse)  //構造体をJSON形式に変換
}

func countriesHandler(w http.ResponseWriter, r *http.Request){
	page, country, artist := getQuery(r)
	country = r.PathValue("countries")
	var countriesResponse []ResponseParams
	
	countriesResponse = doFilter(page, country, artist)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(countriesResponse)	
}

func artistHandler(w http.ResponseWriter, r *http.Request) {
	page, country, artist := getQuery(r)
	artist = r.PathValue("artist")
	var countriesResponse []ResponseParams
	
	countriesResponse = doFilter(page, country, artist)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(countriesResponse)	
}

func main() {
	fmt.Println("starting server")
	go sendRequest()

	http.HandleFunc("/images", imagesHandler)
	http.HandleFunc("/images/countries/{countries}", countriesHandler)
	http.HandleFunc("/images/artists/{artist}", artistHandler)

	http.ListenAndServe(":8000", nil)
}
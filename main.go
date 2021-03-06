// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	//SUrl lib
	"net/url"
	"io/ioutil"
	
	"os"
	"strconv"
	"strings"
	//"sort"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/JustinBeckwith/go-yelp/yelp"
)

var bot *linebot.Client
var richbot *linebot.RichMessageRequest
var o *yelp.AuthOptions

//shorten url
const (
	TINY_URL = 1
	IS_GD    = 2
)

type UrlShortener struct {
	ShortUrl    string
	OriginalUrl string
}

func getResponseData(urlOrig string) string {
	response, err := http.Get(urlOrig)
	if err != nil {
		fmt.Print(err)
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	return string(contents)
}

func tinyUrlShortener(urlOrig string) (string, string) {
	escapedUrl := url.QueryEscape(urlOrig)
	tinyUrl := fmt.Sprintf("http://tinyurl.com/api-create.php?url=%s", escapedUrl)
	return getResponseData(tinyUrl), urlOrig
}

func isGdShortener(urlOrig string) (string, string) {
	escapedUrl := url.QueryEscape(urlOrig)
	isGdUrl := fmt.Sprintf("http://is.gd/create.php?url=%s&format=simple", escapedUrl)
	return getResponseData(isGdUrl), urlOrig
}

func (u *UrlShortener) short(urlOrig string, shortener int) *UrlShortener {
	switch shortener {
	case TINY_URL:
		shortUrl, originalUrl := tinyUrlShortener(urlOrig)
		u.ShortUrl = shortUrl
		u.OriginalUrl = originalUrl
		return u
	case IS_GD:
		shortUrl, originalUrl := isGdShortener(urlOrig)
		u.ShortUrl = shortUrl
		u.OriginalUrl = originalUrl
		return u
	}
	return u
}
//surl

func main() {

        //要存入cookie的map
your := map[string]string{}
your["isuser"] = "isuser"
your["username"] = "username"
your["password"] = "password"
//将map转成json  转换后的是[]byte，需要string(your_byte)后就是json了
your_byte, _ := json.Marshal(your)
//将json base64一下
b64 := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")
your_string := b64.EncodeToString(your_byte)
fmt.Println(your_string)
//第二种base64的方式
your_string = base64.StdEncoding.EncodeToString(your_byte)
fmt.Println(your_string)
//存储cookie
this.Ctx.SetCookie("your", your_string)


	strID := os.Getenv("ChannelID")
	numID, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		log.Fatal("Wrong environment setting about ChannelID")
	}

	bot, err = linebot.NewClient(numID, os.Getenv("ChannelSecret"), os.Getenv("MID"))
	log.Println("Bot:", bot, " err:", err)

	// check environment variables
	o = &yelp.AuthOptions{
		ConsumerKey:       os.Getenv("ConsumerKey"),
		ConsumerSecret:    os.Getenv("ConsumerSecret"),
		AccessToken:       os.Getenv("Token"),
		AccessTokenSecret: os.Getenv("TokenSecret"),
	}

	if o.ConsumerKey == "" || o.ConsumerSecret == "" || o.AccessToken == "" || o.AccessTokenSecret == "" {
		log.Println("Wrong environment setting about yelp-api-keys")
	}

	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	received, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, result := range received.Results {
		content := result.Content()

		if content != nil && content.IsOperation && content.OpType == 4{
			_, err := bot.SendText([]string{result.RawContent.Params[0]}, "Hi～\n歡迎加入 LINE Delicious！\n請輸入'食物 地區' 查詢想吃的美食\n例如:\n義大利麵 新北市新莊區")
			_, err = bot.SendSticker([]string{result.RawContent.Params[0]}, 11, 1, 100)
			if err != nil {
				log.Println("New friend add event.")
			}
		}
		
		if content != nil && content.IsMessage && content.ContentType == linebot.ContentTypeText {
			text, err := content.TextContent()
			c := strings.Split(text.Text, " ")
			// create a new yelp client with the auth keys
			client := yelp.New(o, nil)
			if len(c) == 2{
				// make a simple query
				results, err := client.DoSimpleSearch(c[0], c[1])
				if err != nil {
					log.Println(err)
				}
				/*
				type ByRating []results.Businesses

				func (a ByRating) Len() int           { return len(a) }
				func (a ByRating) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
				func (a ByRating) Less(i, j int) bool { return a[i].Rating < a[j].Rating }
				
				func main() {
					people := []Person{
						{"Bob", 31},
						{"John", 42},
						{"Michael", 17},
						{"Jenny", 26},
					}
				
					fmt.Println(people)
					sort.Sort(ByRating(people))
					fmt.Println(people)
				
				}
				*/
				
				for i := 0; i < 3; i++ {
					imgurl := results.Businesses[i].ImageURL
					weburl := results.Businesses[i].URL
					
					urlOrig := UrlShortener{}
					urlOrig.short(weburl, IS_GD)
					/*
					if strings.HasPrefix(imgurl, "https"){
						imgurl = strings.Replace(imgurl, "https", "http", 1)
					}
					*/
					address := strings.Join(results.Businesses[i].Location.DisplayAddress,",")
					_, err = bot.SendImage([]string{content.From}, imgurl, imgurl)
					_, err = bot.SendText([]string{content.From}, "縮網址: " + urlOrig.ShortUrl)
					bot.SendText([]string{content.From}, "你的位置: " + content.RawContent.Text)
					imgurl = "http://i.imgur.com/lVM92n5.jpg"
					bot.NewRichMessage(1040).
						SetAction("food", "food", weburl).
						SetListener("food", 0, 0, 1040, 1040).
						Send([]string{content.From}, imgurl, "imagURLtest")
					_, err = bot.SendText([]string{content.From}, "店名: " + results.Businesses[i].Name + "\n電話: " + results.Businesses[i].Phone + "\n評比: " + strconv.FormatFloat(float64(results.Businesses[i].Rating), 'f', 1, 64))
					_, err = bot.SendLocation([]string{content.From}, results.Businesses[i].Name, address, float64(results.Businesses[i].Location.Coordinate.Latitude), float64(results.Businesses[i].Location.Coordinate.Longitude))
				}
			}else{
				_, err = bot.NewMultipleMessage().
				AddText("輸入格式錯誤, 請確認").
				AddSticker(1, 1, 100).
				Send([]string{content.From})
			}
			if err != nil {
				log.Println("OK")
			}

			_, err = bot.SendText([]string{content.From}, "Hi～\n歡迎加入 LINE Delicious！\n請輸入'食物 地區' 查詢想吃的美食\n例如:\n義大利麵 新北市新莊區")
			_, err = bot.SendSticker([]string{content.From}, 11, 1, 100)
			if err != nil {
				log.Println("wait for new message")
			}
		}
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

var Db *gorm.DB
var Err error

const URL = "http://api.openweathermap.org/data/2.5/forecast?id=703448&APPID=d8b90697a8eb8e570a3e526e813307c0"

var HostToPost string

type WeatherForecastData struct {
	gorm.Model
	HumanTime          string
	WeatherDescription string
	UnixTimestamp      int64
	Temperature        float64
	Pressure           float32
	Cloud              int
	Wind               float32
	WindDirection      int
	Sunset             int32
	Sunrize            int32
}

func main() {

	err := godotenv.Load("parameters.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	HostToPost = os.Getenv("HOST_TO_POST")
	fmt.Println(HostToPost)

	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	DoEvery(1*time.Hour, GetWeatherForecast)
}

func GetWeatherForecast(t time.Time) {
	fmt.Printf("%v: tick\n", t)

	type sys struct {
		Pod string `json:"pod"`
	}
	type coord struct {
		Lat float32 `json:"lat"`
		Lon float32 `json:"lon"`
	}
	type city struct {
		Id       int    `json:"id"`
		Name     string `json:"name"`
		Coord    coord  `json:"coord"`
		Country  string `json:"country"`
		Timezone int    `json:"timezone"`
		Sunrise  int32  `json:"sunrise"`
		Sunset   int32  `json:"sunset"`
	}
	type wind struct {
		Speed float32 `json:"speed"`
		Deg   int     `json:"deg"`
	}
	type clouds struct {
		All int `json:"all"`
	}
	type weather struct {
		Id          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	}
	type main struct {
		Temp       float64 `json:"temp"`
		Feels_like float32 `json:"feels_like"`
		Temp_min   float32 `json:"temp_min"`
		Temp_max   float32 `json:"temp_max"`
		Pressure   float32 `json:"pressure"`
		Sea_level  int     `json:"sea_level"`
		Grnd_level int     `json:"grnd_level"`
		Humidity   int     `json:"humidity"`
		Temp_kf    float32 `json:"temp_kf"`
	}
	type OneHour struct {
		Dt      int64     `json:"dt"`
		Main    main      `json:"main"`
		Weather []weather `json:"weather"`
		Clouds  clouds    `json:"clouds"`
		Wind    wind      `json:"wind"`
		Sys     sys       `json:"sys"`
		Dt_txt  string    `json:"dt_txt"`
	}
	type jsonPocket struct {
		Cod     string    `json:"cod"`
		Message int       `json:"message"`
		Cnt     int       `json:"cnt"`
		List    []OneHour `json:"list"`
		City    city      `json:"city"`
	}

	resp, err := http.Get(URL)
	if err != nil {
		log.Fatalln(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return
	}

	log.Println(string(body))

	var response jsonPocket
	jsonErr := json.Unmarshal(body, &response)
	if jsonErr != nil {
		log.Fatal(jsonErr)
		return
	}

	log.Println("\n")
	log.Println(response)

	for ind, val := range response.List {
		log.Println(ind)
		log.Println(val)

		tm := time.Unix(val.Dt, 0)
		log.Println(tm)

		var wfd WeatherForecastData

		wfd.HumanTime = time.Unix(val.Dt, 0).Format("2006-01-02 15:04:05")
		wfd.WeatherDescription = val.Weather[0].Description
		wfd.UnixTimestamp = val.Dt
		wfd.Temperature = math.Round((val.Main.Temp-273.15)*100) / 100
		wfd.Pressure = val.Main.Pressure
		wfd.Cloud = val.Clouds.All
		wfd.Wind = val.Wind.Speed
		wfd.WindDirection = val.Wind.Deg
		wfd.Sunset = response.City.Sunset
		wfd.Sunrize = response.City.Sunrise

		fmt.Println("URL:>", HostToPost)

		jsonStr, _ := json.Marshal(wfd)
		//var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
		req, _ := http.NewRequest("POST", HostToPost, bytes.NewBuffer(jsonStr))
		req.Header.Set("X-Custom-Header", "myvalue")
		req.Header.Set("Content-Type", "application/json")

		time.Sleep(1 * time.Second)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
		} else {
			defer resp.Body.Close()

			log.Println("response Status:", resp.Status)
			log.Println("response Headers:", resp.Header)
			body, _ := ioutil.ReadAll(resp.Body)
			log.Println("response Body:", string(body))
		}
	}
	fmt.Printf("%v: tack\n", t)
}

func DoEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

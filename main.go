package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

/*const (
	WEATHER_URL         = "https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric&lang=ru"
	TG_BOT_LINK         = "https://t.me/weather_zx456_bot"
)*/

func main() {
	err := godotenv.Load("properties.env") //тут файл с настройками, шаблон для заполнения properties_example.env
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Создаём новый экземпляр бота, используя токен доступа из переменной TG_BOT_ACCESS_TOKEN
	tqBotToken := os.Getenv("TG_BOT_ACCESS_TOKEN")
	bot, err := tgbotapi.NewBotAPI(tqBotToken)

	// Проверяем, не возникла ли ошибка при создании бота
	if err != nil {
		// Если ошибка есть - выводим её и останавливаем программу
		log.Panic(err)
	}

	// Включаем режим отладки (в консоль будут выводиться все запросы/ответы Telegram API)
	bot.Debug = true

	// Выводим в лог информацию об авторизации и имя бота
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Создаём конфигурацию для получения обновлений, начиная с offset = 0
	u := tgbotapi.NewUpdate(0)

	// Устанавливаем таймаут long polling в 60 секунд
	u.Timeout = 60

	// Получаем канал (channel), через который будут приходить обновления от Telegram
	updates := bot.GetUpdatesChan(u)

	// Запускаем бесконечный цикл, который читает обновления из канала
	for update := range updates {
		// Проверяем, есть ли в обновлении сообщение (не nil)
		if update.Message != nil {
			/*// Логируем имя пользователя и текст сообщения
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			// Создаём новое сообщение для отправки в тот же чат, откуда пришло сообщение
			// В качестве текста используем текст полученного сообщения (эхо)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text+"привет Хрюн")

			// Устанавливаем, что наше сообщение будет ответом на полученное сообщение
			msg.ReplyToMessageID = update.Message.MessageID

			// Отправляем сообщение пользователю
			bot.Send(msg)*/

			if update.Message.Command() == "start" {
				msg := tgbotapi.NewMessage(
					update.Message.Chat.ID,
					"Привет! Отправьте геолокацию или координаты (например: 55.7558 37.6173) для получения погоды!",
				)
				bot.Send(msg)
				continue
			}

			// Если пользователь отправил геолокацию
			if update.Message.Location != nil {
				lat := update.Message.Location.Latitude  // float64
				lon := update.Message.Location.Longitude // float64
				weather, we := getWeather(lat, lon)
				if we != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, weather)
					bot.Send(msg)
					continue
				}

				// Используем координаты
				msg := tgbotapi.NewMessage(
					update.Message.Chat.ID,
					weather,
				)
				bot.Send(msg)
			} else {
				if update.Message.Text != "" {
					lat, lon, err := parseCoordinates(update.Message.Text)
					if err != nil {
						msg := tgbotapi.NewMessage(
							update.Message.Chat.ID,
							fmt.Sprintf("Не удалось распознать координаты:\n%v\n\nОтправьте в виде: 55.7558 37.6173", err),
						)
						bot.Send(msg)
						continue
					}
					weather, we := getWeather(lat, lon)
					if we != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, weather)
						bot.Send(msg)
						continue
					}

					// Используем координаты
					msg := tgbotapi.NewMessage(
						update.Message.Chat.ID,
						weather,
					)
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(
						update.Message.Chat.ID,
						"Геолокация не распознана, отправьте геолокацию или координаты в виде 55.7558 37.6173 или нажмите на скрепку",
					)
					bot.Send(msg)
				}
			}

		}
	}
}

type WeatherResponse struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`

	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`

	Base string `json:"base"`

	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
		SeaLevel  int     `json:"sea_level"`
		GrndLevel int     `json:"grnd_level"`
	} `json:"main"`

	Visibility int `json:"visibility"`

	Wind struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`

	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`

	Dt int64 `json:"dt"`

	Sys struct {
		Type    int    `json:"type"`
		ID      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int64  `json:"sunrise"`
		Sunset  int64  `json:"sunset"`
	} `json:"sys"`

	Timezone int    `json:"timezone"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

func main1() {
	fmt.Println("Weather Bot запущен!")
	weather, _ := getWeather(55.7558, 37.6173)
	fmt.Println(weather)

}

func parseCoordinates(s string) (lat, lon float64, err error) {
	_, err = fmt.Sscanf(s, "%f %f", &lat, &lon)
	return
}

func getWeather(lat, lon float64) (weather string, err error) {
	fmt.Printf("Getting weather for %f,%f\n", lat, lon)
	weatherApiUrl := os.Getenv("WEATHER_URL")
	weatherApiKey := os.Getenv("WEATHER_API_KEY")
	url := fmt.Sprintf(weatherApiUrl, lat, lon, weatherApiKey)

	resp, err := http.Get(url)
	if err != nil {
		return "Ошибка при запросе погоды", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Ошибка при чтении ответа", err
	}
	var weatherResp WeatherResponse
	errOfMarsh := json.Unmarshal(body, &weatherResp)
	if errOfMarsh != nil {
		return "Ошибка декодирования данных", errOfMarsh
	}

	fmt.Println("Город:", weatherResp.Name)
	fmt.Println("Температура:", weatherResp.Main.Temp, "°C")
	fmt.Println("Описание:", weatherResp.Weather[0].Description)
	result := fmt.Sprintf(
		"🌍 Место: %s\n"+
			"🌡 Температура: %.1f°C\n"+
			"🤔 Ощущается как: %.1f°C\n"+
			"☁️ Погода: %s (%s)\n"+
			"💧 Влажность: %d%%\n"+
			"💨 Ветер: %.1f м/с\n"+
			"📍 Координаты: %.4f, %.4f\n",
		weatherResp.Name,
		weatherResp.Main.Temp,
		weatherResp.Main.FeelsLike,
		weatherResp.Weather[0].Main,
		weatherResp.Weather[0].Description,
		weatherResp.Main.Humidity,
		weatherResp.Wind.Speed,
		lat, lon,
	)
	return result, nil
}

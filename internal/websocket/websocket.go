package websocket

import (
	"encoding/json"                // Пакет для работы с JSON
	"fmt"                          // Пакет для форматирования строк
	"github.com/gorilla/websocket" // Пакет для работы с WebSocket
	"log"                          // Пакет для логирования
	"math/rand"                    // Пакет для генерации случайных чисел
	"net/http"                     // Пакет для работы с HTTP
	"sync"                         // Пакет для синхронизации
)

// Создаем объект upgrader для улучшения соединения до WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,                                       // Размер буфера для чтения сообщений
	WriteBufferSize: 1024,                                       // Размер буфера для записи сообщений
	CheckOrigin:     func(r *http.Request) bool { return true }, // Разрешаем все источники
}

// Структура для хранения результата игры
type Result struct {
	Clicks int // Количество кликов
	Time   int // Время игры
}

// Хранилище лучших результатов, где ключом является путь, а значением результат
var results = make(map[string]Result)

// Мьютекс для синхронизации доступа к хранилищу результатов
var resultsMutex sync.Mutex

// Структура для получения сообщения от клиента
type ClientMessage struct {
	State string          `json:"state"` // Состояние (start, finish и т.д.)
	Data  json.RawMessage `json:"date"`  // Данные в виде необработанного JSON
}

// Структура данных для начала игры
type StartStateData struct {
	Path string `json:"path"` // Путь, используемый для генерации карточек
}

// Структура данных для завершения игры
type FinishStateData struct {
	Clicks int `json:"clicks"` // Количество кликов в завершенной игре
	Time   int `json:"time"`   // Время в завершенной игре
}

// Структура данных для карточек
type CardVariationData struct {
	Count     int   `json:"count"`     // Количество карточек
	Variation []int `json:"variation"` // Список карточек
}

// Структура для сообщения о результате
type ResultStateData struct {
	Message string `json:"message"` // Сообщение о результате
}

// Обрабатываем подключения WebSocket
func HandleConnections(w http.ResponseWriter, r *http.Request) {
	// Обновляем соединение до WebSocket, responseHeader?
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error while upgrading connection:", err)
		return
	}
	defer ws.Close()

	var gamePath string // Переменная для хранения пути игры

	for {
		var msg ClientMessage
		// Читаем сообщение от клиента и распаковываем его в переменную msg, структура ClientMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error while reading message: %v", err)
			break
		}

		log.Printf("Received message: %s", msg.State)
		// Обрабатываем сообщение в зависимости от состояния
		switch msg.State {
		case "start":
			// Если состояние "start", начинаем игру
			var data StartStateData
			// Разбираем данные из JSON и заполняем структуру data, структура StartStateData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				log.Printf("Error unmarshalling start data: %v", err)
				break
			}

			gamePath = data.Path // Сохраняем путь игры для дальнейшего использования, структура StartStateData, PATH

			// Генерируем набор карточек на основе PATH
			variation := generateCardVariation(data.Path) // variation = [] int

			// Создаем данные для отправки клиенту
			cardData := CardVariationData{
				Count:     len(variation), // Количество карточек, 6, 8, 12
				Variation: variation,      // Список карточек []int`json:"variation"
			}
			// Отправляем набор карточек обратно клиенту
			sendResponse(ws, "set", cardData) // Соединение, состояние, набор карточек

		case "finish":
			var data FinishStateData
			// Разбираем данные из JSON и заполняем структуру data, структура FinishStateData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				log.Printf("Error unmarshalling finish data: %v", err)
				break
			}

			// Используем сохраненный путь для получения и обновления лучшего результата
			bestResult := updateBestResult(gamePath, data.Clicks, data.Time)
			/*bestResult = "/level1",
			Click,
			Time
			*/

			// Формируем сообщение о результате
			resultMessage := fmt.Sprintf("best result is %d and your result is %d", bestResult, data.Clicks)

			/*func Sprintf(format string, a ...any) string
			return string
			*/

			// Отправляем сообщение о результате обратно клиенту
			resultScore := ResultStateData{Message: resultMessage}

			sendResponse(ws, "result", resultScore)
		}
	}
}

// Функция отправки ответа клиенту через WebSocket

func sendResponse(ws *websocket.Conn, state string, data interface{}) {
	// responseData = []byte, err = _
	responseData, _ := json.Marshal(data)
	// Создаине сообщения  для отправки
	response := ClientMessage{
		State: state,                         // Состояние ответа
		Data:  json.RawMessage(responseData), // Данные ответа не надо Марщшалить
	}
	/*	response = {
			"state": "new",
		Data =
			"date": {
			"count": 6,
				"variation": [1, 2, 3, 3, 2, 1]
		}
		}
	*/

	// Отправляем сообщение клиенту
	if err := ws.WriteJSON(response); err != nil {
		// Если произошла ошибка при отправке сообщения, логируем её
		log.Printf("Error while writing message: %v", err)
	}
}

// Функция преобразования строки path в seed для генератора случайных чисел
func StringToInt64(line string) int64 {
	runes := []rune(line)    // Преобразуем строку в массив рун
	totalRunes := len(runes) // Получаем количество рун в строке
	var result int64         // Переменная для хранения результата

	for _, r := range runes {
		// Для каждой руны вычисляем значение и добавляем к результату
		result += int64(int(r) % totalRunes)
	}
	return result // Возвращаем результат
}

// Генерация набора карточек на основе пути
func generateCardVariation(path string) []int {
	seed := StringToInt64(path)         // Преобразуем путь в seed
	r := rand.New(rand.NewSource(seed)) // Создаем локальный генератор случайных чисел

	countVar := []int{6, 8, 12}              // Возможные количества карточек
	count := countVar[r.Intn(len(countVar))] // Выбираем случайное количество карточек
	const maxCard = 5                        // Максимальное значение карточки
	var variation []int                      // Переменная для хранения карточек

	if count >= 2 {
		for i := 0; i < count/2; i++ {
			// Добавляем карточки в массив
			variation = append(variation, r.Intn(maxCard))
		}
	}

	// Дублируем массив карточек
	variation = append(variation, variation...)

	// Перемешиваем карточки с использованием алгоритма Фишера-Йетса
	for i := len(variation) - 1; i > 0; i-- {
		n := r.Intn(i + 1)                                      // Выбираем случайный индекс для обмена
		variation[i], variation[n] = variation[n], variation[i] // Обмен карточек
	}

	return variation // Возвращаем перемешанный набор карточек
}

// Функция обновления и получения лучшего результата
func updateBestResult(path string, clicks int, time int) int {
	resultsMutex.Lock()
	defer resultsMutex.Unlock()

	log.Printf("Before update: Current results map: %+v", results)

	bestResult, exists := results[path] // Проверяем, существует ли уже результат для этого пути
	/*	results = map[string]Result{
		"/level3": {Clicks: 5,
					Time: 100},
	}*/

	if exists {
		// Уже есть результат для данного пути
		log.Printf("Current best result for path '%s': Clicks=%d, Time=%d", path, bestResult.Clicks, bestResult.Time)
	} else {
		// Еси результат для этого пути не существует
		log.Printf("No result for path '%s', adding new one.", path)
	}

	// Входные параметры на новые результаты, если они есть
	log.Printf("New result for path '%s': Clicks=%d, Time=%d", path, clicks, time)

	if !exists || clicks < bestResult.Clicks { // Если результат не существует или новый результат лучше
		results[path] = Result{ // Обновляем результат
			Clicks: clicks,
			Time:   time,
			/*	path := "/level3"
				clicks := 4
				time := 90
			*/
		}

		// Обновленный результат
		log.Printf("Updated best result for path '%s': Clicks=%d, Time=%d", path, clicks, time)

		return clicks // Возвращаем новый лучший результат
	}

	// Текущий лучший результат остался прежним
	log.Printf("Best result remains unchanged for path '%s': Clicks=%d, Time=%d", path, bestResult.Clicks, bestResult.Time)

	return bestResult.Clicks // Возвращаем текущий лучший результат
}

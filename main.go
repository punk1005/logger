package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"fmt"
	"bytes"
	"encoding/json"

)

const logsDir = "logs"

func main() {
	// 1. Проверяем наличие папки logs, если её нет — создаем
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		err := os.Mkdir(logsDir, 0755)
		if err != nil {
			log.Fatalf("Критическая ошибка: не удалось создать папку %s: %v", logsDir, err)
		}
		log.Printf("Папка '%s' успешно создана", logsDir)
	}

	// 2. Регистрируем универсальный обработчик для всех путей
	http.HandleFunc("/", logHandler)

	log.Println("Лог-сервер запущен на порту :7001...")
	if err := http.ListenAndServe(":7001", nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/favicon.ico" {
		return
	}


	logName := strings.TrimPrefix(r.URL.Path, "/")
	logName = filepath.Base(logName)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела запроса для %s: %v", logName, err)
		return
	}
	defer r.Body.Close()

	if len(bodyBytes) == 0 {
		return
	}

	filePath := filepath.Join(logsDir, logName+".txt")

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Ошибка открытия файла %s: %v", filePath, err)
		return
	}
	defer file.Close()

		// 1. Создаем буфер для форматированного JSON
		var prettyJSON bytes.Buffer

		// 2. Форматируем пришедшие байты. 
		// "" — префикс строки, "  " (два пробела) — отступ для каждого уровня вложенности
		err = json.Indent(&prettyJSON, bodyBytes, "", "  ")
		if err != nil {
			log.Printf("Ошибка форматирования JSON для %s: %v", logName, err)
			// Если JSON почему-то «битый», пишем как есть (старый вариант)
			prettyJSON.Write(bodyBytes) 
		}
	

	// 1. Формируем строку заголовка с текущей датой и временем сервера
	// Формат "2006-01-02 15:04:05" в Go — это эталон для маски YYYY-MM-DD HH:MM:SS
	currentTime := time.Now().Format("2006-01-02, время - 15:04:05")
	headerLine := fmt.Sprintf("/*****// Запрос /%s, Дата - %s\n", logName, currentTime)

	// 2. Склеиваем заголовок и само тело JSON, добавляя перенос строки в самый конец
	//payload := append([]byte(headerLine), bodyBytes...)
	//payload = append(payload, '\n')
		// 4. Собираем всё вместе и добавляем перенос строки в самом конце лога
	payload := append([]byte(headerLine), prettyJSON.Bytes()...)
	payload = append(payload, '\n', '\n')


	// 3. Записываем всё одним махом
	_, err = file.Write(payload)
	if err != nil {
		log.Printf("Ошибка записи в файл %s: %v", filePath, err)
		return
	}

	log.Printf("Данные успешно записаны в лог: %s", filePath)
}

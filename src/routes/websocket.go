package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func RoutesWebSocket() {
	http.HandleFunc("/chat", wsEndpoint)
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error al actualizar la conexión WebSocket: ", err)
		return
	}

	log.Println("Client Successfully Connected...")

	reader(ws)
}

func reader(conn *websocket.Conn) {
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		log.Println(string(p))

		response, err := getOpenAIResponse(string(p))
		if err != nil {
			log.Println("Error al obtener respuesta de OpenAI:", err)
			openAIErrorMessage := "Hubo un error al generar la respuesta utilizando OpenAI."
			if writeErr := conn.WriteMessage(websocket.TextMessage, []byte(openAIErrorMessage)); writeErr != nil {
				log.Println("Error al enviar el mensaje de error de OpenAI al cliente:", writeErr)
			}
			continue
		}

		log.Println("Respuesta de OpenAI generada:", response)

		if err := conn.WriteMessage(messageType, []byte(response)); err != nil {
			log.Println(err)
			return
		}
	}
}

func getOpenAIResponse(message string) (string, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error configuración de variables de entorno")
	}
	openAIKey := os.Getenv("ENV_APIKEY_OPENAI")

	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo", // Modelo de OpenAI a utilizar
		"messages": []Message{
			{Role: "system", Content: "Estás asistiendo a un cliente de Capital Tours, una empresa de taxis en Huancayo, Perú. Este asistente puede proporcionar información sobre los servicios y requisitos de Capital Tours; No tienes permitido abordar otros temas que no sean referente a la empresa"},
			{Role: "assistant", Content: "¡Hola! Estoy aquí para ayudarte con información sobre Capital Tours. Somos una empresa de taxis con sede en Huancayo, Perú. Ofrecemos servicios de suscripción de taxis con requisitos como una copia de tu DNI, registro del vehículo, revisión técnica y una copia de SOAT o AFOCAT."},
			{Role: "assistant", Content: "El costo de unirse es de 150 Soles, que incluye pegatinas del logotipo, tarjeta de circulación y un mes de cuota. Después de eso, el pago mensual es de 25 Soles. También ofrecemos servicios de AFOCAT o SOAT para vehículos públicos por 105 Soles, con requisitos de copia de DNI y tarjeta de circulación."},
			{Role: "assistant", Content: "Para más detalles y solicitudes, puedes visitar: https://www.platcont.lat/solicitar"},
			{Role: "assistant", Content: "No dudes en preguntar si tienes más preguntas. ¡Estoy aquí para ayudarte!"},
			{Role: "assistant", Content: "No tengo información de ese tema, si deseas estaré para ayudarte con tus dudas en Capital Tours."},
			{Role: "user", Content: "Por favor responde de manera concisa"},
			{Role: "user", Content: "Por favor no abordes otros temas que no se relacionen con la empresa Capital Tours"},
			{Role: "user", Content: message},
		},
		"max_tokens":        100, // Número máximo de tokens en la respuesta
		"temperature":       0,
		"top_p":             1,
		"frequency_penalty": 0.5,
		"presence_penalty":  0,
	}

	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(requestBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openAIKey) // Aquí usamos la clave de API de OpenAI

	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal(responseBytes, &responseData); err != nil {
		return "", err
	}

	log.Println(responseData)

	if output, ok := responseData["choices"].([]interface{}); ok && len(output) > 0 {
		if choice, ok := output[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if text, ok := message["content"].(string); ok {
					return text, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no se pudo obtener una respuesta de OpenAI")
}

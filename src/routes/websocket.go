package routes

import (
	"bytes"
	"chatbot/src/database/models/tables"
	"chatbot/src/database/orm"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

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

var clients = make(map[*websocket.Conn]bool) // Map para almacenar conexiones activas
var clientsMutex sync.Mutex

func RoutesWebSocket() {
	http.HandleFunc("/chat", wsChatbotEndpoint)
	http.HandleFunc("/location", wsLocationsEndpoint)
}

func wsChatbotEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error al actualizar la conexión WebSocket: ", err)
		return
	}

	log.Println("Client Successfully Connected...")

	readerIA(ws)
}

func readerIA(conn *websocket.Conn) {
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
			{Role: "assistant", Content: "¡Hola! Estoy aquí para ayudarte con información sobre Capital Tours. Somos una empresa de taxis con sede en Huancayo, Perú. Ofrecemos servicios de suscripción de taxis con requisitos como una copia de tu DNI, registro del vehículo, revisión técnica y una copia de SOAT o AFOCAT. Nos encontramos en Av. Mariscal Castilla & Av. Gral. Cordova Huancayo, Huancayo-Perú. Atendemos de 8:30 am hasta 7:00 pm"},
			{Role: "assistant", Content: "El costo de unirse como taxi es de 150 Soles, que incluye pegatinas del logotipo, tarjeta de circulación y un mes de cuota. Después de eso, el pago mensual es de 25 Soles. También ofrecemos servicios de AFOCAT o SOAT para vehículos públicos por 105 Soles, con requisitos de copia de DNI y tarjeta de circulación."},
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

func wsLocationsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	clientsMutex.Lock()
	clients[ws] = true
	clientsMutex.Unlock()

	log.Println("Locations connected")

	_locations := getLocations()

	bytesResultado, err := json.Marshal(_locations)
	if err != nil {
		log.Println("error al convertir")
	}

	err = ws.WriteMessage(1, []byte(bytesResultado))
	if err != nil {
		log.Println(err)
	}

	readerLocations(ws)
}

func readerLocations(conn *websocket.Conn) {
	defer func() {
		conn.Close()
		// Eliminar la conexión de la lista de clientes cuando se cierra
		clientsMutex.Lock()
		delete(clients, conn)
		clientsMutex.Unlock()
	}()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// TRAER DATA DE EL CLIENTE
		var data map[string]interface{}

		errData := json.Unmarshal(p, &data)
		if errData != nil {
			fmt.Println("Error al deserializar JSON:", err)
			return
		}

		var crudReturn []map[string]interface{}

		if data["typeSend"] == "insert" {
			crudReturn = append(
				crudReturn,
				insertLocation(data["data"].(map[string]interface{})),
			)
		}

		if data["typeSend"] == "delete" {
			crudReturn = append(
				crudReturn,
				deleteLocation(data["data"].(map[string]interface{})),
			)
		}

		fmt.Println(crudReturn)

		// GET LOCATIONS
		_locations := getLocations()
		bytesResultado, err := json.Marshal(_locations)
		if err != nil {
			log.Println("error al convertir")
		}

		// Enviar el mensaje a todos los clientes conectados
		clientsMutex.Lock()
		for client := range clients {
			err = client.WriteMessage(messageType, []byte(bytesResultado))
			if err != nil {
				log.Println(err)
			}
		}
		clientsMutex.Unlock()
	}
}

func getLocations() []map[string]interface{} {
	_data_solicitudes := orm.NewQuerys("locations").Select().Exec(orm.Config_Query{Cloud: true}).All()

	if len(_data_solicitudes) <= 0 {
		var response []map[string]interface{}
		response = append(response, map[string]interface{}{
			"msg": "error al obtener solicitudes",
		})
		return response
	}
	return _data_solicitudes
}

func insertLocation(data map[string]interface{}) map[string]interface{} {
	data_insert := append([]map[string]interface{}{}, data)

	schema, table := tables.Locations_GetSchema()
	solicitudes := orm.SqlExec{}
	err := solicitudes.New(data_insert, table).Insert(schema)
	if err != nil {
		return map[string]interface{}{
			"msg":   "Ocurrió un error al insertar Locacion",
			"error": err,
		}
	}

	err = solicitudes.Exec()
	if err != nil {
		return map[string]interface{}{
			"msg":   "Ocurrió un error al insertar Locacion",
			"error": err,
		}
	}

	return solicitudes.Data[0]
}

func deleteLocation(data map[string]interface{}) map[string]interface{} {
	var data_delete []map[string]interface{}

	data_delete = append(data_delete, map[string]interface{}{
		"id_location": data["id_location"],
	})

	schema, table := tables.Locations_GetSchema()
	solicitudes := orm.SqlExec{}
	err := solicitudes.New(data_delete, table).Delete(schema)
	if err != nil {
		return map[string]interface{}{
			"msg":   "Ocurrió un error al eliminar Locacion",
			"error": err,
		}
	}

	err = solicitudes.Exec()
	if err != nil {
		return map[string]interface{}{
			"msg":   "Ocurrió un error al eliminar Locacion",
			"error": err,
		}
	}

	return solicitudes.Data[0]
}

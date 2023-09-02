package library

import (
	"chatbot/src/auth"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Tipo de codigo: n° serie, n° comprobante
type CodigoType int

const (
	Comprobante CodigoType = iota
	Serie
	// otros tipos
)

type Codigo struct {
	Tipo  CodigoType
	Valor string
}

// end struct codigo_type

func GetTokenKey(r *http.Request, key string) interface{} {
	token := r.Header.Get("Access-Token")
	if token == "" {
		fmt.Println("Error al obtener información de la session")
		return nil
	}
	data, err := auth.ValidateToken(token)
	if err != nil {
		fmt.Println("Session a expirado")
		return nil
	}
	map_data := map[string]interface{}{
		"email": data.Email,
		"us":    data.IdUser,
	}
	value := map_data[key]
	if value == nil {
		fmt.Println("No se encontró la key de la session")
		return nil
	}
	return value
}

func InterfaceToString(params ...interface{}) string {
	typeValue := reflect.TypeOf(params[0]).String()
	value := params[0]
	valueReturn := ""
	if strings.Contains(typeValue, "string") {
		toSql := false
		if len(params) == 2 && reflect.TypeOf(params[1]).Kind() == reflect.Bool {
			toSql = params[1].(bool)
		}

		if toSql {
			valueReturn = fmt.Sprintf("'%s'", value)
		} else {
			valueReturn = fmt.Sprintf("%s", value)
		}
	} else if strings.Contains(typeValue, "int") {
		valueReturn = fmt.Sprintf("%d", value)
	} else if strings.Contains(typeValue, "float") {
		valueReturn = fmt.Sprintf("%f", value)
	} else if strings.Contains(typeValue, "bool") {
		valueReturn = fmt.Sprintf("%t", value)
	}
	return valueReturn
}

func IndexOf_String(arreglo []string, search string) int {
	for indice, valor := range arreglo {
		if valor == search {
			return indice
		}
	}
	// -1 porque no existe
	return -1
}

func IndexOf_String_Map(arreglo []map[string]interface{}, key, search string) int {
	for indice, valor := range arreglo {
		if valor[key] == search {
			return indice
		}
	}
	// -1 porque no existe
	return -1
}

func GenerateCodigo(input Codigo) (string, error) {
	len_input := len(input.Valor)
	input.Valor = strings.TrimLeft(input.Valor, "0")

	if input.Valor == "" {
		if input.Tipo == Comprobante {
			input.Valor = "0000000001"
			return input.Valor, nil
		} else if input.Tipo == Serie {
			input.Valor = "0001"
			return input.Valor, nil
		}
	}

	numero, err := strconv.Atoi(input.Valor)
	if err != nil {
		return "", err
	}

	numero++
	nuevoString := fmt.Sprintf("%0"+strconv.Itoa(len_input)+"d", numero)

	return nuevoString, nil
}

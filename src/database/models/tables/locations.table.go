package tables

import (
	"chatbot/src/database/models"

	"github.com/google/uuid"
)

func Locations_GetSchema() ([]models.Base, string) {
	var locations []models.Base
	tableName := "locations"
	locations = append(locations, models.Base{ //id_location
		Name:        "id_location",
		Description: "id_location",
		Default:     uuid.New().String(),
		Required:    true,
		Important:   true,
		Type:        "string",
		Strings:     models.Strings{},
	})
	locations = append(locations, models.Base{ //nombre
		Name:        "nombre",
		Description: "nombre",
		Required:    true,
		Update:      true,
		Type:        "string",
		Strings: models.Strings{
			Min: 5,
			Max: 150,
		},
	})
	locations = append(locations, models.Base{ //latitud
		Name:        "latitud",
		Description: "latitud",
		Required:    true,
		Update:      true,
		Type:        "float64",
		Float: models.Floats{
			Negativo: true,
		},
	})
	locations = append(locations, models.Base{ //longitud
		Name:        "longitud",
		Description: "longitud",
		Required:    true,
		Update:      true,
		Type:        "float64",
		Float: models.Floats{
			Negativo: true,
		},
	})
	locations = append(locations, models.Base{ //numero_placa
		Name:        "numero_placa",
		Description: "numero_placa",
		Required:    true,
		Update:      true,
		Type:        "string",
		Strings: models.Strings{
			Min:       7,
			Max:       7,
			UpperCase: true,
		},
	})
	locations = append(locations, models.Base{ //numero_flota
		Name:        "numero_flota",
		Description: "numero_flota",
		Required:    true,
		Update:      true,
		Type:        "uint64",
		Uint:        models.Uints{},
	})

	return locations, tableName
}

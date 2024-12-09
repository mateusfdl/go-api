package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	http_adapter "github.com/mateusfdl/go-api/adapters/http"
	"github.com/mateusfdl/go-api/adapters/logger"
	"github.com/mateusfdl/go-api/adapters/mongo"
	"github.com/mateusfdl/go-api/config"
	"github.com/mateusfdl/go-api/internal/crops"
	"github.com/mateusfdl/go-api/internal/farms"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	s  *http_adapter.HTTP
	db *mongo.Mongo
)

type FarmResponse struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Address           string `json:"address"`
	LandArea          int    `json:"landArea"`
	UnitOfMeasurement string `json:"unitOfMeasurement"`
	Crops             []struct {
		Type        string `json:"type"`
		IsIrrigated bool   `json:"isIrrigated"`
		IsInsured   bool   `json:"isInsured"`
	} `json:"crops"`
}

func TestFarm(t *testing.T) {
	if err := godotenv.Load(".env-test"); err != nil {
		panic(err)
	}

	ctx := context.Background()
	c, err := config.NewAppConfig()
	if err != nil {
		panic(err)
	}

	l := logger.New(&c.Logger)
	db = mongo.New(ctx, l, &c.Mongo)
	s = http_adapter.New(l, &c.HTTP)

	cropsModule := crops.New(db.DB)
	farmsModule := farms.New(l, &cropsModule.Repository, s, db.DB)

	mongo.HookOnStart(ctx, db, l)

	http_adapter.RegisterRoutes(farmsModule.Controller)

	defer func() {
		mongo.GracefulShutdown(ctx, db, l)
		s.GracefulShutdown(ctx)
	}()

	go s.Listen()

	t.Run("Create Farm", CreateFarm)
	t.Run("List Farms", ListFarms)
	t.Run("Get Farm", FarmGet)
	t.Run("Update Farm", FarmUpdate)
	t.Run("Delete Farm", FarmDelete)
}

func CreateFarm(t *testing.T) {
	t.Run("Create Farm Without Crops", CreateFarmWithoutCrops)
	t.Run("Create Farm With Crops", CreateFarmWithCrops)
	t.Run("Create Farm Compliance", CreateFarmComplianceFields)
}

func CreateFarmWithoutCrops(t *testing.T) {
	body := strings.NewReader(`{
    "name": "Farm 1",
    "landArea": 87,
    "unitOfMeasurement": "hectares",
    "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
    "crops": []
  }`)

	w := performRequest("POST", "/farms", body)
	assertStatusCode(t, w, http.StatusCreated)

	var farmResponse FarmResponse
	err := json.Unmarshal(w.Body.Bytes(), &farmResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal response body")
	}
	if farmResponse.ID == "" {
		t.Errorf("Expected id, but got empty")
	}
}

func CreateFarmWithCrops(t *testing.T) {
	b := strings.NewReader(`{
    "name": "Farm 1",
    "landArea": 29,
    "unitOfMeasurement": "hectares",
    "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
    "crops": [
        {
            "type": "CORN",
            "isIrrigated": true,
            "isInsured": true
        },
        {
            "type": "SOYBEANS",
            "isIrrigated": false,
            "isInsured": false
        },
        {
            "type": "RICE",
            "isIrrigated": true,
            "isInsured": false
        },
        {
            "type": "BEAN",
            "isIrrigated": false,
            "isInsured": true
        },
        {
            "type": "COFFEE",
            "isIrrigated": true,
            "isInsured": true
        }
    ]
  }`)
	w := performRequest("POST", "/farms", b)
	assertStatusCode(t, w, http.StatusCreated)

	var farmResponse FarmResponse
	err := json.Unmarshal(w.Body.Bytes(), &farmResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal response body")
	}
	if farmResponse.ID == "" {
		t.Errorf("Expected id, but got empty")
	}
}

func CreateFarmComplianceFields(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected int
	}{
		{
			name: "Missing Name",
			body: `{ "landArea": 29, "unitOfMeasurement": "hectares", "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil", "crops": [] }`,
		},
		{
			name: "Missing Land Area",
			body: `{ "name": "Farm 1", "unitOfMeasurement": "hectares", "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil", "crops": [] }`,
		},
		{
			name: "Missing Unit of Measurement",
			body: `{ "name": "Farm 1", "landArea": 29, "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil", "crops": [] }`,
		},
		{
			name: "Missing Address",
			body: `{ "name": "Farm 1", "landArea": 29, "unitOfMeasurement": "hectares", "crops": [] }`,
		},
		{
			name: "Missing Crop Type",
			body: `{ "name": "Farm 1", "landArea": 29, "unitOfMeasurement": "hectares", "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil", "crops": [ { "isIrrigated": true, "isInsured": true } ] }`,
		},
		{
			name: "Invalid Crop Type",
			body: `{ "name": "Farm 1", "landArea": 29, "unitOfMeasurement": "hectares", "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil", "crops": [ { "type": "INVALID", "isIrrigated": true, "isInsured": true } ] }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest("POST", "/farms", strings.NewReader(tt.body))
			assertStatusCode(t, w, http.StatusBadRequest)
		})
	}
}

func ListFarms(t *testing.T) {
	wipeCollections(t, "farms", "crops")
	firstFarm := map[string]interface{}{
		"name":              "Farm 1",
		"landArea":          29,
		"unitOfMeasurement": "hectares",
		"address":           "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
		"crops":             []interface{}{map[string]interface{}{"type": "CORN", "isIrrigated": true, "isInsured": true}},
	}

	secondFarm := map[string]interface{}{
		"name":              "Farm 2",
		"landArea":          39,
		"unitOfMeasurement": "hectares",
		"address":           "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
		"crops":             []interface{}{map[string]interface{}{"type": "COFFEE", "isIrrigated": true, "isInsured": true}},
	}

	farmsMap := make(map[string]interface{})
	for _, farm := range []interface{}{firstFarm, secondFarm} {
		b, err := json.Marshal(farm)
		if err != nil {
			t.Errorf("Failed to marshal farm")
		}
		var farmResponse FarmResponse

		w := performRequest("POST", "/farms", strings.NewReader(string(b)))
		err = json.Unmarshal(w.Body.Bytes(), &farmResponse)
		if err != nil {
			t.Errorf("Failed to unmarshal response body")
		}

		farmsMap[farmResponse.ID] = farm
	}

	var farmsResponse []FarmResponse

	t.Run("List farms", func(t *testing.T) {
		w := performRequest("GET", "/farms?skip=0&limit=10", nil)
		assertStatusCode(t, w, http.StatusOK)

		err := json.Unmarshal(w.Body.Bytes(), &farmsResponse)
		if err != nil {
			t.Errorf("Failed to unmarshal response body")
		}

		if len(farmsResponse) != 2 {
			t.Errorf("Expected 2 farms, but got %d", len(farmsResponse))
		}

		for _, farm := range farmsResponse {
			expectedFarm := farmsMap[farm.ID].(map[string]interface{})
			if farm.Name != expectedFarm["name"] {
				t.Errorf("Expected farm name %s, but got %s", expectedFarm["name"], farm.Name)
			}

			if farm.Address != expectedFarm["address"] {
				t.Errorf("Expected farm address %s, but got %s", expectedFarm["address"], farm.Address)
			}

			if farm.LandArea != expectedFarm["landArea"] {
				t.Errorf("Expected farm land area %d, but got %v", expectedFarm["landArea"], farm.LandArea)
			}

			if farm.UnitOfMeasurement != expectedFarm["unitOfMeasurement"] {
				t.Errorf("Expected farm unit of measurement %s, but got %s", expectedFarm["unitOfMeasurement"], farm.UnitOfMeasurement)
			}

			if len(farm.Crops) != len(expectedFarm["crops"].([]interface{})) {
				t.Errorf("Expected farm crops length %d, but got %d", len(expectedFarm["crops"].([]interface{})), len(farm.Crops))
			}

			for i, crop := range farm.Crops {
				expectedCrop := expectedFarm["crops"].([]interface{})[i].(map[string]interface{})
				if crop.Type != expectedCrop["type"] {
					t.Errorf("Expected crop type %s, but got %s", expectedCrop["type"], crop.Type)
				}

				if crop.IsIrrigated != expectedCrop["isIrrigated"] {
					t.Errorf("Expected crop isIrrigated %t, but got %t", expectedCrop["isIrrigated"], crop.IsIrrigated)
				}

				if crop.IsInsured != expectedCrop["isInsured"] {
					t.Errorf("Expected crop isInsured %t, but got %t", expectedCrop["isInsured"], crop.IsInsured)
				}
			}
		}
	})

	t.Run("Filter farms by land area", func(t *testing.T) {
		w := performRequest("GET", "/farms?skip=0&limit=1&landArea=39", nil)

		err := json.Unmarshal(w.Body.Bytes(), &farmsResponse)
		if err != nil {
			t.Errorf("Failed to unmarshal response body")
		}

		if len(farmsResponse) != 1 {
			t.Errorf("Expected 1 farm1, but got %d", len(farmsResponse))
		}

		expectedFarm := farmsMap[farmsResponse[0].ID].(map[string]interface{})

		if farmsResponse[0].LandArea != expectedFarm["landArea"] {
			t.Errorf("Expected farm land area %d, but got %v", expectedFarm["landArea"], farmsResponse[0].LandArea)
		}
	})

	t.Run("Filter farms by crop type", func(t *testing.T) {
		w := performRequest("GET", "/farms?skip=0&limit=1&cropType=CORN", nil)

		err := json.Unmarshal(w.Body.Bytes(), &farmsResponse)
		if err != nil {
			t.Errorf("Failed to unmarshal response body")
		}

		if len(farmsResponse) != 1 {
			t.Errorf("Expected 1 farm, but got %d", len(farmsResponse))
		}

		if len(farmsResponse[0].Crops) != 1 {
			t.Errorf("Expected farm crops length 1, but got %d", len(farmsResponse[0].Crops))
		}

		if farmsResponse[0].Crops[0].Type != "CORN" {
			t.Errorf("Expected farm crop type CORN, but got %s", farmsResponse[0].Crops[0].Type)
		}
	})

	t.Run("Returns paginated results", func(t *testing.T) {
		w := performRequest("GET", "/farms?skip=0&limit=1", nil)

		err := json.Unmarshal(w.Body.Bytes(), &farmsResponse)
		if err != nil {
			t.Errorf("Failed to unmarshal response body")
		}

		if len(farmsResponse) != 1 {
			t.Errorf("Expected 1 farm1, but got %d", len(farmsResponse))
		}

		if farmsResponse[0].Name != "Farm 1" {
			t.Errorf("Expected farm name Farm 1, but got %s", farmsResponse[0].Name)
		}

		w = performRequest("GET", "/farms?skip=1&limit=1", nil)

		err = json.Unmarshal(w.Body.Bytes(), &farmsResponse)
		if err != nil {
			t.Errorf("Failed to unmarshal response body")
		}

		if len(farmsResponse) != 1 {
			t.Errorf("Expected 1 farm1, but got %d", len(farmsResponse))
		}

		if farmsResponse[0].Name != "Farm 2" {
			t.Errorf("Expected farm name Farm 2, but got %s", farmsResponse[0].Name)
		}
	})
}

func FarmGet(t *testing.T) {
	body := strings.NewReader(`{
    "name": "Farm 1",
    "landArea": 87,
    "unitOfMeasurement": "hectares",
    "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
    "crops": []
  }`)

	w := performRequest("POST", "/farms", body)
	assertStatusCode(t, w, http.StatusCreated)

	var farmResponse FarmResponse
	err := json.Unmarshal(w.Body.Bytes(), &farmResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal response body")
	}

	w = performRequest("GET", fmt.Sprintf("/farms/%v", farmResponse.ID), nil)
	assertStatusCode(t, w, http.StatusOK)

	err = json.Unmarshal(w.Body.Bytes(), &farmResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal response body")
	}

	if farmResponse.Name != "Farm 1" {
		t.Errorf("Expected farm name Farm 1, but got %s", farmResponse.Name)
	}

	if farmResponse.Address != "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil" {
		t.Errorf("Expected farm address Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil, but got %s", farmResponse.Address)
	}

	if farmResponse.LandArea != 87 {
		t.Errorf("Expected farm land area 87, but got %v", farmResponse.LandArea)
	}

	if farmResponse.UnitOfMeasurement != "hectares" {
		t.Errorf("Expected farm unit of measurement hectares, but got %s", farmResponse.UnitOfMeasurement)
	}

	if len(farmResponse.Crops) != 0 {
		t.Errorf("Expected farm crops length 0, but got %d", len(farmResponse.Crops))
	}
}

func FarmUpdate(t *testing.T) {
	body := strings.NewReader(`{
    "name": "Farm 1",
    "landArea": 87,
    "unitOfMeasurement": "hectares",
    "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
    "crops": []
  }`)

	w := performRequest("POST", "/farms", body)
	assertStatusCode(t, w, http.StatusCreated)

	var farmResponse FarmResponse
	err := json.Unmarshal(w.Body.Bytes(), &farmResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal response body")
	}

	body = strings.NewReader(`{ "name": "Farm 1 Updated" }`)

	w = performRequest("PUT", fmt.Sprintf("/farms/%v", farmResponse.ID), body)

	assertStatusCode(t, w, http.StatusOK)

	w = performRequest("GET", fmt.Sprintf("/farms/%v", farmResponse.ID), nil)
	err = json.Unmarshal(w.Body.Bytes(), &farmResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal response body")
	}

	if farmResponse.Name != "Farm 1 Updated" {
		t.Errorf("Expected farm name Farm 1 Updated, but got %s", farmResponse.Name)
	}

	// Keep untouched fields
	if farmResponse.Address != "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil" {
		t.Errorf("Expected farm address Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil, but got %s", farmResponse.Address)
	}

	if farmResponse.LandArea != 87 {
		t.Errorf("Expected farm land area 87, but got %v", farmResponse.LandArea)
	}

	if farmResponse.UnitOfMeasurement != "hectares" {
		t.Errorf("Expected farm unit of measurement hectares, but got %s", farmResponse.UnitOfMeasurement)
	}
}

func FarmDelete(t *testing.T) {
	body := strings.NewReader(`{
    "name": "Farm 1",
    "landArea": 87,
    "unitOfMeasurement": "hectares",
    "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
    "crops": []
  }`)

	w := performRequest("POST", "/farms", body)
	assertStatusCode(t, w, http.StatusCreated)

	var farmResponse FarmResponse
	err := json.Unmarshal(w.Body.Bytes(), &farmResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal response body")
	}

	w = performRequest("DELETE", fmt.Sprintf("/farms/%v", farmResponse.ID), nil)
	assertStatusCode(t, w, http.StatusNoContent)

	w = performRequest("GET", fmt.Sprintf("/farms/%v", farmResponse.ID), nil)
	assertStatusCode(t, w, http.StatusNotFound)
}

func assertStatusCode(t *testing.T, response *httptest.ResponseRecorder, expected int) {
	if response.Code != expected {
		t.Errorf("Expected status code %d, but got %d", expected, response.Code)
	}
}

func performRequest(method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	s.Router.ServeHTTP(w, req)
	return w
}

func wipeCollections(t *testing.T, collectionNames ...string) {
	for _, collectionName := range collectionNames {
		_, err := db.DB.Collection(collectionName).DeleteMany(context.Background(), bson.M{})
		if err != nil {
			t.Errorf("Failed to wipe collection %s", collectionName)
		}
	}
}

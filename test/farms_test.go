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
	var farmResponse FarmResponse

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
            "type": "BEANS",
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
	parseResponse(t, w.Body.Bytes(), &farmResponse)

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
		var farmResponse FarmResponse
		b, err := json.Marshal(farm)
		if err != nil {
			t.Errorf("Failed to marshal farm")
		}

		w := performRequest("POST", "/farms", strings.NewReader(string(b)))
		parseResponse(t, w.Body.Bytes(), &farmResponse)

		farmsMap[farmResponse.ID] = farm
	}

	var farmsResponse []FarmResponse

	t.Run("List farms", func(t *testing.T) {
		w := performRequest("GET", "/farms?skip=0&limit=10", nil)
		assertStatusCode(t, w, http.StatusOK)
		parseResponse(t, w.Body.Bytes(), &farmsResponse)
		assertEqual(t, len(farmsResponse), 2, "Number of farms")

		for _, farm := range farmsResponse {
			expectedFarm := farmsMap[farm.ID].(map[string]interface{})

			assertEqual(t, farm.Name, expectedFarm["name"], "Farm name")
			assertEqual(t, farm.Address, expectedFarm["address"], "Farm address")
			assertEqual(t, farm.LandArea, expectedFarm["landArea"], "Farm land area")
			assertEqual(t, farm.UnitOfMeasurement, expectedFarm["unitOfMeasurement"], "Farm unit of measurement")
			assertEqual(t, len(farm.Crops), len(expectedFarm["crops"].([]interface{})), "Farm crops length")

			for i, crop := range farm.Crops {
				expectedCrop := expectedFarm["crops"].([]interface{})[i].(map[string]interface{})
				assertEqual(t, crop.Type, expectedCrop["type"], "Crop type")
				assertEqual(t, crop.IsIrrigated, expectedCrop["isIrrigated"], "Crop isIrrigated")
				assertEqual(t, crop.IsInsured, expectedCrop["isInsured"], "Crop isInsured")
			}
		}
	})

	t.Run("Filter farms by land area", func(t *testing.T) {
		w := performRequest("GET", "/farms?skip=0&limit=1&landArea=39", nil)
		parseResponse(t, w.Body.Bytes(), &farmsResponse)
		expectedFarm := farmsMap[farmsResponse[0].ID].(map[string]interface{})

		assertEqual(t, len(farmsResponse), 1, "Number of farms")
		assertEqual(t, farmsResponse[0].LandArea, expectedFarm["landArea"], "Farm land area")
	})

	t.Run("Filter farms by crop type", func(t *testing.T) {
		w := performRequest("GET", "/farms?skip=0&limit=1&cropType=CORN", nil)
		parseResponse(t, w.Body.Bytes(), &farmsResponse)

		assertEqual(t, len(farmsResponse), 1, "Number of farms")
		assertEqual(t, len(farmsResponse[0].Crops), 1, "Number of crops")
		assertEqual(t, farmsResponse[0].Crops[0].Type, "CORN", "Crop type")
	})

	t.Run("Returns paginated results", func(t *testing.T) {
		w := performRequest("GET", "/farms?skip=0&limit=1", nil)
		parseResponse(t, w.Body.Bytes(), &farmsResponse)

		assertEqual(t, len(farmsResponse), 1, "Number of farms")
		assertEqual(t, farmsResponse[0].Name, "Farm 1", "Farm name")

		w = performRequest("GET", "/farms?skip=1&limit=1", nil)
		parseResponse(t, w.Body.Bytes(), &farmsResponse)

		assertEqual(t, len(farmsResponse), 1, "Number of farms")
		assertEqual(t, farmsResponse[0].Name, "Farm 2", "Farm name")
	})
}

func FarmGet(t *testing.T) {
	var farmResponse FarmResponse
	body := strings.NewReader(`{
    "name": "Farm 1",
    "landArea": 87,
    "unitOfMeasurement": "hectares",
    "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
    "crops": []
  }`)

	w := performRequest("POST", "/farms", body)
	assertStatusCode(t, w, http.StatusCreated)
	parseResponse(t, w.Body.Bytes(), &farmResponse)

	w = performRequest("GET", fmt.Sprintf("/farms/%v", farmResponse.ID), nil)
	assertStatusCode(t, w, http.StatusOK)
	parseResponse(t, w.Body.Bytes(), &farmResponse)

	assertEqual(t, farmResponse.Name, "Farm 1", "Farm name")
	assertEqual(t, farmResponse.Address, "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil", "Farm address")
	assertEqual(t, farmResponse.LandArea, 87, "Farm land area")
	assertEqual(t, farmResponse.UnitOfMeasurement, "hectares", "Farm unit of measurement")
	assertEqual(t, len(farmResponse.Crops), 0, "Farm crops length")
}

func FarmUpdate(t *testing.T) {
	var farmResponse FarmResponse
	body := strings.NewReader(`{
    "name": "Farm 1",
    "landArea": 87,
    "unitOfMeasurement": "hectares",
    "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
    "crops": []
  }`)

	w := performRequest("POST", "/farms", body)
	assertStatusCode(t, w, http.StatusCreated)
	parseResponse(t, w.Body.Bytes(), &farmResponse)

	body = strings.NewReader(`{ "name": "Farm 1 Updated" }`)

	w = performRequest("PUT", fmt.Sprintf("/farms/%v", farmResponse.ID), body)

	assertStatusCode(t, w, http.StatusOK)

	w = performRequest("GET", fmt.Sprintf("/farms/%v", farmResponse.ID), nil)
	parseResponse(t, w.Body.Bytes(), &farmResponse)
	assertEqual(t, farmResponse.Name, "Farm 1 Updated", "Farm name")

	// Keep untouched fields
	assertEqual(t, farmResponse.Address, "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil", "Farm address")
	assertEqual(t, farmResponse.LandArea, 87, "Farm land area")
	assertEqual(t, farmResponse.UnitOfMeasurement, "hectares", "Farm unit of measurement")
}

func FarmDelete(t *testing.T) {
	var farmResponse FarmResponse
	body := strings.NewReader(`{
    "name": "Farm 1",
    "landArea": 87,
    "unitOfMeasurement": "hectares",
    "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
    "crops": []
  }`)

	w := performRequest("POST", "/farms", body)
	assertStatusCode(t, w, http.StatusCreated)
	parseResponse(t, w.Body.Bytes(), &farmResponse)

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

func assertEqual(t *testing.T, got, want interface{}, message string) {
	if got != want {
		t.Errorf("%s: expected %v, but got %v", message, want, got)
	}
}

func parseResponse(t *testing.T, data []byte, v interface{}) {
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}
}

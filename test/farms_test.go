package test

import (
	"context"
	"encoding/json"
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

	cropsModule := crops.New(db.DB, l)
	farmsModule := farms.New(l, &cropsModule.Repository, s, db.DB)

	mongo.HookOnStart(ctx, db, l)

	http_adapter.RegisterRoutes(farmsModule.Controller)

	defer func() {
		mongo.GracefulShutdown(ctx, db, l)
		s.GracefulShutdown(ctx)
	}()

	go s.Listen()

	t.Run("Create Farm Without Crops", CreateFarmWithoutCrops)
	t.Run("Create Farm With Crops", CreateFarmWithCrops)
	t.Run("Create Farm Compliance", CreateFarmComplianceFields)
	t.Run("List Farms", ListFarms)
}

func CreateFarmWithoutCrops(t *testing.T) {
	body := strings.NewReader(`{
    "name": "Farm 1",
    "landArea": 87,
    "unitOfMeasurement": "hectares",
    "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
    "crops": []
  }`)

	req := httptest.NewRequest("POST", "/farms", body)
	w := httptest.NewRecorder()
	s.Router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code 200, but got %d", w.Code)
	}
	var farmResponse struct {
		ID string `json:"id"`
	}
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
	req := httptest.NewRequest("POST", "/farms", b)

	w := httptest.NewRecorder()

	s.Router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code 200, but got %d", w.Code)
	}

	var farmResponse struct {
		ID string `json:"id"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &farmResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal response body")
	}
	if farmResponse.ID == "" {
		t.Errorf("Expected id, but got empty")
	}
}

func CreateFarmComplianceFields(t *testing.T) {
	t.Run("Does not create farm without name", func(t *testing.T) {
		b := strings.NewReader(`{
      "landArea": 29,
      "unitOfMeasurement": "hectares",
      "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
      "crops": []
    }`)
		req := httptest.NewRequest("POST", "/farms", b)
		w := httptest.NewRecorder()
		s.Router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code 400, but got %d", w.Code)
		}
	})

	t.Run("Does not create farm without land area", func(t *testing.T) {
		b := strings.NewReader(`{
      "name": "Farm 1",
      "unitOfMeasurement": "hectares",
      "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
      "crops": []
    }`)
		req := httptest.NewRequest("POST", "/farms", b)
		w := httptest.NewRecorder()
		s.Router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code 400, but got %d", w.Code)
		}
	})

	t.Run("Does not create farm without unit of measurement", func(t *testing.T) {
		b := strings.NewReader(`{
      "name": "Farm 1",
      "landArea": 29,
      "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
      "crops": []
    }`)
		req := httptest.NewRequest("POST", "/farms", b)
		w := httptest.NewRecorder()
		s.Router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code 400, but got %d", w.Code)
		}
	})

	t.Run("Does not create farm without address", func(t *testing.T) {
		b := strings.NewReader(`{
      "name": "Farm 1",
      "landArea": 29,
      "unitOfMeasurement": "hectares",
      "crops": []
    }`)
		req := httptest.NewRequest("POST", "/farms", b)
		w := httptest.NewRecorder()
		s.Router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code 400, but got %d", w.Code)
		}
	})

	t.Run("Does not create farm without crop type", func(t *testing.T) {
		b := strings.NewReader(`{
      "name": "Farm 1",
      "landArea": 29,
      "unitOfMeasurement": "hectares",
      "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
      "crops": [
        {
            "isIrrigated": true,
            "isInsured": true
        }
      ]
    }`)
		req := httptest.NewRequest("POST", "/farms", b)

		w := httptest.NewRecorder()

		s.Router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code 400, but got %d", w.Code)
		}
	})

	t.Run("Does not accept invalid crop type", func(t *testing.T) {
		b := strings.NewReader(`{
      "name": "Farm 1",
      "landArea": 29,
      "unitOfMeasurement": "hectares",
      "address": "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil",
      "crops": [
        {
            "type": "INVALID",
            "isIrrigated": true,
            "isInsured": true
        }
      ]
    }`)
		req := httptest.NewRequest("POST", "/farms", b)

		w := httptest.NewRecorder()

		s.Router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code 400, but got %d", w.Code)
		}
	})
}

func ListFarms(t *testing.T) {
	_, err := db.DB.Collection("farms").DeleteMany(context.Background(), bson.M{})
	if err != nil {
		t.Errorf("Failed to wipe farms before test")
	}
	_, err = db.DB.Collection("crops").DeleteMany(context.Background(), bson.M{})
	if err != nil {
		t.Errorf("Failed to wipe crops before test")
	}
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
		w := httptest.NewRecorder()
		b, err := json.Marshal(farm)
		if err != nil {
			t.Errorf("Failed to marshal farm")
		}

		var farmResponse struct {
			ID string `json:"id"`
		}

		req := httptest.NewRequest("POST", "/farms", strings.NewReader(string(b)))

		s.Router.ServeHTTP(w, req)
		err = json.Unmarshal(w.Body.Bytes(), &farmResponse)
		if err != nil {
			t.Errorf("Failed to unmarshal response body")
		}

		if w.Code != http.StatusCreated {
			t.Errorf("Failed to create farm")
		}

		farmsMap[farmResponse.ID] = farm
	}

	var farmsResponse []struct {
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

	t.Run("List farms", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/farms?skip=0&limit=10", nil)
		w := httptest.NewRecorder()
		s.Router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, but got %d", w.Code)
		}

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
		req := httptest.NewRequest("GET", "/farms?skip=0&limit=1&$landArea=39", nil)
		w := httptest.NewRecorder()
		s.Router.ServeHTTP(w, req)

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
		req := httptest.NewRequest("GET", "/farms?skip=0&limit=1&cropType=CORN", nil)
		w := httptest.NewRecorder()
		s.Router.ServeHTTP(w, req)

		err := json.Unmarshal(w.Body.Bytes(), &farmsResponse)
		if err != nil {
			t.Errorf("Failed to unmarshal response body")
		}

		if len(farmsResponse) != 1 {
			t.Errorf("Expected 1 farm1, but got %d", len(farmsResponse))
		}

		if len(farmsResponse[0].Crops) != 1 {
			t.Errorf("Expected farm crops length 1, but got %d", len(farmsResponse[0].Crops))
		}

		if farmsResponse[0].Crops[0].Type != "CORN" {
			t.Errorf("Expected farm crop type CORN, but got %s", farmsResponse[0].Crops[0].Type)
		}
	})

	t.Run("Returns paginated results", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/farms?skip=0&limit=1", nil)
		w := httptest.NewRecorder()
		s.Router.ServeHTTP(w, req)

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

		req = httptest.NewRequest("GET", "/farms?skip=1&limit=1", nil)
		w = httptest.NewRecorder()
		s.Router.ServeHTTP(w, req)

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

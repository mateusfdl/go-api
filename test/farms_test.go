package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

var (
	driver *Driver = NewDriver()
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
	driver.Start()
	defer driver.Close()

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

	w := driver.PerformRequest("POST", "/farms", body)
	AssertStatusCode(t, w, http.StatusCreated)

	var farmResponse FarmResponse
	err := json.Unmarshal(w.Body.Bytes(), &farmResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal response body")
	}
	if farmResponse.ID == "" {
		t.Errorf("Expect id, but got empty")
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
	w := driver.PerformRequest("POST", "/farms", b)
	AssertStatusCode(t, w, http.StatusCreated)
	ParseResponse(t, w.Body.Bytes(), &farmResponse)

	if farmResponse.ID == "" {
		t.Errorf("Expect id, but got empty")
	}
}

func CreateFarmComplianceFields(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		expect int
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
			w := driver.PerformRequest("POST", "/farms", strings.NewReader(tt.body))
			AssertStatusCode(t, w, http.StatusBadRequest)
		})
	}
}

func ListFarms(t *testing.T) {
	driver.WipeCollections(t, "farms", "crops")
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

		w := driver.PerformRequest("POST", "/farms", strings.NewReader(string(b)))
		ParseResponse(t, w.Body.Bytes(), &farmResponse)

		farmsMap[farmResponse.ID] = farm
	}

	var farmsResponse []FarmResponse

	t.Run("List farms", func(t *testing.T) {
		w := driver.PerformRequest("GET", "/farms?skip=0&limit=10", nil)
		AssertStatusCode(t, w, http.StatusOK)
		ParseResponse(t, w.Body.Bytes(), &farmsResponse)
		AssertEqual(t, len(farmsResponse), 2, "Number of farms")

		for _, farm := range farmsResponse {
			expectFarm := farmsMap[farm.ID].(map[string]interface{})

			AssertEqual(t, farm.Name, expectFarm["name"], "Farm name")
			AssertEqual(t, farm.Address, expectFarm["address"], "Farm address")
			AssertEqual(t, farm.LandArea, expectFarm["landArea"], "Farm land area")
			AssertEqual(t, farm.UnitOfMeasurement, expectFarm["unitOfMeasurement"], "Farm unit of measurement")
			AssertEqual(t, len(farm.Crops), len(expectFarm["crops"].([]interface{})), "Farm crops length")

			for i, crop := range farm.Crops {
				expectCrop := expectFarm["crops"].([]interface{})[i].(map[string]interface{})
				AssertEqual(t, crop.Type, expectCrop["type"], "Crop type")
				AssertEqual(t, crop.IsIrrigated, expectCrop["isIrrigated"], "Crop isIrrigated")
				AssertEqual(t, crop.IsInsured, expectCrop["isInsured"], "Crop isInsured")
			}
		}
	})

	t.Run("Filter farms by land area", func(t *testing.T) {
		w := driver.PerformRequest("GET", "/farms?skip=0&limit=1&landArea=39", nil)
		ParseResponse(t, w.Body.Bytes(), &farmsResponse)
		expectFarm := farmsMap[farmsResponse[0].ID].(map[string]interface{})

		AssertEqual(t, len(farmsResponse), 1, "Number of farms")
		AssertEqual(t, farmsResponse[0].LandArea, expectFarm["landArea"], "Farm land area")
	})

	t.Run("Filter farms by crop type", func(t *testing.T) {
		w := driver.PerformRequest("GET", "/farms?skip=0&limit=1&cropType=CORN", nil)
		ParseResponse(t, w.Body.Bytes(), &farmsResponse)

		AssertEqual(t, len(farmsResponse), 1, "Number of farms")
		AssertEqual(t, len(farmsResponse[0].Crops), 1, "Number of crops")
		AssertEqual(t, farmsResponse[0].Crops[0].Type, "CORN", "Crop type")
	})

	t.Run("Returns paginated results", func(t *testing.T) {
		w := driver.PerformRequest("GET", "/farms?skip=0&limit=1", nil)
		ParseResponse(t, w.Body.Bytes(), &farmsResponse)

		AssertEqual(t, len(farmsResponse), 1, "Number of farms")
		AssertEqual(t, farmsResponse[0].Name, "Farm 1", "Farm name")

		w = driver.PerformRequest("GET", "/farms?skip=1&limit=1", nil)
		ParseResponse(t, w.Body.Bytes(), &farmsResponse)

		AssertEqual(t, len(farmsResponse), 1, "Number of farms")
		AssertEqual(t, farmsResponse[0].Name, "Farm 2", "Farm name")
	})

	driver.WipeCollections(t, "farms", "crops")
	t.Run("Returns empty results", func(t *testing.T) {
		w := driver.PerformRequest("GET", "/farms?skip=0&limit=25", nil)
		ParseResponse(t, w.Body.Bytes(), &farmsResponse)

		AssertEqual(t, len(farmsResponse), 0, "Number of farms")
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

	w := driver.PerformRequest("POST", "/farms", body)
	AssertStatusCode(t, w, http.StatusCreated)
	ParseResponse(t, w.Body.Bytes(), &farmResponse)

	w = driver.PerformRequest("GET", fmt.Sprintf("/farms/%v", farmResponse.ID), nil)
	AssertStatusCode(t, w, http.StatusOK)
	ParseResponse(t, w.Body.Bytes(), &farmResponse)

	AssertEqual(t, farmResponse.Name, "Farm 1", "Farm name")
	AssertEqual(t, farmResponse.Address, "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil", "Farm address")
	AssertEqual(t, farmResponse.LandArea, 87, "Farm land area")
	AssertEqual(t, farmResponse.UnitOfMeasurement, "hectares", "Farm unit of measurement")
	AssertEqual(t, len(farmResponse.Crops), 0, "Farm crops length")
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

	w := driver.PerformRequest("POST", "/farms", body)
	AssertStatusCode(t, w, http.StatusCreated)
	ParseResponse(t, w.Body.Bytes(), &farmResponse)

	body = strings.NewReader(`{ "name": "Farm 1 Updated" }`)

	w = driver.PerformRequest("PUT", fmt.Sprintf("/farms/%v", farmResponse.ID), body)

	AssertStatusCode(t, w, http.StatusOK)

	w = driver.PerformRequest("GET", fmt.Sprintf("/farms/%v", farmResponse.ID), nil)
	ParseResponse(t, w.Body.Bytes(), &farmResponse)
	AssertEqual(t, farmResponse.Name, "Farm 1 Updated", "Farm name")

	// Keep untouched fields
	AssertEqual(t, farmResponse.Address, "Rua 1, 123, Bairro 2, Porto Alegre - RS, Brasil", "Farm address")
	AssertEqual(t, farmResponse.LandArea, 87, "Farm land area")
	AssertEqual(t, farmResponse.UnitOfMeasurement, "hectares", "Farm unit of measurement")
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

	w := driver.PerformRequest("POST", "/farms", body)
	AssertStatusCode(t, w, http.StatusCreated)
	ParseResponse(t, w.Body.Bytes(), &farmResponse)

	w = driver.PerformRequest("DELETE", fmt.Sprintf("/farms/%v", farmResponse.ID), nil)
	AssertStatusCode(t, w, http.StatusNoContent)

	w = driver.PerformRequest("GET", fmt.Sprintf("/farms/%v", farmResponse.ID), nil)
	AssertStatusCode(t, w, http.StatusNotFound)
}

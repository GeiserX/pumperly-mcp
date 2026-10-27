package tools

import (
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// FuelType is one fuel code the Pumperly API accepts.
type FuelType struct {
	Code     string `json:"code"`
	Label    string `json:"label"`
	Category string `json:"category"`
	// Unit is what the price is per: L, kg or kWh.
	Unit string `json:"unit"`
}

// FuelTypes mirrors FUEL_TYPE_CODES in Pumperly's src/types/fuel.ts. The API
// answers 400 to any other code.
var FuelTypes = []FuelType{
	{"B7", "Diesel (B7)", "diesel", "L"},
	{"B7_PREMIUM", "Premium diesel", "diesel", "L"},
	{"B10", "Diesel B10", "diesel", "L"},
	{"B_AGRICULTURAL", "Agricultural diesel", "diesel", "L"},
	{"HVO", "Renewable diesel (HVO)", "diesel", "L"},
	{"E5", "Petrol 95 (E5)", "gasoline", "L"},
	{"E5_PREMIUM", "Premium petrol 95", "gasoline", "L"},
	{"E10", "Petrol 95 E10", "gasoline", "L"},
	{"E5_98", "Petrol 98 (super plus)", "gasoline", "L"},
	{"E98_E10", "Petrol 98 E10", "gasoline", "L"},
	{"LPG", "LPG / autogas", "gas", "L"},
	{"CNG", "Compressed natural gas (CNG)", "gas", "kg"},
	{"LNG", "Liquefied natural gas (LNG)", "gas", "kg"},
	{"H2", "Hydrogen", "hydrogen", "kg"},
	{"EV", "EV charging", "electric", "kWh"},
	{"ADBLUE", "AdBlue", "other", "L"},
}

func fuelCodes() []string {
	codes := make([]string, len(FuelTypes))
	for i, f := range FuelTypes {
		codes[i] = f.Code
	}
	return codes
}

func fuelUnit(code string) string {
	for _, f := range FuelTypes {
		if f.Code == code {
			return f.Unit
		}
	}
	return ""
}

// fuelParam is the shared, required fuel argument.
func fuelParam() mcp.ToolOption {
	return mcp.WithString("fuel",
		mcp.Required(),
		mcp.Enum(fuelCodes()...),
		mcp.Description("Fuel code. Common ones: B7 = diesel, E5 = petrol 95, E10 = petrol 95 E10, E5_98 = petrol 98, LPG = autogas, EV = charging. The pumperly://fuel-types resource lists all 16 with labels and price units."),
	)
}

func requireFuel(req mcp.CallToolRequest) (string, error) {
	fuel, err := req.RequireString("fuel")
	if err != nil {
		return "", err
	}
	fuel = strings.ToUpper(strings.TrimSpace(fuel))
	if fuelUnit(fuel) == "" {
		return "", fmt.Errorf("unknown fuel %q; use one of %s", fuel, strings.Join(fuelCodes(), ", "))
	}
	return fuel, nil
}

package main

// Import necessary packages including the gin web framework and uuid for generating random ids
import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Defines a struct to represent the receipt JSON
type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

// Defines a struct to represent an item on the receipt
type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

// Defines a map to store the receipts in memory
var receipts = make(map[string]Receipt)

// Defines a function to generate a new ID for the receipt
func generateID() string {
	// Generate a new UUID
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err) //something is wrong with uuids so print an error
	}

	// Return the UUID as a string
	return id.String()
}

// ~Note about (c *gin.Context) input~
// gin Context is a structure that contains both the http.Request and the http.Response
// that a normal http.Handler would use, plus some useful methods and shortcuts to manipulate those

// Takes in a JSON receipt, generates id and stores it in map, and returns a JSON object with the ID
func processReceipts(c *gin.Context) {
	// Parses the JSON payload from the request, stores it in the var receipt
	var receipt Receipt
	if err := c.BindJSON(&receipt); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()}) //Bad Request error
		return
	}

	// Verify that retailer is not empty and contains only alphabets, numbers, or spaces.
	if !regexp.MustCompile(`^[A-Za-z0-9\s]+$`).MatchString(receipt.Retailer) {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid retailer name"})
		return
	}

	// Ensure that the purchaseDate is a valid date string in the format "yyyy-mm-dd"
	if _, err := time.Parse("2006-01-02", receipt.PurchaseDate); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid purchase date"})
		return
	}

	// Ensure that the purchaseTime is a valid time string in the format "hh:mm"
	if _, err := time.Parse("15:04", receipt.PurchaseTime); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid purchase time"})
		return
	}

	// Verify that each item in the "items" array has a non-empty "shortDescription" field and a valid "price" field
	for i, item := range receipt.Items {
		if item.ShortDescription == "" {
			c.AbortWithStatusJSON(400, gin.H{"error": "Item description cannot be empty"})
			return
		}
		if _, err := strconv.ParseFloat(item.Price, 64); err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": "Invalid item price"})
			return
		}
		receipt.Items[i] = item
	}

	// Generates a new ID for the receipt
	id := generateID()

	// Store the receipt in memory
	receipts[id] = receipt

	// Returns the ID in the response
	// Serializes the data as JSON and sends it to the client with an indentation of 4 spaces
	// Returns a status code of 200 (OK)
	// "id" is the string literal representing the key to the JSON object while id is the variable which contains the value for that key
	c.IndentedJSON(200, gin.H{"id": id})
}

// Helper function converting string to float 64
func stringToFloat64(total string) float64 {
	f, _ := strconv.ParseFloat(total, 64)
	return f
}

// Function to calculate the points for a receipt based on the 10 rules
func calculatePoints(receipt Receipt) int {
	points := 0

	// Rule 1: One point for every alphanumeric character in the retailer name.
	for _, c := range receipt.Retailer {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			points++
		}
	}

	// Rule 2: 50 points if the total is a round dollar amount with no cents.
	if receipt.Total == "0" || receipt.Total[len(receipt.Total)-3:] == ".00" {
		points += 50
	}

	// Rule 3: 25 points if the total is a multiple of 0.25.
	if receipt.Total == "0" || math.Mod(stringToFloat64(receipt.Total), 0.25) == 0 {
		points += 25
	}

	// Rule 4: 5 points for every two items on the receipt.
	points += len(receipt.Items) / 2 * 5

	// Rule 5: If the trimmed length of the item description is a multiple of 3, multiply the price by 0.2 and round up to the nearest integer.
	for _, item := range receipt.Items {
		if len(strings.TrimSpace(item.ShortDescription))%3 == 0 {
			points += int(math.Ceil(stringToFloat64(item.Price) * 0.2))
		}
	}

	//Rule 6: 6 points if the day in the purchase date is odd
	purchaseDate, err := time.Parse("2006-01-02", receipt.PurchaseDate)
	if err == nil && purchaseDate.Day()%2 == 1 {
		points += 6
	}

	//Rule 7: 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	purchaseTime, err := time.Parse("15:04", receipt.PurchaseTime)
	if err == nil {
		hour := purchaseTime.Hour()
		if hour >= 14 && hour < 16 {
			points += 10
		}
	}

	return points
}

// Getter endpoint that looks up the receipt by the ID and returns an object specifying the points awarded
func getPoints(c *gin.Context) {
	// Gets the ID from the URL parameter
	id := c.Param("id")

	// Looks up the receipt by ID
	receipt, ok := receipts[id]
	if !ok {
		c.AbortWithStatusJSON(404, gin.H{"error": "Receipt not found"})
		return
	}

	// Calculates the points for the receipt
	points := calculatePoints(receipt)

	// Returns the points in the response
	c.IndentedJSON(200, gin.H{"points": points})
}

func main() {

	//creates a new Gin router, Default() returns a new instance of the gin.Engine struct, which represents the main router of the Gin web framework
	router := gin.Default()

	// Defines the Process Receipts endpoint
	router.POST("/receipts/process", processReceipts)

	// Defines the Get Points endpoint
	router.GET("/receipts/:id/points", getPoints)

	// Starts the router and listens for HTPP requests on port 8080
	router.Run("localhost:8080")
}

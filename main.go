package main

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Definea a struct to represent the receipt JSON
type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

// Definea a struct to represent an item on the receipt
type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

// Definea a map to store the receipts in memory
var receipts = make(map[string]Receipt)

// Definea a function to generate a new ID for the receipt
func generateID() string {
	// Generate a new UUID
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	// Return the UUID as a string
	return id.String()
}

func processReceipts(c *gin.Context) {
	// Parse the JSON payload from the request
	var receipt Receipt
	if err := c.BindJSON(&receipt); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	// Generate a new ID for the receipt
	id := generateID()

	// Store the receipt in memory
	receipts[id] = receipt

	// Return the ID in the response
	c.IndentedJSON(200, gin.H{"id": id})
}

// Helper function converting string to float 64
func stringToFloat64(total string) float64 {
	f, _ := strconv.ParseFloat(total, 64)
	return f
}

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
	if receipt.Total == "0" || (float64(int(100*stringToFloat64(receipt.Total))/25)*0.25) == stringToFloat64(receipt.Total) {
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

func getPoints(c *gin.Context) {
	// Get the ID from the URL parameter
	id := c.Param("id")

	// Look up the receipt by ID
	receipt, ok := receipts[id]
	if !ok {
		c.AbortWithStatusJSON(404, gin.H{"error": "Receipt not found"})
		return
	}

	// Calculate the points for the receipt
	points := calculatePoints(receipt)

	// Return the points in the response
	c.IndentedJSON(200, gin.H{"points": points})
}

func main() {
	router := gin.Default()

	// Define the Process Receipts endpoint
	router.POST("/receipts/process", processReceipts)

	// Define the Get Points endpoint
	router.GET("/receipts/:id/points", getPoints)

	router.Run("localhost:8080")
}

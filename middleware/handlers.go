package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	_ "github.com/lib/pq"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nthskyradiated/stocks-api-go-postgres/models"
)

type response struct{
	ID int64 `json:"id"`
	Message string `json:"message,omitempty"`
}

func createConnection() *sql.DB{
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading env file")
	}

	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err !=nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("connected to db")
	return db
}

func CreateStock(w http.ResponseWriter, r *http.Request){
	var stock models.Stock

	err := json.NewDecoder(r.Body).Decode(&stock)
	if err != nil {
		log.Fatalf("unable to decode the request body. %v", err)
	}
	insertID := insertStock(stock)
	res := response{
		ID: insertID,
		Message: "stock created successfully",
	}
	json.NewEncoder(w).Encode(res)
}

func GetStock(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("unable to convert string %v", err)
	}
	stock, err := getStock(int64(id))

	if err != nil {
		log.Fatalf("unable to get stock %v", err)
	}
	json.NewEncoder(w).Encode(stock)
}

func GetAllStock(w http.ResponseWriter, r *http.Request){
	stocks, err := getAllStocks()
	if err != nil {
		log.Fatalf("unable to get all stocks. %v", err)
	}
	json.NewEncoder(w).Encode(stocks)
}

func UpdateStock(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert string to int. %v", err)
	}
	var stock models.Stock

	err = json.NewDecoder(r.Body).Decode(&stock)
	if err != nil {
		log.Fatalf("Unable to decode the request body. %v", err)
	}
	updatedRows := updateStock(int64(id), stock)
	msg := fmt.Sprintf("Stock updated successfully. Total rows/records affected %v", updatedRows)
	res := response {
		ID: int64(id),
		Message: msg,
	}
	json.NewEncoder(w).Encode(res)
}

func DeleteStock(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	// convert the id in string to int
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}

	// call the deleteStock, convert the int to int64
	deletedRows := deleteStock(int64(id))

	// format the message string
	msg := fmt.Sprintf("Stock updated successfully. Total rows/record affected %v", deletedRows)

	// format the reponse message
	res := response{
		ID:      int64(id),
		Message: msg,
	}

	// send the response
	json.NewEncoder(w).Encode(res)
}

func insertStock(stock models.Stock) int64{
db := createConnection()
defer db.Close()
sqlStatement := `INSERT INTO stocks(name, price, company) VALUES ($1, $2, $3) RETURNING stockid`
var id int64
err := db.QueryRow(sqlStatement, stock.Name, stock.Price, stock.Company).Scan(&id)
if err != nil {
	log.Fatalf("unable to execute query. %v", err)
}
fmt.Printf("inserted a single record. %v", id)
return id
}

func getStock(id int64)(models.Stock, error){
	db := createConnection()
	defer db.Close()

	var stock models.Stock

	sqlStatement := `SELECT * FROM stocks WHERE stockid=$1`
	row := db.QueryRow(sqlStatement, id)

	err := row.Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("no rows returned")
		return stock, nil
	case nil:
		return stock, nil
	default:
		log.Fatalf("unable to scan row. %v", err)
	}
	return stock, err
}

func getAllStocks()([]models.Stock, error){
	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	var stocks []models.Stock

	// create the select sql query
	sqlStatement := `SELECT * FROM stocks`

	// execute the sql statement
	rows, err := db.Query(sqlStatement)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	// close the statement
	defer rows.Close()

	// iterate over the rows
	for rows.Next() {
		var stock models.Stock

		// unmarshal the row object to stock
		err = rows.Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)

		if err != nil {
			log.Fatalf("Unable to scan the row. %v", err)
		}

		// append the stock in the stocks slice
		stocks = append(stocks, stock)

	}

	// return empty stock on error
	return stocks, err
}

func updateStock(id int64, stock models.Stock) int64{
		db := createConnection()

		defer db.Close()
		sqlStatement := `UPDATE stocks SET name=$2, price=$3, company=$4 WHERE stockid=$1`
		res, err := db.Exec(sqlStatement, id, stock.Name, stock.Price, stock.Company)

		if err != nil {
			log.Fatalf("unable to execute the query. %v", err)
		}
		rowsAffected, err := res.RowsAffected()

		if err != nil {
			log.Fatalf("error while checking affected rows. %v", err)
		}
		fmt.Printf("total rows affected %v", rowsAffected)
		return rowsAffected

}
func deleteStock(id int64) int64 {

	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	// create the delete sql query
	sqlStatement := `DELETE FROM stocks WHERE stockid=$1`

	// execute the sql statement
	res, err := db.Exec(sqlStatement, id)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	// check how many rows affected
	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}
package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/zamibaru89/gophermart/internal/config"
	"github.com/zamibaru89/gophermart/internal/functions"
	"github.com/zamibaru89/gophermart/internal/storage"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

var SECRET_KEY = []byte("gosecretkey")
var tokenAuth *jwtauth.JWTAuth

type Claims struct {
	Username string `json:"username"`
	ID       int    `json:"ID"`
	jwt.StandardClaims
}

func SignUp(config config.ServerConfig, st storage.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var user storage.Credentials
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			//render.JSON(w, r, user)

			return

		}
		err = st.RegisterUser(user)
		if err != nil {
			log.Println(err)
		}
		expirationTime := time.Now().Add(5 * time.Minute)

		claims := &Claims{
			Username: user.Username,
			ID:       user.ID,
			StandardClaims: jwt.StandardClaims{
				// In JWT, the expiry time is expressed as unix milliseconds
				ExpiresAt: expirationTime.Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenString, err := token.SignedString(SECRET_KEY)
		if err != nil {
			// If there is an error in creating the JWT return an internal server error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:    "jwt",
			Value:   tokenString,
			Expires: expirationTime,
		})
	}
}
func Welcome(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	w.Write([]byte(fmt.Sprintf("Hello %v ID %v", claims["username"], claims["ID"])))
}

func PostOrder(config config.ServerConfig, st storage.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, _ := jwtauth.FromContext(r.Context())
		var order storage.Order
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {

			w.WriteHeader(http.StatusBadRequest)
		}
		orderId, err := strconv.Atoi(string(body))
		order.OrderID = int64(orderId)
		ID := claims["ID"]

		order.UserID = int(ID.(float64))

		checkOrder, err := st.GetOrderByOrderID(order.OrderID)

		if checkOrder.UserID != 0 {
			if checkOrder.UserID == order.UserID {
				w.WriteHeader(http.StatusOK)
				return
			} else {
				w.WriteHeader(http.StatusConflict)
				return
			}

		}
		if !functions.CheckOrderId(order.OrderID) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		order.State = "New"
		order.UploadedAt = time.Now()
		err = st.PostOrder(order)
		if err != nil {
			return
		}
	}
}

func GetOrders(config config.ServerConfig, st storage.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var order storage.Order
		_, claims, _ := jwtauth.FromContext(r.Context())
		ID := claims["ID"]

		order.UserID = int(ID.(float64))

		listOrders, err := st.GetOrdersByUserID(int64(order.UserID))

		if err != nil {

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(listOrders) == 0 {
			w.WriteHeader(http.StatusNoContent)
		}
		render.JSON(w, r, listOrders)

	}
}
func GetBalance(config config.ServerConfig, st storage.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var balance storage.Balance
		_, claims, _ := jwtauth.FromContext(r.Context())
		ID := claims["ID"]

		balance.UserID = int(ID.(float64))

		current, err := st.GetBalanceByUserID(int64(balance.UserID))

		if err != nil {

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		balance.Current = current

		render.JSON(w, r, balance)

	}
}

func PostWithdrawal(config config.ServerConfig, st storage.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, _ := jwtauth.FromContext(r.Context())
		var withdrawal storage.Withdrawal
		err := json.NewDecoder(r.Body).Decode(&withdrawal)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			//render.JSON(w, r, user)
			fmt.Println(err)
			return

		}

		ID := claims["ID"]

		withdrawal.UserID = int(ID.(float64))

		if !functions.CheckOrderId(withdrawal.OrderID) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		currentBalance, err := st.GetBalanceByUserID(int64(withdrawal.UserID))
		balanceAfter := currentBalance - withdrawal.Sum
		if balanceAfter >= 0 {
			withdrawal.ProcessedAt = time.Now()
			err = st.PostWithdrawal(withdrawal)
			if err != nil {
				return
			}
			err = st.SetBalanceByUserID(int64(withdrawal.UserID), balanceAfter)
			if err != nil {
				return
			}
		} else {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}

	}
}
func GetWithdrawals(config config.ServerConfig, st storage.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var withdrawal storage.Withdrawal
		_, claims, _ := jwtauth.FromContext(r.Context())
		ID := claims["ID"]

		withdrawal.UserID = int(ID.(float64))

		listWithdrawals, err := st.GetOrdersByUserID(int64(withdrawal.UserID))

		if err != nil {

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(listWithdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
		}
		render.JSON(w, r, listWithdrawals)

	}
}
func SignIn(config config.ServerConfig, st storage.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var user storage.Credentials
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			//render.JSON(w, r, user)

			return

		}
		user, err = st.SignIn(user)
		if err != nil {
			log.Println(err)
		}
		fmt.Println(user.ID)
		if user.ID != 0 {
			expirationTime := time.Now().Add(5 * time.Minute)
			// Create the JWT claims, which includes the username and expiry time
			claims := &Claims{
				Username: user.Username,
				ID:       user.ID,
				StandardClaims: jwt.StandardClaims{
					// In JWT, the expiry time is expressed as unix milliseconds
					ExpiresAt: expirationTime.Unix(),
				},
			}

			// Declare the token with the algorithm used for signing, and the claims
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			// Create the JWT string
			tokenString, err := token.SignedString(SECRET_KEY)
			if err != nil {
				// If there is an error in creating the JWT return an internal server error
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Add("jwt", tokenString)
			w.Header().Add("Set-Cookie", "jwt="+tokenString)

			http.SetCookie(w, &http.Cookie{
				Name:    "jwt",
				Value:   tokenString,
				Expires: expirationTime,
			})
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}

	}
}

func main() {
	tokenAuth = jwtauth.New("HS256", SECRET_KEY, nil)
	ServerConfig, err := config.LoadServerConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	Server, Conn, err := storage.NewPostgresStorage(ServerConfig)
	fmt.Println(Server, Conn)
	if err != nil {
		log.Fatal(err)
		return

	}
	r := chi.NewRouter()

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", SignUp(ServerConfig, Server))
		r.Post("/login", SignIn(ServerConfig, Server))
		r.With(jwtauth.Verifier(tokenAuth)).Post("/welcome", Welcome)
		r.With(jwtauth.Verifier(tokenAuth)).Post("/orders", PostOrder(ServerConfig, Server))
		r.With(jwtauth.Verifier(tokenAuth)).Get("/orders", GetOrders(ServerConfig, Server))
		r.With(jwtauth.Verifier(tokenAuth)).Get("/balance", GetBalance(ServerConfig, Server))
		r.With(jwtauth.Verifier(tokenAuth)).Post("/balance/withdraw", PostWithdrawal(ServerConfig, Server))
		r.With(jwtauth.Verifier(tokenAuth)).Post("/balance/withdrawals", PostWithdrawal(ServerConfig, Server))
	})
	http.ListenAndServe(ServerConfig.Address, r)
}

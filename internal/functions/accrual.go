package functions

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/zamibaru89/gophermart/internal/config"
	"github.com/zamibaru89/gophermart/internal/storage"
	"log"
	"net/http"
	"strconv"
	"time"
)

func AccrualUpdate(repo storage.Repo, conf config.ServerConfig) error {
	updateList, err := repo.GetOrdersForUpdate()
	if err != nil {
		return err
	}

	if len(updateList) != 0 {
		req := resty.New().
			SetBaseURL(conf.AccrualAddress).
			R().
			SetHeader("Content-Type", "application/json")
		for _, order := range updateList {

			orderNum := strconv.FormatInt(order.OrderID, 10)
			resp, err := req.Get("/api/orders/" + orderNum)
			if err != nil {
				return err
			}
			log.Println(resp)
			status := resp.StatusCode()
			switch status {
			case http.StatusTooManyRequests:
				time.Sleep(60 * time.Second)
				return nil

			case http.StatusOK:
				var toUpdateOrder storage.Accrual
				err = json.Unmarshal(resp.Body(), &toUpdateOrder)
				if err != nil {
					return err
				}
				order.State = toUpdateOrder.State
				order.OrderID, err = strconv.ParseInt(toUpdateOrder.OrderID, 10, 64)
				if err != nil {
					return err
				}
				order.Accrual = toUpdateOrder.Accrual

				err = repo.PostOrder(order)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

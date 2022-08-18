package functions

import (
	"github.com/go-resty/resty/v2"
	"github.com/zamibaru89/gophermart/internal/config"
	"github.com/zamibaru89/gophermart/internal/storage"
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
			status := resp.StatusCode()
			switch status {
			case http.StatusTooManyRequests:
				time.Sleep(60 * time.Second)
				return nil

			case http.StatusOK:
				var toUpdateOrder storage.Order
				toUpdateOrder.State = order.State
				toUpdateOrder.Accrual = order.Accrual
				toUpdateOrder.OrderID = order.OrderID
				err = repo.PostOrder(order)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

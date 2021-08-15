package web

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type RuleListItem struct {
	Key       int    `json:"key"`
	Disabled  bool   `json:"disabled"`
	Href      string `json:"href"`
	Avatar    string `json:"avatar"`
	Name      string `json:"name"`
	Owner     string `json:"owner"`
	Desc      string `json:"desc"`
	CallNo    int    `json:"callNo"`
	Status    int    `json:"status"`
	UpdatedAt string `json:"updatedAt"`
	CreatedAt string `json:"createdAt"`
	Progress  int    `json:"progress"`
}

type RuleList struct {
	Data    []RuleListItem `json:"data"`
	Total   int            `json:"total"`
	Success bool           `json:"success"`
}

func genList(current, size int) (ls []RuleListItem) {
	a := []string{
		"https://gw.alipayobjects.com/zos/rmsportal/eeHMaZBwmTvLdIwMfBpg.png",
		"https://gw.alipayobjects.com/zos/rmsportal/udxAbMEhpwthVVcjLXik.png",
	}
	date := time.Now().Format("2006-01-02")
	for i := 0; i < size; i++ {
		index := (current-1)*10 + i
		ls = append(ls, RuleListItem{
			Key:       index,
			Disabled:  (i % 6) == 0,
			Href:      "https://ant.design",
			Avatar:    a[i%2],
			Name:      fmt.Sprintf("TradeCode %d", index),
			Owner:     "曲丽丽",
			Desc:      "这是一段描述",
			CallNo:    rand.Int() * 1000,
			Status:    (rand.Int() * 10) % 4,
			Progress:  rand.Int() * 100,
			CreatedAt: date,
			UpdatedAt: date,
		})
	}
	return
}

var tableListDataSource = genList(1, 100)

func GetRule(w http.ResponseWriter, r *http.Request) {
	send(w, http.StatusOK, RuleList{
		Data:    tableListDataSource,
		Total:   len(tableListDataSource),
		Success: true,
	})
}

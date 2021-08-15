package web

import "net/http"

type Notice struct {
	ID          string `json:"id"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	Extra       string `json:"extra"`
	Status      string `json:"status"`
	Datetime    string `json:"datetime"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Read        bool   `json:"read"`
	ClickClose  bool   `json:"ClickClose"`
}

func GetNotice(w http.ResponseWriter, r *http.Request) {
	send(w, http.StatusOK, map[string]interface{}{
		"data": []Notice{
			{
				ID:       "000000001",
				Avatar:   "https://gw.alipayobjects.com/zos/rmsportal/ThXAXghbEsBCCSDihZxY.png",
				Title:    "你收到了 14 份新周报",
				Datetime: "2017-08-09",
				Type:     "notification",
			},
			{
				ID:       "000000002",
				Avatar:   "https://gw.alipayobjects.com/zos/rmsportal/OKJXDXrmkNshAMvwtvhu.png",
				Title:    "你推荐的 曲妮妮 已通过第三轮面试",
				Datetime: "2017-08-08",
				Type:     "notification",
			},
			{
				ID:       "000000003",
				Avatar:   "https://gw.alipayobjects.com/zos/rmsportal/kISTdvpyTAhtGxpovNWd.png",
				Title:    "这种模板可以区分多种通知类型",
				Datetime: "2017-08-07",
				Read:     true,
				Type:     "notification",
			},
			{
				ID:       "000000004",
				Avatar:   "https://gw.alipayobjects.com/zos/rmsportal/GvqBnKhFgObvnSGkDsje.png",
				Title:    "左侧图标用于区分不同的类型",
				Datetime: "2017-08-07",
				Type:     "notification",
			},
			{
				ID:       "000000005",
				Avatar:   "https://gw.alipayobjects.com/zos/rmsportal/ThXAXghbEsBCCSDihZxY.png",
				Title:    "内容不要超过两行字，超出时自动截断",
				Datetime: "2017-08-07",
				Type:     "notification",
			},
			{
				ID:          "000000006",
				Avatar:      "https://gw.alipayobjects.com/zos/rmsportal/fcHMVNCjPOsbUGdEduuv.jpeg",
				Title:       "曲丽丽 评论了你",
				Description: "描述信息描述信息描述信息",
				Datetime:    "2017-08-07",
				Type:        "message",
				ClickClose:  true,
			},
			{
				ID:          "000000007",
				Avatar:      "https://gw.alipayobjects.com/zos/rmsportal/fcHMVNCjPOsbUGdEduuv.jpeg",
				Title:       "朱偏右 回复了你",
				Description: "这种模板用于提醒谁与你发生了互动，左侧放『谁』的头像",
				Datetime:    "2017-08-07",
				Type:        "message",
				ClickClose:  true,
			},
			{
				ID:          "000000008",
				Avatar:      "https://gw.alipayobjects.com/zos/rmsportal/fcHMVNCjPOsbUGdEduuv.jpeg",
				Title:       "标题",
				Description: "这种模板用于提醒谁与你发生了互动，左侧放『谁』的头像",
				Datetime:    "2017-08-07",
				Type:        "message",
				ClickClose:  true,
			},
			{
				ID:          "000000009",
				Title:       "任务名称",
				Description: "任务需要在 2017-01-12 20:00 前启动",
				Extra:       "未开始",
				Status:      "todo",
				Type:        "event",
			},
			{
				ID:          "000000010",
				Title:       "第三方紧急代码变更",
				Description: "冠霖提交于 2017-01-06，需在 2017-01-07 前完成代码变更任务",
				Extra:       "马上到期",
				Status:      "urgent",
				Type:        "event",
			},
			{
				ID:          "000000011",
				Title:       "信息安全考试",
				Description: "指派竹尔于 2017-01-09 前完成更新并发布",
				Extra:       "已耗时 8 天",
				Status:      "doing",
				Type:        "event",
			},
			{
				ID:          "000000012",
				Title:       "ABCD 版本发布",
				Description: "冠霖提交于 2017-01-06，需在 2017-01-07 前完成代码变更任务",
				Extra:       "进行中",
				Status:      "processing",
				Type:        "event",
			},
		},
	})
}

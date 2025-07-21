package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"one-api/common"
	"one-api/model"
	"one-api/setting"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/webhook"
)

type StripeRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	TopUpCode     string `json:"top_up_code"`
}

func GetStripeClient() bool {
	if setting.StripeSecretKey == "" {
		return false
	}
	stripe.Key = setting.StripeSecretKey
	// 设置 API 版本为最新的 basil 版本
	stripe.SetAppInfo(&stripe.AppInfo{
		Name:    "aiapi-services",
		Version: "1.0.0",
		URL:     "https://aiapi.services",
	})
	return true
}

func RequestStripe(c *gin.Context) {
	var req StripeRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	if req.Amount < getMinTopup() {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getMinTopup())})
		return
	}

	if !GetStripeClient() {
		c.JSON(200, gin.H{"message": "error", "data": "管理员未配置Stripe secret key"})
		return
	}

	if setting.StripePriceID == "" {
		c.JSON(200, gin.H{"message": "error", "data": "管理员未配置Stripe价格ID"})
		return
	}

	id := c.GetInt("id")
	user, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "用户不存在"})
		return
	}

	topUp := &model.TopUp{
		UserId:  user.Id,
		Amount:  req.Amount,
		Status:  "pending",
		TradeNo: fmt.Sprintf("stripe_%s", common.GetUUID()),
	}
	err = topUp.Insert()
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",         // 银行卡支付 (信用卡/借记卡)
			"alipay",       // 支付宝
			"wechat_pay",   // 微信支付
			"crypto",       // 稳定币支付 (USDC, USDT 等)
			"paypal",       // PayPal 支付
			"link",         // Stripe Link (快速结账)
			"amazon_pay",   // Amazon Pay
			"cashapp",      // Cash App Pay
			"click_to_pay", // Click To Pay
			"google_pay",   // Google Pay
			"apple_pay",    // Apple Pay (需要验证域名)
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(setting.StripePriceID),
				Quantity: stripe.Int64(req.Amount),
			},
		},
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(fmt.Sprintf("%s/topup?trade_no=%s", setting.ServerAddress, topUp.TradeNo)),
		CancelURL:         stripe.String(fmt.Sprintf("%s/topup", setting.ServerAddress)),
		ClientReferenceID: stripe.String(topUp.TradeNo),
	}
	s, err := session.New(params)
	if err != nil {
		log.Printf("Stripe session.New failed for user %d, amount %d: %v", user.Id, req.Amount, err)

		// 根据不同的错误类型返回不同的提示
		var errorMessage string
		if stripeErr, ok := err.(*stripe.Error); ok {
			switch stripeErr.Code {
			case stripe.ErrorCodeParameterInvalidStringEmpty:
				errorMessage = "支付参数配置错误，请联系管理员"
			case stripe.ErrorCodeParameterInvalidInteger:
				errorMessage = "支付金额无效"
			default:
				errorMessage = fmt.Sprintf("支付服务暂时不可用: %s", stripeErr.Msg)
			}
		} else {
			errorMessage = "创建支付会话失败"
		}

		c.JSON(200, gin.H{"message": "error", "data": errorMessage})
		return
	}
	c.JSON(200, gin.H{"message": "success", "data": s.URL})
}

func StripeWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		c.Status(http.StatusServiceUnavailable)
		return
	}

	event, err := webhook.ConstructEvent(body, c.Request.Header.Get("Stripe-Signature"), setting.StripeWebhookSecret)
	if err != nil {
		log.Printf("webhook.ConstructEvent: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook error"})
		return
	}

	if event.Type == "checkout.session.completed" {
		var checkoutSession stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &checkoutSession)
		if err != nil {
			log.Printf("json.Unmarshal: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "json.Unmarshal error"})
			return
		}
		tradeNo := checkoutSession.ClientReferenceID
		topUp := model.GetTopUpByTradeNo(tradeNo)
		if topUp == nil {
			log.Printf("model.GetTopUpByTradeNo: tradeNo %s not found", tradeNo)
			c.JSON(http.StatusOK, gin.H{"error": "topup not found"})
			return
		}
		if topUp.Status == "paid" {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
			return
		}
		topUp.Status = "paid"
		err = topUp.Update()
		if err != nil {
			log.Printf("topUp.Update: %v", err)
			c.JSON(http.StatusOK, gin.H{"error": "update topup status failed"})
			return
		}
		err = model.IncreaseUserQuota(topUp.UserId, int(topUp.Amount), true)
		if err != nil {
			log.Printf("IncreaseUserQuota: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

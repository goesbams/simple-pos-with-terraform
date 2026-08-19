package service

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type MidtransService interface {
	CreateSnapTransaction(orderID string, amount int64, paymentType string) (string, string, error)
	VerifySignatureKey(orderID, statusCode, grossAmount, signatureKey string) bool
}

type midtransService struct {
	serverKey string
	client    snap.Client
}

func NewMidtransService(serverKey, clientKey string, isProduction bool) MidtransService {
	c := snap.Client{}
	env := midtrans.Sandbox
	if isProduction {
		env = midtrans.Production
	}
	c.New(serverKey, env)

	return &midtransService{
		serverKey: serverKey,
		client:    c,
	}
}

func (s *midtransService) CreateSnapTransaction(orderID string, amount int64, paymentType string) (string, string, error) {
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: amount,
		},
	}

	snapResp, err := s.client.CreateTransaction(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to create midtrans snap transaction: %w", err)
	}

	return snapResp.Token, snapResp.RedirectURL, nil
}

func (s *midtransService) VerifySignatureKey(orderID, statusCode, grossAmount, signatureKey string) bool {
	// Midtrans Signature Formula: SHA512(order_id + status_code + gross_amount + ServerKey)
	raw := orderID + statusCode + grossAmount + s.serverKey
	hasher := sha512.New()
	hasher.Write([]byte(raw))
	expectedSignature := hex.EncodeToString(hasher.Sum(nil))

	return expectedSignature == signatureKey
}

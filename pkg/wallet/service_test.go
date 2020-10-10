package wallet

import (
	"fmt"
	"reflect"
	"testing"
	"github.com/coursar/wallet/pkg/types"
	"github.com/google/uuid"
)

type testService struct {
	*Service
}

func newTestService() *testService {
	return &testService{Service: &Service{}}
}

func (s *testService) addAccountWithBalance(phone types.Phone, balance types.Money) (*types.Account, error) {
	// регистрация пользователя
	account, err := s.RegisterAccount(phone)
	if err != nil {
		return nil, fmt.Errorf("can't register account, error = %v", err)
	}

	// пополнение счёта
	err = s.Deposit(account.ID, balance)
	if err != nil {
		return nil, fmt.Errorf("can't deposit account, error = %v", err)
	}

	return account, nil
}

type testAccount struct {
	phone types.Phone
	balance types.Money
	payments []struct {
		amount types.Money
		category types.PaymentCategory
	}
}

var defaultTestAccount = testAccount{
	phone: "+992000000001",
	balance: 10_000_00,
	payments: []struct {
		amount types.Money
		category types.PaymentCategory
	}{
		{amount: 1_000_00, category: "auto"},
	},
}

func (s *testService) addAccount(data testAccount) (*types.Account, []*types.Payment, error) {
	// регистрация пользователя
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("can't register account, error = %v", err)
	}

	// пополнение счёта
	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("can't deposit account, error = %v", err)
	}
	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("can't make payment, error = %v", err)
		}
	}
	return account, payments, nil
}

func TestService_RegisterAccount_success(t *testing.T) {
	svc := &Service{}
	expected, _ := svc.RegisterAccount("+992000000001")
	result, _ := svc.FindAccountByID(1)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("invalid result, expected: %v, actual: %v", expected, result)
	}
}

func TestService_RegisterAccount_fail(t *testing.T) {
	svc := &Service{}
	svc.RegisterAccount("+992000000001")
	expected := ErrAccountNotFound
	_, result := svc.FindAccountByID(2)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("invalid result, expected: %v, actual: %v", expected, result)
	}
}

func TestService_RegisterAccount_exist(t *testing.T) {
	svc := &Service{}
	svc.RegisterAccount("+992000000001")
	expected := ErrPhoneRegistered
	_, result := svc.RegisterAccount("+992000000001")

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("invalid result, expected: %v, actual: %v", expected, result)
	}
}

func TestService_Reject_success(t *testing.T) {
	svc := &Service{}
	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("Reject(): can't register account, error = %v", err)
		return
	}
	err = svc.Deposit(account.ID, 10_000_00)
	if err != nil {
		t.Errorf("Reject(): can't deposit account, error = %v", err)
		return
	}
	payment, err := svc.Pay(account.ID, 1000_00, "auto")
	if err != nil {
		t.Errorf("Reject(): can't create payment, error = %v", err)
		return
	}
	err = svc.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): can't reject payment, error = %v", err)
		return
	}
}

func TestService_FindPaymentByID_success(t *testing.T) {
	s := newTestService()
	account, err := s.addAccountWithBalance("+992000000001", 10_000_00)
	if err != nil {
		t.Errorf("FindPaymentByID(): can't add account, error = %v", err)
		return
	}

	payment, err := s.Pay(account.ID, 1000_00, "auto")
	if err != nil {
		t.Errorf("FindPaymentByID(): can't create payment, error = %v", err)
		return
	}

	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): can't find payment, error = %v", err)
		return
	}

	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID(): wrong payment returned = %v", err)
		return	
	}
}

func TestService_FindPaymentByID_fail(t *testing.T) {
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}

	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Errorf("FindPaymentByID(): must return error, returned nil")
		return
	}

	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentByID(): must return ErrPaymentNotFound, returned = %v", err)
		return
	}
}

func TestService_Repeat_success(t *testing.T) {
	s := newTestService()
	account, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Errorf("Repeat(): error = %v", err)
		return
	}

	payment, err := s.Pay(account.ID, 1000_00, "auto")
	if err != nil {
		t.Errorf("Repeat(): can't create payment, error = %v", err)
		return
	}
	_, err = s.Repeat(payment.ID)
	if err != nil {
		t.Errorf("Reject(): can't repeat payment, error = %v", err)
		return
	}
}

func TestService_Repeat_fail(t *testing.T) {
	s := newTestService()
	
	_, err := s.Repeat(uuid.New().String())
	if err == nil {
		t.Errorf("Reject(): must return ErrPaymentNotFound, returned = %v", err)
		return
	}
}

func TestService_FavoritePayment_success(t *testing.T) {
	s := newTestService()
	account, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Errorf("FavoritePayment(): error = %v", err)
		return
	}

	payment, err := s.Pay(account.ID, 1000_00, "auto")
	if err != nil {
		t.Errorf("FavoritePayment(): can't create payment, error = %v", err)
		return
	}
	_, err = s.FavoritePayment(payment.ID, "my auto")
	if err != nil {
		t.Errorf("FavoritePayment(): can't create favorite, error = %v", err)
		return
	}
}

func TestService_FavoritePayment_fail(t *testing.T) {
	s := newTestService()
	
	_, err := s.FavoritePayment(uuid.New().String(), "my favorite")
	if err == nil {
		t.Errorf("FavoritePayment(): must return ErrPaymentNotFound, returned = %v", err)
		return
	}
}

func TestService_PayFromFavorite_success(t *testing.T) {
	s := newTestService()
	account, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Errorf("PayFromFavorite(): error = %v", err)
		return
	}

	payment, err := s.Pay(account.ID, 1000_00, "auto")
	if err != nil {
		t.Errorf("PayFromFavorite(): can't create payment, error = %v", err)
		return
	}

	favorite, err := s.FavoritePayment(payment.ID, "my auto")
	if err != nil {
		t.Errorf("PayFromFavorite(): can't create favorite, error = %v", err)
		return
	}

	_, err = s.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Errorf("PayFromFavorite(): can't pay from favorite, error = %v", err)
		return
	}
}

func TestService_PayFromFavorite_fail(t *testing.T) {
	s := newTestService()
	
	_, err := s.PayFromFavorite(uuid.New().String())
	if err == nil {
		t.Errorf("PayFromFavorite(): must return ErrFavoriteNotFound, returned = %v", err)
		return
	}
}
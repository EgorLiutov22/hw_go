package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

var ErrBudgetExceeded = errors.New("budget exceeded")

type Transaction struct {
	ID       string  `json:"id"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
}

type Budget struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period,omitempty"`
}

type Ledger struct {
	transactions []Transaction
	budgets      map[string]Budget
}

func NewLedger() *Ledger {
	return &Ledger{
		transactions: make([]Transaction, 0),
		budgets:      make(map[string]Budget),
	}
}

func (l *Ledger) SetBudget(b Budget) {
	l.budgets[b.Category] = b
}

func (l *Ledger) AddTransaction(tx Transaction) error {
	budget, exists := l.budgets[tx.Category]
	if exists {
		var currentTotal float64
		for _, t := range l.transactions {
			if t.Category == tx.Category {
				currentTotal += t.Amount
			}
		}

		if currentTotal+tx.Amount > budget.Limit {
			return ErrBudgetExceeded
		}
	}

	l.transactions = append(l.transactions, tx)
	return nil
}

func (l *Ledger) LoadBudgets(r io.Reader) error {
	var budgets []Budget
	decoder := json.NewDecoder(r)

	if err := decoder.Decode(&budgets); err != nil {
		return fmt.Errorf("failed to decode budgets JSON: %w", err)
	}

	for _, b := range budgets {
		if b.Category == "" {
			return errors.New("category cannot be empty")
		}
		if b.Limit < 0 {
			return fmt.Errorf("limit for '%s' cannot be negative", b.Category)
		}
		l.SetBudget(b)
	}

	return nil
}

func (l *Ledger) GetTransactions() []Transaction {
	return l.transactions
}

func main() {
	ledger := NewLedger()

	ledger.SetBudget(Budget{Category: "еда", Limit: 5000})

	txOK := Transaction{ID: "1", Category: "еда", Amount: 3000}
	if err := ledger.AddTransaction(txOK); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	} else {
		fmt.Println("Успешно добавлено: Еда 3000")
	}

	txFail := Transaction{ID: "2", Category: "еда", Amount: 3000}
	if err := ledger.AddTransaction(txFail); err != nil {
		fmt.Printf("Отказ (ожидаемо): %v\n", err)
	} else {
		fmt.Println("Успешно добавлено: Еда 3000")
	}

	file, err := os.Open("budgets.json")
	if err != nil {
		fmt.Printf("Ошибка открытия файла: %v\n", err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	if err := ledger.LoadBudgets(reader); err != nil {
		fmt.Printf("Ошибка парсинга JSON: %v\n", err)
		return
	}

	for _, tx := range ledger.GetTransactions() {
		fmt.Printf("- [%s] Категория: %s, Сумма: %.2f\n", tx.ID, tx.Category, tx.Amount)
	}
}

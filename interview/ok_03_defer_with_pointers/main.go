package main

import "fmt"

type Account struct {
	Balance int
}

// Что будет выведено на экран?
func main() {
	account := &Account{Balance: 1000}

	defer printBalance("Изначальный баланс", account.Balance)
	defer printBalance("Текущий баланс", account.Balance)
	defer printAccountBalance("Указатель на баланс", account)

	account.Balance += 500
	updateBalance(account, 200)
	account = &Account{Balance: 300}
}

func updateBalance(acc *Account, amount int) {
	acc.Balance -= amount
}

func printBalance(label string, balance int) {
	fmt.Printf("%s: %d\n", label, balance)
}

func printAccountBalance(label string, acc *Account) {
	fmt.Printf("%s: %d\n", label, acc.Balance)
}

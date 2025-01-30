package main

import (
	"context"
	"log"
	"os"
	"runtime/debug"

	"github.com/kirill-a-belov/solana_token_manager/cmd"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			log.Fatal("panic while command execution",
				"error", err,
			)
		}
	}()

	ctx := context.Background()

	if err := cmd.New(ctx).ExecuteContext(ctx); err != nil {
		os.Exit(-1)
	}

	/*// Создаем клиента для подключения к Solana
	c := client.NewClient("https://api.mainnet-beta.solana.com")

	// Генерация ключей для токена и владельца
	owner := types.NewAccount()
	mint := types.NewAccount()

	fmt.Printf("Адрес владельца: %s\n", owner.PublicKey.ToBase58())
	fmt.Printf("Адрес токена: %s\n", mint.PublicKey.ToBase58())

	// Создаем транзакцию для создания нового токена
	createMintTx, err := tokenprog.NewCreateMintInstruction(
		mint.PublicKey,           // токен-аккаунт
		owner.PublicKey,          // адрес владельца токена
		owner.PublicKey,          // адрес, имеющий право на дальнейшую эмиссию
		0,                        // кол-во десятичных знаков
		tokenprog.TokenProgramID, // ID программы SPL Token
	)
	if err != nil {
		log.Fatalf("Ошибка создания токена: %v", err)
	}

	// Подписываем и отправляем транзакцию
	simulateAndSendTx(c, []types.Instruction{createMintTx}, []types.Account{owner, mint})

	// Выпуск токенов
	mintAmount := uint64(1000000) // 1,000,000 токенов
	mintToTx, err := tokenprog.NewMintToInstruction(
		mint.PublicKey,                     // адрес токена
		owner.PublicKey,                    // кошелек для зачисления токенов
		mintAmount,                         // сумма эмиссии
		[]types.PublicKey{owner.PublicKey}, // подписываем транзакцию
		tokenprog.TokenProgramID,
	)
	if err != nil {
		log.Fatalf("Ошибка выпуска токенов: %v", err)
	}

	// Подписываем и отправляем транзакцию
	simulateAndSendTx(c, []types.Instruction{mintToTx}, []types.Account{owner})

	// Блокировка дальнейшей эмиссии
	setAuthorityTx, err := tokenprog.NewSetAuthorityInstruction(
		mint.PublicKey,                    // токен, для которого устанавливается новый владелец
		owner.PublicKey,                   // текущий владелец
		nil,                               // блокировка (установка пустого владельца)
		tokenprog.AuthorityTypeMintTokens, // тип операции (блокировка эмиссии)
		[]types.PublicKey{owner.PublicKey},
		tokenprog.TokenProgramID,
	)
	if err != nil {
		log.Fatalf("Ошибка блокировки эмиссии: %v", err)
	}

	// Подписываем и отправляем транзакцию
	simulateAndSendTx(c, []types.Instruction{setAuthorityTx}, []types.Account{owner})

	fmt.Println("Токен создан и эмиссия завершена.")*/
}

package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"gintegra/internal/database"
	"gintegra/internal/models"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	// Load .env
	godotenv.Load()

	// Define commands
	resetPassword := flag.Bool("reset-password", false, "Сбросить пароль администратора")
	generateSecret := flag.Bool("generate-secret", false, "Сгенерировать случайный URL для приложения")
	flag.Parse()

	database.Connect()

	if *resetPassword {
		resetAdminPassword()
	} else if *generateSecret {
		generateRandomSecret()
	} else {
		showMenu()
	}
}

func showMenu() {
	fmt.Println("=== GIntegra Admin CLI ===")
	fmt.Println("1. Сбросить пароль администратора")
	fmt.Println("2. Сгенерировать случайный URL")
	fmt.Println("0. Выход")
	fmt.Print("\nВыберите действие: ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		resetAdminPassword()
	case "2":
		generateRandomSecret()
	case "0":
		fmt.Println("Выход...")
	default:
		fmt.Println("Неверный выбор")
	}
}

func resetAdminPassword() {
	fmt.Println("\n=== Сброс пароля администратора ===")
	
	var admin models.User
	if err := database.DB.Where("role = ?", "admin").First(&admin).Error; err != nil {
		log.Fatal("Администратор не найден:", err)
	}

	fmt.Printf("Найден администратор: %s\n", admin.Username)
	fmt.Print("Введите новый пароль: ")
	
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("Ошибка чтения пароля:", err)
	}
	fmt.Println()

	if len(password) < 4 {
		log.Fatal("Пароль должен быть не менее 4 символов")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Ошибка хеширования пароля:", err)
	}

	admin.Password = string(hashedPassword)
	if err := database.DB.Save(&admin).Error; err != nil {
		log.Fatal("Ошибка сохранения пароля:", err)
	}

	fmt.Println("✅ Пароль успешно изменен!")
}

func generateRandomSecret() {
	fmt.Println("\n=== Генерация случайного URL ===")
	
	// Generate random path
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal("Ошибка генерации:", err)
	}
	randomPath := hex.EncodeToString(bytes)

	fmt.Printf("\n🔐 Случайный путь для приложения:\n")
	fmt.Printf("   /%s\n\n", randomPath)
	fmt.Println("Добавьте в .env файл:")
	fmt.Printf("   APP_PATH=/%s\n\n", randomPath)
	fmt.Println("После этого приложение будет доступно по адресу:")
	fmt.Printf("   http://localhost:8080/%s\n\n", randomPath)
	fmt.Println("⚠️  Не забудьте перезапустить сервер!")
}

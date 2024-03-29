package main

import (
	"flag"

	"database/sql"
	"log"
	"os"
	cnv "strconv"
	"time"

	tgbotapi "github.com/Syfaro/telegram-bot-api"

	_ "github.com/mattn/go-sqlite3"
)

var (
	// глобальная переменная в которой храним токен
	telegramBotToken string
)

func init() {
	// принимаем на входе флаг -telegrambottoken
	flag.StringVar(&telegramBotToken, "telegrambottoken", "Your_API_Token", "Telegram Bot Token")
	flag.Parse()

	// без него не запускаемся
	if telegramBotToken == "" {
		log.Print("-telegrambottoken is required")
		os.Exit(1)
	}
}

type item struct {
	id    int
	item1 string
	item2 string
	item3 string
}

var now = time.Now().Format("02.01")

func createBase() {
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	result, err := db.Exec("CREATE TABLE database (" +
		"id        INTEGER PRIMARY KEY AUTOINCREMENT," +
		"item1     TEXT, " +
		"item2     TEXT, " +
		"item3     TEXT );")
	if err != nil {
		panic(err)
	}
	log.Println("База данных создана")
	log.Println(result.RowsAffected()) // количество добавленных строк

}
func getAllItems() []item {

	var items = []item{}
	dbase, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	defer dbase.Close()
	rows, err := dbase.Query("SELECT * FROM database")
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		db := item{}
		err := rows.Scan(&db.id, &db.item1, &db.item2, &db.item3)
		if err != nil {
			log.Printf("%d", err)
			continue
		}
		items = append(items, db)
	}

	for _, db := range items {
		log.Printf("%d, %s, %s, %s", db.id, db.item1, db.item2, db.item3)
	}
	return items
}

func insertItem() {
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	res, err := db.Exec("INSERT INTO database (item1, item2, item3) VALUES (?,?,?)",
		"one", "two", "three")
	if err != nil {
		panic(err)
	}
	lastId, err := res.LastInsertId()
	log.Printf("Элемент %d добавлен.", lastId)
	if err != nil {
		panic(err)
	}
}

func deleteItem(n int) {
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	res, err := db.Exec("DELETE FROM database WHERE id = ?", n)
	if err != nil {
		panic(err)
	}
	rows, err := res.RowsAffected()
	log.Printf("Элемент %d удалён.", rows)
	if err != nil {
		panic(err)
	}
}

func getStringFromItems(items []item) string {

	str := "Сегодня " + now + "\n["
	sep := "\n"
	for i := range items {
		if i == len(items)-1 || len(items) == 0 {
			sep = "]"
		}
		str = str + items[i].item1 + ", " + items[i].item2 + ", " + items[i].item3 + sep
	}
	return str
}

func getItemById(n int) []item {
	items := []item{}
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	res, err := db.Query("SELECT * FROM database WHERE id = ?", n)
	if err != nil {
		panic(err)
	}

	for res.Next() {
		db := item{}
		err := res.Scan(&db.id, &db.item1, &db.item2, &db.item3)
		if err != nil {
			log.Printf("%d", err)
			continue
		}
	}

	return items
}
func main() {
	var action, reply string
	var confirm bool
	// используя токен создаем новый инстанс бота
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// u - структура с конфигом для получения апдейтов
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// используя конфиг u создаем канал в который будут прилетать новые сообщения
	updates, err := bot.GetUpdatesChan(u)

	// в канал updates прилетают структуры типа Update
	// вычитываем их и обрабатываем
	if err != nil {
		panic(err)
	}
	for update := range updates {
		if update.Message.Text == "да" || update.Message.Text == "Да" {
			confirm = true
		} else if update.Message.Text == "нет" || update.Message.Text == "Нет" {
			confirm = false
		}
		switch action {
		case "create":
		case "insert":
		case "delete":
			{
				n, err := cnv.Atoi(update.Message.Text)
				if err != nil {
					panic(err)
				} else {
					reply = "Вы действительно хотите удалить " + getStringFromItems(getItemById(n)) + "?"
					if confirm {
						deleteItem(n)
						reply = "Данные удалены из строки " + cnv.Itoa(n)
					}
				}
			}
		}
		// универсальный ответ на любое сообщение
		reply = "Не знаю что сказать"

		if update.Message == nil {
			continue
		}

		// логируем от кого какое сообщение пришло
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// свитч на обработку комманд
		// комманда - сообщение, начинающееся с "/"
		switch update.Message.Command() {
		case "start":
			reply = "Привет. Я телеграм-бот"
		case "hello":
			reply = "world"
		case "status":
			reply = "Бот работает"
		case "items":
			reply = getStringFromItems(getAllItems())
		case "create":
			createBase()
			reply = "База данных создана"
		case "insert":
			action = "insert"
			insertItem()
			reply = "Данные добавлены"
		case "delete":
			action = "delete"
			reply = "Введите номер строки"
		}

		// создаем ответное сообщение
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		// отправляем
		bot.Send(msg)

	}

}

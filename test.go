package main

import (
    "fmt"
     "sync"
    "net/http"
//     "encoding/json"
    "strings"
    "strconv"
    "os"
    "time"
)

//
type data_users struct {
     lock sync.RWMutex
     list map[string]int64
}

var (
   users data_users
   time_out int64 = 30 // таймаут минуты
)

func init() {
    users.list = make(map[string]int64)
//     Запуск таймера
    go timer()
}

//  Таймер для чистки списка пользователей бум проверять раз в минуту
func timer() {
    for {
        time.Sleep(time.Minute)
//         logger(time.Now().Format(time.RFC3339)," Check users ")
        users.check_time(time_out)
    }
}

func main() {

    PORT := ":8001"
    arguments := os.Args

    if len(arguments) == 1 {
        logger("Using default port number: ", PORT)
    } else {
        PORT = ":" + arguments[1]
    }

    http.HandleFunc("/", route)
    err := http.ListenAndServe(PORT, nil)
    if err != nil {
        logger(err)
        return
    }
}

// request handler
func route(w http.ResponseWriter, r *http.Request) {

    //  Вывод в журнал данных запроса
    now := time.Now()
    logger(now.Format(time.RFC3339)," ",strings.Split(r.RemoteAddr, ":")[0]," ",r.Method," ", r.URL.Path)

//     Обработка запроса
    switch r.Method {
	case "GET":
		route_get(w,r)
	default:
        w.WriteHeader(http.StatusNotFound)
// 		fmt.Fprintf(w, "Ok")
	}

}

// Обработка get запросов
func route_get(w http.ResponseWriter, r *http.Request) {

    switch r.URL.Path {
	case "/user/ping":
		users.write(w,r)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "{}")
    case "/admin/users":
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, users.read())
	default:
        w.WriteHeader(http.StatusNotFound)
	}
}

// Проверка на таймаут
func (users *data_users) check_time(timeout int64 ) {
    current_time := time.Now().Unix()
    users.lock.Lock()
        for key, value := range users.list {
            if  current_time - value > timeout*60 {
                delete(users.list,key)
//                 logger(time.Now().Format(time.RFC3339)," Delete ",key)
            }
        }
     users.lock.Unlock()
}

// Записать данные
func (users *data_users) write(w http.ResponseWriter, r *http.Request) {
    users.lock.Lock()
    users.list[ strings.Split(r.RemoteAddr, ":")[0] ] =  time.Now().Unix()
    users.lock.Unlock()
}

// Получить список пользователей
func (users *data_users) read()(string) {

    js :="["
    users.lock.RLock()
        for key, value := range users.list { // так проще, хотя можно и json.Marshal
            js += `{ "ip_address": "`+key+`", "since": `+strconv.FormatInt(value,10)+` },`
        }
    users.lock.RUnlock()
    return strings.TrimRight(js, ",")+"]"
}

// Вывод в журнал
func logger( message ...interface{} ){
    fmt.Println(message...)
}

//     resp := make(map[string]string)
// 	resp["message"] = "Status OK"
// 	jsonResp, err := json.Marshal(resp)
// 	if err != nil {
// 		fmt.Println("Error happened in JSON marshal. Err: %s", err)
// 	}
//  	w.Write([]byte("{}"))
